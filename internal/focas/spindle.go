package focas

/*
#cgo CFLAGS: -I../../
#cgo LDFLAGS: -L../../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../../ -lfwlib32

#include "c_helpers.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

	"github.com/iwtcode/fanucService/internal/domain"
)

// ReadSpindleData считывает информацию о скорости активного шпинделя.
func ReadSpindleData(handle uint16) ([]domain.SpindleInfo, error) {
	// Размер структуры ODBSPEED равен 24 байтам (2 * SPEEDELM, где SPEEDELM = 12 байт)
	const odbspeedSize = 24
	buffer := make([]byte, odbspeedSize)

	// Вызываем C-функцию. type = 1 для получения скорости шпинделя (acts).
	rc := C.go_cnc_rdspeed(C.ushort(handle), 1, (*C.ODBSPEED)(unsafe.Pointer(&buffer[0])))
	if rc != C.EW_OK {
		return nil, fmt.Errorf("cnc_rdspeed failed: rc=%d", int16(rc))
	}

	// Структура ODBSPEED содержит 2 поля SPEEDELM: actf (подача) и acts (шпиндель).
	// Нас интересует второе поле - acts, которое начинается со смещения 12.
	// Схема SPEEDELM: data(4), dec(2), unit(2), disp(2), name(1), suff(1) = 12 байт.
	offset := 12

	// Читаем 4-байтное значение скорости
	speedDataVal := int32(binary.LittleEndian.Uint32(buffer[offset+0 : offset+4]))
	// Читаем 2-байтное значение десятичного разделителя
	speedDecVal := int16(binary.LittleEndian.Uint16(buffer[offset+4 : offset+6]))

	// Рассчитываем скорость с учетом разделителя
	speed := float64(speedDataVal) / math.Pow(10, float64(speedDecVal))

	// Функция cnc_rdspeed возвращает данные только для одного (активного) шпинделя.
	spindleInfos := []domain.SpindleInfo{
		{
			Number:   1,
			SpeedRPM: speed,
		},
	}

	return spindleInfos, nil
}
