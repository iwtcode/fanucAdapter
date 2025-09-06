using Confluent.Kafka;
using FanucFocas.Domain.Dtos;
using FanucFocas.Domain.Interfaces;
using Microsoft.AspNetCore.Mvc;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading.Tasks;
using static FanucFocas.Domain.Dtos.UnifiedMachineData;
using static Focas1;


namespace FanucFocas.Adapters
{
    public abstract class FanucAdapter : IMachineAdapter
    {
        protected readonly ILogger _logger;
        public string _ip { get; set; }

        protected FanucAdapter(ILogger logger)
        {
            _logger = logger;
        }

        // Абстрактные свойства для параметров
        protected abstract short PartsCountParameter { get; }
        protected abstract short OperatingTimeParameter { get; }
        protected abstract short PowerOnTimeParameter { get; }
        protected abstract short CycleTimeParameter { get; }
        protected abstract short CuttingTimeParameter { get; }

        public UnifiedMachineData ConvertToUnifiedData(ushort handle, string ip)
        {
            var data = new UnifiedMachineData
            {
                Timestamp = DateTime.UtcNow,
                AxisInfos = new List<AxisInfo>(),
                SpindleInfos = new List<SpindleInfo>(),
                Alarms = new List<AlarmDetail>(),
                CurrentProgram = new ProgramStatus(),
                IsEnabled = true
            };

            _ip = ip;

            try
            {
                ReadActiveToolNumber(handle, data);

                ReadMachineState(handle, data);
                ReadCurrentGCode(handle, data);
                ReadAxisData(handle, data);
                ReadSpindleData(handle, data);
                ReadFeedData(handle, data);
                ReadErrors(handle, data);


                ReadContourFeedRate(handle, data);

                ReadFeedOverride(handle, data);
                ReadJogOverride(handle, data);



                // Чтение данных из параметров станка
                ReadPartsCount(handle, data);
                ReadOperatingTime(handle, data);
                ReadPowerOnTime(handle, data);
                ReadCycleTime(handle, data);
                ReadCuttingTime(handle, data);

                //ReadAndPrintProcessingTimeStamps(handle);
                //ReadCncData(handle);

                //ReadToolControlData(handle);

                //ReadToolInfo(handle);

                //ReadToolData(handle, 1);

                //ReadAllTools(handle);

            }
            catch (Exception ex)
            {
                _logger.Error($"[{_ip}] [FanucAdapter] Error while polling data: {ex.Message}");
                data.IsEnabled = false;
            }

            return data;
        }

        protected abstract void ReadMachineState(ushort handle, UnifiedMachineData data);

        public abstract ControlProgram GetControlProgram(ushort handle, string ip);

        protected void ReadAxisData(ushort handle, UnifiedMachineData data)
        {
            short numAxes = 8;
            var pos = new ODBPOS();
            short ret = Focas1.cnc_rdposition(handle, 0, ref numAxes, pos);

            if (ret != EW_OK || numAxes <= 0)
            {
                _logger.Warn($"[{_ip}] [FanucAdapter] cnc_rdposition Error code:{ret}");
                return;
            }

            // Считываем позиции осей без отражения
            var axes = new[] { pos.p1, pos.p2, pos.p3, pos.p4, pos.p5, pos.p6, pos.p7, pos.p8 };

            // Считываем нагрузку на оси
            var svLoad = new ODBSVLOAD();
            ret = Focas1.cnc_rdsvmeter(handle, ref numAxes, svLoad);

            for (int i = 0; i < numAxes && i < axes.Length; i++)
            {
                var absPos = axes[i];
                string name = absPos.abs.name.ToString().TrimEnd('\0');
                string suffix = absPos.abs.suff.ToString().TrimEnd('\0');

                string axisName = !string.IsNullOrEmpty(suffix) ? name + suffix : name;
                double position = absPos.abs.data / Math.Pow(10, absPos.abs.dec);
                short load = GetAxisLoad(svLoad, i);

                // Создаем объект AxisInfo
                var axisInfo = new AxisInfo
                {
                    Name = axisName,
                    Position = position,
                    LoadPercent = load
                };

                // Чтение данных диагностики для оси
                try
                {
                    short axisIndex = (short)(i + 1);
                    axisInfo.ServoTemperature = ReadDiagnosisByte(handle, 308, axisIndex);
                    axisInfo.CoderTemperature = ReadDiagnosisByte(handle, 309, axisIndex);
                    axisInfo.PowerConsumption = ReadDiagnosisDoubleWord(handle, 4901, axisIndex);
                }
                catch (Exception ex)
                {
                    _logger.Warn($"[{_ip}] [FanucAdapter] [ReadAxisData] cnc_diagnoss, error message: {ex.Message}");
                }

                // Добавляем информацию об оси в список
                data.AxisInfos.Add(axisInfo);
            }
        }

        private short GetAxisLoad(ODBSVLOAD svLoad, int index)
        {
            switch (index)
            {
                case 0: return (short)svLoad.svload1.data;
                case 1: return (short)svLoad.svload2.data;
                case 2: return (short)svLoad.svload3.data;
                case 3: return (short)svLoad.svload4.data;
                case 4: return (short)svLoad.svload5.data;
                case 5: return (short)svLoad.svload6.data;
                case 6: return (short)svLoad.svload7.data;
                case 7: return (short)svLoad.svload8.data;
                default: return 0;
            }
        }

        private void ReadSpindleData(ushort handle, UnifiedMachineData data)
        {
            short spindleCount = 4;
            var buffer = new ODBSPLOAD();

            // Чтение основных данных шпинделя
            short ret = Focas1.cnc_rdspmeter(handle, -1, ref spindleCount, buffer);

            if (ret != Focas1.EW_OK || spindleCount <= 0)
            {
                _logger.Warn($"[{_ip}] [FanucAdapter] cnc_rdspmeter, error code: {ret}");
                return;
            }

            var list = new[] { buffer.spload1, buffer.spload2, buffer.spload3, buffer.spload4 };

            // Чтение данных корректоров шпинделя
            var spindleOverrideData = new ODBSPN();
            ret = Focas1.cnc_rdspload(handle, -1, spindleOverrideData);

            if (ret != Focas1.EW_OK)
            {
                _logger.Warn($"[{_ip}] [FanucAdapter] cnc_rdspload, error code: {ret}");
                return;
            }

            for (short i = 0; i < spindleCount && i < list.Length; i++)
            {
                int load = list[i].spload.data;
                int speed = list[i].spspeed.data;

                // Создаем объект SpindleInfo
                var spindleInfo = new SpindleInfo
                {
                    Number = i + 1,
                    SpeedRPM = speed,
                    LoadPercent = load,
                    OverridePercent = spindleOverrideData.data[i]
                };

                // Чтение данных диагностики для шпинделя
                try
                {
                    var spindleNumber = (short)(i + (short)1);
                    //spindleInfo.DiagnosticLoadMin = ReadDiagnosisWord(handle, 411, spindleNumber);
                    //spindleInfo.DiagnosticCoderFeedback = ReadDiagnosisDoubleWord(handle, 417, spindleNumber);
                    //spindleInfo.DiagnosticLoopDeviation = ReadDiagnosisDoubleWord(handle, 418, spindleNumber);
                    //spindleInfo.DiagnosticSyncError = ReadDiagnosisDoubleWord(handle, 425, spindleNumber);
                    //spindleInfo.DiagnosticRevCount1 = ReadDiagnosisDoubleWord(handle, 1520, spindleNumber);
                    //spindleInfo.DiagnosticRevCount2 = ReadDiagnosisDoubleWord(handle, 1521, spindleNumber);
                    spindleInfo.PowerConsumption = ReadDiagnosisDoubleWord(handle, 4902, spindleNumber);
                }
                catch (Exception ex)
                {
                    _logger.Warn($"[{_ip}] [FanucAdapter] [ReadSpindleData] cnc_diagnoss, message: {ex.Message}");
                }

                // Добавляем информацию о шпинделе в список
                data.SpindleInfos.Add(spindleInfo);
            }
        }

        private void ReadFeedData(ushort handle, UnifiedMachineData data)
        {
            // Считываем фактическую подачу (feed rate)
            ODBSPEED speed = new ODBSPEED();
            if (cnc_rdspeed(handle, 2, speed) == EW_OK)
            {
                data.FeedRate = speed.actf.data;
            }

            // Считываем коэффициент подачи (override)
            IODBPSD_1 param = new IODBPSD_1();
            if (cnc_rdparam(handle, 20, 0, 8, param) == EW_OK)
            {
                data.FeedOverride = param.cdata;
            }
        }

        private void ReadContourFeedRate(ushort handle, UnifiedMachineData data)
        {
            ODBACT feedData = new ODBACT();
            short ret = Focas1.cnc_actf(handle, feedData);
            if (ret != Focas1.EW_OK)
            {
                _logger.Warn($"[{_ip}] [FanucAdapter] cnc_actf, error code: {ret}");
                return;
            }

            data.ContourFeedRate = feedData.data; // Текущая подача в мм/мин или дюйм/мин
        }

        private void ReadFeedOverride(ushort handle, UnifiedMachineData data)
        {
            ODBTOFS overrideData = new ODBTOFS();
            short ret = Focas1.cnc_rdtofs(handle, 1, 0, 8, overrideData); // Чтение корректора подачи (F%)
            if (ret != Focas1.EW_OK)
            {
                _logger.Warn($"[{_ip}] [FanucAdapter] overrideData cnc_rdtofs, error code: {ret}");
                return;
            }

            data.FeedOverride = overrideData.data; // Процент корректора подачи
        }

        private void ReadJogOverride(ushort handle, UnifiedMachineData data)
        {
            ODBTOFS jogOverrideData = new ODBTOFS();
            short ret = Focas1.cnc_rdtofs(handle, 1, 1, 8, jogOverrideData); // Чтение корректора JOG%
            if (ret != Focas1.EW_OK)
            {
                _logger.Warn($"[{_ip}] [FanucAdapter] jogOverrideData cnc_rdtofs, error code: {ret}");
                return;
            }

            data.JogOverride = jogOverrideData.data; // Процент корректора JOG
        }

        private void ReadErrors(ushort handle, UnifiedMachineData data)
        {
            short count = 10;
            ODBALMMSG alarmMsg = new ODBALMMSG();

            if (Focas1.cnc_rdalmmsg(handle, -1, ref count, alarmMsg) == Focas1.EW_OK && count > 0)
            {
                data.HasAlarms = true;

                ODBALMMSG_data[] messages = new ODBALMMSG_data[]
                {
                    alarmMsg.msg1,
                    alarmMsg.msg2,
                    alarmMsg.msg3,
                    alarmMsg.msg4,
                    alarmMsg.msg5,
                    alarmMsg.msg6,
                    alarmMsg.msg7,
                    alarmMsg.msg8,
                    alarmMsg.msg9,
                    alarmMsg.msg10
                };

                for (int i = 0; i < count && i < messages.Length; i++)
                {
                    var alarm = messages[i];
                    string message = alarm.alm_msg != null
                        ? alarm.alm_msg.TrimEnd('\0', ' ')
                        : string.Empty;

                    if (!string.IsNullOrWhiteSpace(message))
                    {
                        string typeDescription = "Unknown";

                        switch (alarm.type)
                        {
                            case 0: typeDescription = "SW – Parameter switch on"; break;
                            case 1: typeDescription = "PW – Power off parameter set"; break;
                            case 2: typeDescription = "IO – I/O error"; break;
                            case 3: typeDescription = "PS – Foreground P/S"; break;
                            case 4: typeDescription = "OT – Overtravel / External data"; break;
                            case 5: typeDescription = "OH – Overheat alarm"; break;
                            case 6: typeDescription = "SV – Servo alarm"; break;
                            case 7: typeDescription = "SR – Data I/O error"; break;
                            case 8: typeDescription = "MC – Macro alarm"; break;
                            case 9: typeDescription = "SP – Spindle alarm"; break;
                            case 10: typeDescription = "DS – Other alarm"; break;
                            case 11: typeDescription = "IE – Malfunction prevention function"; break;
                            case 12: typeDescription = "BG – Background P/S"; break;
                            case 13: typeDescription = "SN – Synchronization error"; break;
                            case 14: typeDescription = "(Reserved)"; break;
                            case 15: typeDescription = "EX – External alarm message"; break;
                        }

                        var alarmDetail = new UnifiedMachineData.AlarmDetail
                        {
                            ErrorCode = alarm.alm_no.ToString(),
                            ErrorTypeDescription = typeDescription,
                            ErrorMessage = message
                        };

                        data.Alarms.Add(alarmDetail);
                    }
                }
            }
            else
            {
                data.HasAlarms = false;
            }
        }

        private void UploadNCProgram(ushort handle, string ncProgram)
        {
            short progNum = 1;

            short retStart = Focas1.cnc_upstart(handle, progNum);
            if (retStart != Focas1.EW_OK)
            {
                Console.WriteLine($"Ошибка запуска загрузки (cnc_upstart): {retStart}");
                return;
            }

            byte[] allBytes = Encoding.ASCII.GetBytes(ncProgram);
            int totalSent = 0;

            while (totalSent < allBytes.Length)
            {
                int toSend = Math.Min(256, allBytes.Length - totalSent);

                ODBUP packet = new ODBUP
                {
                    dummy = new short[2],
                    data = new char[256]
                };

                for (int i = 0; i < toSend; i++)
                    packet.data[i] = (char)allBytes[totalSent + i];

                ushort actualLength = (ushort)toSend;
                short ret = Focas1.cnc_upload(handle, packet, ref actualLength);

                Console.WriteLine($"[UPLOAD] cnc_upload: ret={ret}, sent={actualLength}");

                if (ret != Focas1.EW_OK)
                {
                    Console.WriteLine("!! Ошибка передачи блока.");
                    break;
                }

                totalSent += toSend;
                if (toSend < 256) break;
            }

            short endRet = Focas1.cnc_dncend(handle);
            Console.WriteLine($"[UPLOAD] cnc_dncend: ret={endRet}");
        }
        private void ReadCurrentGCode(ushort handle, UnifiedMachineData data)
        {
            ushort length = 32;
            short blockNumber = 0;
            byte[] buffer = new byte[length];

            short ret = Focas1.cnc_rdexecprog(handle, ref length, out blockNumber, buffer);

            if (ret != Focas1.EW_OK)
            {
                _logger.Warn($"[{_ip}] [FanucAdapter] readCurrentGCode cnc_rdexecprog, error code: {ret}");
                return;
            }

            string gCodeLine = Encoding.ASCII.GetString(buffer, 0, length).Trim();

            data.CurrentProgram.GCodeLine = gCodeLine.Split('\n')[0];

            var exePrg = new ODBEXEPRG();
            ret = Focas1.cnc_exeprgname(handle, exePrg);

            if (ret != Focas1.EW_OK)
            {
                throw new Exception($"cnc_exeprgname, error code: {ret}");
            }

            // Извлечение имени и номера программы из структуры
            string programName = new string(exePrg.name).Trim('\0').Trim(); // Убираем нулевые символы и пробелы
            short programNumber = (short)exePrg.o_num;

            // Заполняем соответствующие поля результата
            data.CurrentProgram.ProgramName = programName;
            data.CurrentProgram.ProgramNumber = programNumber;

        }


        // Не работает
        public void ReadAndPrintProcessingTimeStamps(ushort handle)
        {
            var ptime = new ODBPTIME();
            short ret = cnc_rdproctime(handle, ptime);

            if (ret != 0) // EW_OK == 0
            {
                Console.WriteLine($"Ошибка при чтении данных о времени выполнения. Код: {ret}");
                return;
            }

            if (ptime.num <= 0)
            {
                Console.WriteLine("Нет данных о времени выполнения.");
                return;
            }

            Console.WriteLine($"Количество записей: {ptime.num}");

            // Массив данных для удобного доступа
            ODBPTIME_data[] records = new ODBPTIME_data[10]
            {
                ptime.data.data1, ptime.data.data2, ptime.data.data3, ptime.data.data4, ptime.data.data5,
                ptime.data.data6, ptime.data.data7, ptime.data.data8, ptime.data.data9, ptime.data.data10
            };

            for (int i = 0; i < ptime.num; i++)
            {
                var record = records[i];
                Console.WriteLine($"Запись {i + 1}:");
                Console.WriteLine($"  Номер программы: {record.prg_no}");
                Console.WriteLine($"  Время выполнения: {record.hour} часов, {record.minute} минут, {record.second} секунд");
            }
        }


        // Чтение данных из диагностических параметров
        protected int ReadDiagnosisByte(ushort handle, short number, short axis = 0)
        {
            return ReadDiagnosis(handle, number, axis, 4 + 1).Data; // 1 байт
        }

        protected int ReadDiagnosisWord(ushort handle, short number, short axis = 0)
        {
            return ReadDiagnosis(handle, number, axis, 4 + 2).Data; // 2 байта
        }

        protected long ReadDiagnosisDoubleWord(ushort handle, short number, short axis = 0)
        {
            return ReadDiagnosis(handle, number, axis, 4 + 4).Data; // 4 байта
        }

        private (bool Success, int Data) ReadDiagnosis(ushort handle, short number, short axis, short length)
        {
            var diag = new ODBDGN_1();
            short ret = Focas1.cnc_diagnoss(handle, number, axis, length, diag);

            if (ret != Focas1.EW_OK)
            {
                throw new Exception($"cnc_diagnoss, error code: {ret}");
            }

            // Определяем тип данных на основе поля type
            switch (diag.type & 0xFF00) // Выделяем старший байт type для определения типа данных
            {
                case 0x0000: // Byte type
                    return (true, diag.cdata);
                case 0x0100: // Word type
                    return (true, diag.idata);
                case 0x0200: // Double word type
                    return (true, diag.ldata);
                default:
                    throw new Exception($"Unknown data type for cnc_diagnoss {number}, ось {axis}. Type: {diag.type}");
            }
        }



        // Чтение данных из параметров
        protected int ReadParameter(ushort handle, short parameter)
        {
            IODBPSD_1 paramData = new IODBPSD_1();
            short ret = Focas1.cnc_rdparam(handle, parameter, 0, 8, paramData);
            if (ret != Focas1.EW_OK)
            {
                throw new Exception($"cnc_rdparam({parameter}), error code: {ret}");
            }
            return paramData.cdata;
        }

        protected void ReadTimeParameter(ushort handle, UnifiedMachineData data, short parameter, Action<TimeSpan> setTimeAction)
        {
            try
            {
                int value = ReadParameter(handle, parameter);
                setTimeAction(TimeSpan.FromMinutes(value));
            }
            catch (Exception ex)
            {
                _logger.Warn($"[{_ip}] [FanucAdapter] readTimeParameter {ex.Message}");
            }
        }



        // Реализация методов с использованием абстрактных свойств
        protected void ReadPartsCount(ushort handle, UnifiedMachineData data)
        {
            try
            {
                int value = ReadParameter(handle, PartsCountParameter);
                data.PartsCount = value;
            }
            catch (Exception ex)
            {
                _logger.Warn($"[{_ip}] [FanucAdapter] readPartsCount {ex.Message}");
            }
        }

        protected void ReadOperatingTime(ushort handle, UnifiedMachineData data)
        {
            ReadTimeParameter(handle, data, OperatingTimeParameter, time => data.OperatingTime = time);
        }

        protected void ReadPowerOnTime(ushort handle, UnifiedMachineData data)
        {
            ReadTimeParameter(handle, data, PowerOnTimeParameter, time => data.PowerOnTime = time);
        }

        protected void ReadCycleTime(ushort handle, UnifiedMachineData data)
        {
            ReadTimeParameter(handle, data, CycleTimeParameter, time => data.CycleTime = time);
        }

        protected void ReadCuttingTime(ushort handle, UnifiedMachineData data)
        {
            ReadTimeParameter(handle, data, CuttingTimeParameter, time => data.CuttingTime = time);
        }


        private void ReadActiveToolNumber(ushort handle, UnifiedMachineData data)
        {
            IODBTLCTL toolControlData = new IODBTLCTL();
            short ret = Focas1.cnc_rdtlctldata(handle, toolControlData);

            if (ret != Focas1.EW_OK)
            {
                _logger.Warn($"[{_ip}] [FanucAdapter] cnc_rdtlctldata, error code: {ret}");
                return;
            }

            // Сохраняем данные в объект UnifiedMachineData
            short activeToolNumber = toolControlData.used_tool;
            data.ActiveToolNumber = activeToolNumber;
        }

        public void ReadAllTools(ushort handle)
        {
            short maxTools = 5; // Максимальное количество групп инструментов (например, 5)
            short toolCount = 2;

            IODBTLMNG toolManagementData = new IODBTLMNG();
            short ret = Focas1.cnc_rdtool(handle, maxTools, ref toolCount, toolManagementData);

            if (ret == Focas1.EW_OK)
            {
                Console.WriteLine("=== All Tools Data ===");
                Console.WriteLine($"Total Tool Groups: {toolCount}");

                // Перебор данных о каждой группе инструментов
                var toolGroups = new[] { toolManagementData.data1, toolManagementData.data2, toolManagementData.data3, toolManagementData.data4, toolManagementData.data5 };
                for (int i = 0; i < toolCount; i++)
                {
                    var toolGroup = toolGroups[i];
                    Console.WriteLine($"Tool Group {i + 1}:");
                    Console.WriteLine($"  T_code: {toolGroup.T_code}");
                    Console.WriteLine($"  Life Count: {toolGroup.life_count}");
                    Console.WriteLine($"  Max Life: {toolGroup.max_life}");
                    Console.WriteLine($"  Rest Life: {toolGroup.rest_life}");
                    Console.WriteLine($"  Spindle Speed: {toolGroup.spindle_speed}");
                    Console.WriteLine($"  Feedrate: {toolGroup.feedrate}");
                    Console.WriteLine($"  Magazine: {toolGroup.magazine}");
                    Console.WriteLine($"  Pot: {toolGroup.pot}");
                    Console.WriteLine($"  H_code: {toolGroup.H_code}");
                    Console.WriteLine($"  D_code: {toolGroup.D_code}");
                }
            }
            else
            {
                Console.WriteLine($"Error reading all tools data. Error code: {ret}");
            }
        }

        public void ReadToolData(ushort handle, short toolNumber)
        {
            // Инициализация структуры для хранения данных инструмента
            IODBTLDT toolData = new IODBTLDT();
            short number = 1; // Количество инструментов для чтения (здесь 1)

            // Вызов функции cnc_rdtooldata
            short ret = Focas1.cnc_rdtooldata(handle, toolNumber, ref number, toolData);

            if (ret == Focas1.EW_OK)
            {
                Console.WriteLine("=== Tool Data ===");

                // Проверка флага slct для вывода только валидных данных
                if ((toolData.data1.slct & 0x0001) != 0) // Бит 0: Tool number
                    Console.WriteLine($"Tool Number: {toolData.data1.tool_no}");

                if ((toolData.data1.slct & 0x0002) != 0) // Бит 1: X-axis offset
                    Console.WriteLine($"X-axis Offset: {toolData.data1.x_axis_ofs}");

                if ((toolData.data1.slct & 0x0004) != 0) // Бит 2: Y-axis offset
                    Console.WriteLine($"Y-axis Offset: {toolData.data1.y_axis_ofs}");

                if ((toolData.data1.slct & 0x0008) != 0) // Бит 3: Turret position
                    Console.WriteLine($"Turret Position: {toolData.data1.turret_pos}");

                if ((toolData.data1.slct & 0x0010) != 0) // Бит 4: Tool number to be changed
                    Console.WriteLine($"Tool Number to be Changed: {toolData.data1.chg_tl_no}");

                if ((toolData.data1.slct & 0x0020) != 0) // Бит 5: Number of punch operation
                    Console.WriteLine($"Punch Count: {toolData.data1.punch_count}");

                if ((toolData.data1.slct & 0x0040) != 0) // Бит 6: Tool life
                    Console.WriteLine($"Tool Life: {toolData.data1.tool_life}");

                if ((toolData.data1.slct & 0x0080) != 0) // Бит 7: Radius of multiple tool
                    Console.WriteLine($"Multiple Tool Radius: {toolData.data1.m_tl_radius}");

                if ((toolData.data1.slct & 0x0100) != 0) // Бит 8: Angle of multiple tool
                    Console.WriteLine($"Multiple Tool Angle: {toolData.data1.m_tl_angle}");

                if ((toolData.data1.slct & 0x0200) != 0) // Бит 9: Tool shape
                    Console.WriteLine($"Tool Shape: {toolData.data1.tl_shape}");

                if ((toolData.data1.slct & 0x0400) != 0) // Бит 10: Tool size I
                    Console.WriteLine($"Tool Size I: {toolData.data1.tl_size_i}");

                if ((toolData.data1.slct & 0x0800) != 0) // Бит 11: Tool size J
                    Console.WriteLine($"Tool Size J: {toolData.data1.tl_size_j}");

                if ((toolData.data1.slct & 0x1000) != 0) // Бит 12: Tool angle K
                    Console.WriteLine($"Tool Angle K: {toolData.data1.tl_angle}");
            }
            else
            {
                Console.WriteLine($"Error reading tool data for tool {toolNumber}. Error code: {ret}");
            }
        }

        public void ReadToolInfo(ushort handle)
        {
            ODBPTLINF toolInfo = new ODBPTLINF();
            short ret = Focas1.cnc_rdtoolinfo(handle, toolInfo);

            if (ret == Focas1.EW_OK)
            {
                Console.WriteLine("=== Tool Information ===");
                Console.WriteLine($"Max Tools: {toolInfo.tld_max}");
                Console.WriteLine($"Max Tool Groups: {toolInfo.mlt_max}");
                Console.WriteLine($"Tool Sizes: {string.Join(", ", toolInfo.tld_size)}");
                Console.WriteLine($"Group Sizes: {string.Join(", ", toolInfo.mlt_size)}");
            }
            else
            {
                Console.WriteLine($"Error reading tool info. Error code: {ret}");
            }
        }

        public void ReadToolControlData(ushort handle)
        {
            IODBTLCTL toolControlData = new IODBTLCTL();
            short ret = Focas1.cnc_rdtlctldata(handle, toolControlData);

            if (ret == Focas1.EW_OK)
            {
                Console.WriteLine("=== Tool Control Data ===");
                Console.WriteLine($"Active Tool Number: {toolControlData.used_tool}");
                Console.WriteLine($"Turret Index: {toolControlData.turret_indx}");
                Console.WriteLine($"Zero Tool Number: {toolControlData.zero_tl_no}");
                Console.WriteLine($"Total Punches: {string.Join(", ", toolControlData.total_punch)}");
            }
            else
            {
                Console.WriteLine($"Error reading tool control data. Error code: {ret}");
            }
        }
    }
}