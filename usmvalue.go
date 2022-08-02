package parser

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type USMValueInfo struct {
	Name string
	Size int
}

var values = map[byte]USMValueInfo{
	0x10: {"Char", 1},
	0x11: {"Unsigned Char", 1},
	0x12: {"Short", 2},
	0x13: {"Unsigned Short", 2},
	0x14: {"Integer", 4},
	0x15: {"Unsigned Integer", 4},
	0x16: {"Long long", 8},
	0x17: {"Unsigned long long", 8},
	0x18: {"Float", 4},
	0x19: {"Double", 8},
	// Pointer to the start of string
	0x1A: {"String", 4},
	// Bytes start and end pointers
	0x1B: {"Bytes", 4},
}

// GetValue return what USMValue type byte represents
// formula looks like:
// valueType + (valueOccurrence << 5)
//
// valueOccurrence:
// - 1: occurring (1 << 5 = 0x20)
// - 2: unique (2 << 5 = 0x40)
func GetValue(c byte) (USMValueInfo, bool) {
	var unique bool
	if c >= 0x50 {
		// unique
		c -= 0x40
		unique = true
	} else if c >= 0x20 {
		// occurring
		c -= 0x20
	}

	return values[c], unique
}

type Entry struct {
	Type  USMValueInfo
	Value []byte
}

func (e Entry) String() string {
	if e.Type.Name == "Bytes" {
		// return bytes
		return fmt.Sprintf("% x", e.Value)
	}

	if e.Type.Name == "String" {
		// encode string
		return fmt.Sprintf("%s", string(e.Value))
	}

	//return fmt.Sprintf("%d", e.Value)

	var signed = true
	if strings.HasPrefix(e.Type.Name, "Unsigned") {
		signed = false
	}

	switch e.Type.Size {
	case 1:
		if signed {
			return fmt.Sprintf("%d", int8(e.Value[0]))
		} else {
			return fmt.Sprintf("%d", e.Value)
		}
	case 2:
		val := binary.BigEndian.Uint16(e.Value)
		if signed {
			return fmt.Sprintf("%d", int16(val))
		} else {
			return fmt.Sprintf("%d", val)
		}
	case 4:
		val := binary.BigEndian.Uint32(e.Value)
		if e.Type.Name == "Float" {
			return fmt.Sprintf("%f", float32(val))
		}
		if signed {
			return fmt.Sprintf("%d", int32(val))
		} else {
			return fmt.Sprintf("%d", val)
		}
	case 8:
		val := binary.BigEndian.Uint64(e.Value)
		if e.Type.Name == "Double" {
			return fmt.Sprintf("%f", float64(val))
		}
		if signed {
			return fmt.Sprintf("%d", int64(val))
		} else {
			return fmt.Sprintf("%d", val)
		}
	}

	return fmt.Sprintf("%d", e.Value)
}
