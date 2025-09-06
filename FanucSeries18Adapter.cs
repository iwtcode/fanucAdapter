using FanucFocas.Domain.Dtos;
using FanucFocas.Domain.Interfaces;
using System;
using System.Collections.Generic;
using System.Data.Odbc;
using System.IO;
using System.Runtime.InteropServices;
using System.Text;
using static FanucFocas.Domain.Dtos.UnifiedMachineData;
using static Focas1;

namespace FanucFocas.Adapters
{
    public class FanucSeries18Adapter : FanucAdapter
    {

        public FanucSeries18Adapter(ILogger logger) : base(logger) { }

        protected override short PartsCountParameter => 6711;
        protected override short OperatingTimeParameter => 6752;
        protected override short PowerOnTimeParameter => 6750;
        protected override short CycleTimeParameter => 6758;
        protected override short CuttingTimeParameter => 6754;

        protected override void ReadMachineState(ushort handle, UnifiedMachineData data)
        {
            ODBST stat = new ODBST();
            short ret = Focas1.cnc_statinfo(handle, stat);
            if (ret != EW_OK)
            {
                _logger.Warn($"[{_ip}] [FanucSeries18Adapter] cnc_statinfo, error code: {ret}");
                return;
            }

            // 2. Режим T/M (tmmode)
            switch (stat.tmmode)
            {
                case 0: data.TmMode = "T"; break; // Токарный режим (T mode).
                case 1: data.TmMode = "M"; break; // Фрезерный режим (M mode).
                default: data.TmMode = "UNKNOWN"; break; // Неизвестный режим.
            }

            // 3. Режим работы (AUTOMATIC/MANUAL mode selection)
            switch (stat.aut)
            {
                case 0: data.ProgramMode = "MDI"; break; // Режим ручного ввода данных (MDI).
                case 1: data.ProgramMode = "MEMory"; break; // Режим выполнения программы из памяти (MEMory).
                case 2: data.ProgramMode = "No Selection"; break; // Режим не выбран.
                case 3: data.ProgramMode = "EDIT"; break; // Режим редактирования программы.
                case 4: data.ProgramMode = "HaNDle"; break; // Режим ручного управления (Handle).
                case 5: data.ProgramMode = "JOG"; break; // Режим JOG (ручное перемещение осей).
                case 6: data.ProgramMode = "Teach in JOG"; break; // Обучение в режиме JOG.
                case 7: data.ProgramMode = "Teach in HaNDle"; break; // Обучение в режиме Handle.
                case 8: data.ProgramMode = "INC·feed"; break; // Режим инкрементной подачи.
                case 9: data.ProgramMode = "REFerence"; break; // Режим возврата в исходную точку (Reference).
                case 10: data.ProgramMode = "ReMoTe"; break; // Удалённый режим (Remote).
                default: data.ProgramMode = "UNKNOWN"; break; // Неизвестный режим.
            }

            // 4. Статус выполнения программы (run)
            switch (stat.run)
            {
                case 0: data.MachineState = "Reset"; break; // Программа сброшена (Reset).
                case 1: data.MachineState = "STOP"; break; // Программа остановлена (Stop).
                case 2: data.MachineState = "HOLD"; break; // Программа приостановлена (Hold).
                case 3: data.MachineState = "START"; break; // Программа запущена (Start).
                case 4:
                    data.MachineState = "MSTR (during retraction and re-positioning of tool retraction and recovery, and operation of JOG MDI)";
                    break; // Специальный режим: ретракция и восстановление инструмента, а также операции JOG MDI.
                default: data.MachineState = "UNKNOWN"; break; // Неизвестное состояние выполнения программы.
            }

            // 5. Движение осей (motion)
            switch (stat.motion)
            {
                case 0: data.AxisMovementStatus = "None"; break; // Нет движения осей.
                case 1: data.AxisMovementStatus = "Motion"; break; // Оси находятся в движении.
                case 2: data.AxisMovementStatus = "Dwell"; break; // Оси находятся в состоянии задержки (Dwell).
                default: data.AxisMovementStatus = "UNKNOWN"; break; // Неизвестное состояние движения осей.
            }

            // 6. Статус M/S/T/B (mstb)
            data.MstbStatus = (stat.mstb == 1) ? "FIN" : "Other";
            // Если stat.mstb == 1, то статус "FIN" (завершение выполнения блока). В противном случае — "Other".

            // 7. Аварийный стоп (emergency)
            switch (stat.emergency)
            {
                case 0: data.EmergencyStatus = "Not Emergency"; break; // Не авария.
                case 1: data.EmergencyStatus = "EMerGency"; break; // Активирован аварийный стоп.
                case 2: data.EmergencyStatus = "ReSET"; break; // Система находится в состоянии сброса после аварии.
                case 3: data.EmergencyStatus = "WAIT (FS35i only)"; break; // Состояние ожидания (специфично для FS35i).
                default: data.EmergencyStatus = "UNKNOWN"; break; // Неизвестное состояние аварийного стопа.
            }

            // 8. Статус тревоги (alarm)
            switch (stat.alarm)
            {
                case 0: data.AlarmStatus = "Others"; break; // Другие состояния (нет активных тревог).
                case 1: data.AlarmStatus = "ALarM"; break; // Активна общая тревога.
                case 2: data.AlarmStatus = "BATtery Low"; break; // Низкий уровень заряда батареи.
                case 3: data.AlarmStatus = "FAN (NC or Servo amplifier)"; break; // Проблемы с вентилятором ЧПУ или сервоприводом.
                case 4: data.AlarmStatus = "PS Warning"; break; // Предупреждение о проблемах с источником питания.
                case 5: data.AlarmStatus = "FSsB Warning"; break; // Предупреждение о проблемах с FSSB (оптическая шина).
                case 6: data.AlarmStatus = "INSulate Warning"; break; // Предупреждение об изоляции.
                case 7: data.AlarmStatus = "ENCoder Warning"; break; // Предупреждение об энкодере.
                case 8: data.AlarmStatus = "PMC Alarm"; break; // Тревога от PMC (программируемого контроллера машины).
                default: data.AlarmStatus = "UNKNOWN"; break; // Неизвестное состояние тревоги.
            }

            // 9. Статус редактирования (edit)
            InterpretEditStatus(stat.tmmode, stat.edit, data);
           
        }

        private void InterpretEditStatus(short tmmode, short editValue, UnifiedMachineData data)
        {
            // Определяем режим работы станка (T mode или M mode)
            switch (tmmode)
            {
                case 0: // T mode (токарный станок)
                    switch (editValue)
                    {
                        case 0: data.EditStatus = "Not Editing"; break; // Не редактируется
                        case 1: data.EditStatus = "EDIT"; break; // Редактирование программы
                        case 2: data.EditStatus = "SEARCH"; break; // Поиск в программе
                        case 3: data.EditStatus = "OUTPUT"; break; // Вывод данных
                        case 4: data.EditStatus = "INPUT"; break; // Ввод данных
                        case 5: data.EditStatus = "COMPARE"; break; // Сравнение данных
                        case 6: data.EditStatus = "OFFSET"; break; // Режим записи компенсации длины инструмента
                        case 7: data.EditStatus = "Work Shift"; break; // Режим записи смещения заготовки
                        case 9: data.EditStatus = "Restart"; break; // Перезапуск программы
                        case 10: data.EditStatus = "RVRS"; break; // Обратное перемещение
                        case 11: data.EditStatus = "RTRY"; break; // Повторное прогрессирование
                        case 12: data.EditStatus = "RVED"; break; // Завершение обратного перемещения
                        case 14: data.EditStatus = "PTRR"; break; // Режим ретракции и восстановления инструмента
                        case 16: data.EditStatus = "AICC"; break; // Управление контуром с использованием ИИ
                        case 21: data.EditStatus = "HPCC"; break; // Выполнение операций RISC
                        case 23: data.EditStatus = "NANO HP"; break; // Высокоточное управление контуром с использованием ИИ
                        case 25: data.EditStatus = "5-AXIS"; break; // Обработка на 5-осевом станке
                        case 26: data.EditStatus = "OFSX"; break; // Изменить значение ручного активного смещения: режим изменения смещения по оси X
                        case 27: data.EditStatus = "OFSZ"; break; // Изменить значение ручного активного смещения: режим изменения смещения по оси Z
                        case 28: data.EditStatus = "WZR"; break; // Изменить значение ручного активного смещения: режим изменения смещения начала координат заготовки
                        case 29: data.EditStatus = "OFSY"; break; // Изменить значение ручного активного смещения: режим изменения смещения по оси Y
                        case 31: data.EditStatus = "TOFS"; break; // Изменить значение ручного активного смещения: режим изменения смещения инструмента
                        case 39: data.EditStatus = "TCP"; break; // Управление точкой центра инструмента при 5-осевой обработке
                        case 40: data.EditStatus = "TWP"; break; // Команда наклонной рабочей плоскости
                        case 41: data.EditStatus = "TCP+TWP"; break; // Управление точкой центра инструмента и команда наклонной рабочей плоскости
                        case 42: data.EditStatus = "APC"; break; // Расширенное предварительное управление
                        case 43: data.EditStatus = "PRG-CHK"; break; // Быстрая проверка программы
                        case 44: data.EditStatus = "APC"; break; // Расширенное предварительное управление
                        case 45: data.EditStatus = "S-TCP"; break; // Плавное управление точкой центра инструмента
                        case 59: data.EditStatus = "ALLSAVE"; break; // Сохранение программ в процессе
                        case 60: data.EditStatus = "NOTSAVE"; break; // Программы не сохранены
                        default: data.EditStatus = "UNKNOWN"; break; // Неизвестный статус
                    }
                    break;

                case 1: // M mode (фрезерный станок)
                    switch (editValue)
                    {
                        case 0: data.EditStatus = "Not Editing"; break; // Не редактируется
                        case 1: data.EditStatus = "EDIT"; break; // Редактирование программы
                        case 2: data.EditStatus = "SEARCH"; break; // Поиск в программе
                        case 3: data.EditStatus = "OUTPUT"; break; // Вывод данных
                        case 4: data.EditStatus = "INPUT"; break; // Ввод данных
                        case 5: data.EditStatus = "COMPARE"; break; // Сравнение данных
                        case 6: data.EditStatus = "Label Skip"; break; // Режим пропуска меток
                        case 7: data.EditStatus = "Restart"; break; // Перезапуск программы
                        case 8: data.EditStatus = "HPCC"; break; // Выполнение операций RISC
                        case 9: data.EditStatus = "PTRR"; break; // Режим ретракции и восстановления инструмента
                        case 10: data.EditStatus = "RVRS"; break; // Обратное перемещение
                        case 11: data.EditStatus = "RTRY"; break; // Повторное прогрессирование
                        case 12: data.EditStatus = "RVED"; break; // Завершение обратного перемещения
                        case 13: data.EditStatus = "HANDLE"; break; // Режим совмещения ручного управления
                        case 14: data.EditStatus = "OFFSET"; break; // Режим измерения длины инструмента
                        case 15: data.EditStatus = "Work Offset"; break; // Режим измерения нулевой точки заготовки
                        case 16: data.EditStatus = "AICC"; break; // Управление контуром с использованием ИИ
                        case 17: data.EditStatus = "Memory Check"; break; // Проверка памяти ленты
                        case 21: data.EditStatus = "AI APC"; break; // Расширенное предварительное управление с использованием ИИ
                        case 22: data.EditStatus = "MBL APC"; break; // Многоблочное расширенное предварительное управление
                        case 23: data.EditStatus = "NANO HP"; break; // Высокоточное управление контуром с использованием ИИ
                        case 24: data.EditStatus = "AI HPCC"; break; // Нано-высокоточное управление контуром с использованием ИИ
                        case 25: data.EditStatus = "5-AXIS"; break; // Обработка на 5-осевом станке
                        case 26: data.EditStatus = "LEN"; break; // Изменить значение ручного активного смещения: режим изменения длины смещения
                        case 27: data.EditStatus = "RAD"; break; // Изменить значение ручного активного смещения: режим изменения радиуса смещения
                        case 28: data.EditStatus = "WZR"; break; // Изменить значение ручного активного смещения: режим изменения смещения начала координат заготовки
                        case 39: data.EditStatus = "TCP"; break; // Управление точкой центра инструмента при 5-осевой обработке
                        case 40: data.EditStatus = "TWP"; break; // Команда наклонной рабочей плоскости
                        case 41: data.EditStatus = "TCP+TWP"; break; // Управление точкой центра инструмента и команда наклонной рабочей плоскости
                        case 42: data.EditStatus = "APC"; break; // Расширенное предварительное управление
                        case 43: data.EditStatus = "PRG-CHK"; break; // Быстрая проверка программы
                        case 44: data.EditStatus = "APC"; break; // Расширенное предварительное управление
                        case 45: data.EditStatus = "S-TCP"; break; // Плавное управление точкой центра инструмента
                        case 59: data.EditStatus = "ALLSAVE"; break; // Сохранение программ в процессе
                        case 60: data.EditStatus = "NOTSAVE"; break; // Программы не сохранены
                        default: data.EditStatus = "UNKNOWN"; break; // Неизвестный статус
                    }
                    break;

                default:
                    data.EditStatus = "UNKNOWN"; // Неизвестный режим
                    break;
            }
        }

        public override ControlProgram GetControlProgram(ushort handle, string ip)
        {
            try
            {
                // Инициализация результата с базовыми данными
                var result = new ControlProgram
                {
                    MachineId = ip,
                    Timestamp = DateTime.UtcNow,
                    ProgramName = null,
                    ProgramNumber = -1,
                    Program = null
                };

                // Шаг 1: Получение имени и номера выполняемой программы
                var exePrg = new ODBEXEPRG();
                short ret = Focas1.cnc_exeprgname(handle, exePrg);

                if (ret != Focas1.EW_OK)
                {
                    throw new Exception($"[{_ip}] [FanucAdapter] cnc_exeprgnamm, error code: {ret}");
                }

                // Извлечение имени и номера программы из структуры
                string programName = new string(exePrg.name).Trim('\0').Trim(); // Убираем нулевые символы и пробелы
                short programNumber = (short)exePrg.o_num;


                // Заполняем соответствующие поля результата
                result.ProgramName = programName;
                result.ProgramNumber = programNumber;


                // Шаг 2: Начало загрузки программы
                ret = Focas1.cnc_upstart4(handle, 0, programNumber);
                if (ret != Focas1.EW_OK)
                {
                    throw new Exception($"[{_ip}] [FanucAdapter] cnc_upstart({programNumber}), error code: {ret}");
                }

                // Шаг 3: Чтение тела программы построчно
                var buffer = new Focas1.ODBUP();
                const int MAX_RETRIES = 1000; // Максимальное количество попыток чтения строк
                var fullProgram = new StringBuilder(); // Буфер для хранения полного текста программы

                for (int i = 0; i < MAX_RETRIES; i++)
                {
                    ushort length = 256; // Максимальная длина строки
                    ret = Focas1.cnc_upload(handle, buffer, ref length);

                    if (ret != Focas1.EW_OK || length == 0)
                    {
                        // Если произошла ошибка или достигнут конец программы, завершаем чтение
                        break;
                    }

                    string line = new string(buffer.data).Trim('\0').Trim(); // Убираем нулевые символы и пробелы
                    if (string.IsNullOrEmpty(line))
                    {
                        // Если строка пустая, завершаем чтение
                        break;
                    }

                    fullProgram.AppendLine(line); // Добавляем строку в общий текст программы
                }

                // Шаг 4: Завершение загрузки программы
                Console.WriteLine($"{fullProgram.ToString().Trim()}");
                Focas1.cnc_upend(handle);

                result.Program = fullProgram.ToString().Trim();
                fullProgram.Clear();

                return result;
            }
            catch (Exception ex)
            {
                throw new Exception($"{ex.Message}");
            }
        }
    }
}
