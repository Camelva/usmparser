package parser

import (
	"fmt"
)

var PayloadType = map[byte]string{
	0: "PayloadTypeStream",
	1: "PayloadTypeHeader",
	2: "PayloadTypeEnd",
	3: "PayloadTypeSeek",
}

const (
	PayloadTypeStream = 0
	PayloadTypeHeader = 1
	PayloadTypeEnd    = 2
	PayloadTypeSeek   = 3
)

type Payload struct {
	Header      Header
	PayloadData PayloadData
}

func (d Payload) String() string {
	return fmt.Sprintf(`{`+
		`"Header": %v, `+
		`"Data": {%v}`+
		`}`,
		d.Header, d.PayloadData)
}

type PayloadData struct {
	PayloadFixedData PayloadFixedData
	PayloadFlexData  PayloadFlexData
}

func (d PayloadData) Size() int {
	return 24 + len(d.PayloadFlexData.SharedArray) +
		len(d.PayloadFlexData.UniqueArray) +
		len(d.PayloadFlexData.StringArray) +
		len(d.PayloadFlexData.ByteArray)
}

func (d PayloadData) String() string {
	return fmt.Sprintf(`{`+
		`%v, `+
		`%v`+
		`}`,
		d.PayloadFixedData,
		d.PayloadFlexData,
	)
}

type PayloadFixedData struct {
	UniqueArrayOffset            uint32
	StringArrayOffset            uint32
	ByteArrayOffset              uint32
	PayloadNameOffset            uint32
	ItemsPerDictionary           uint16
	UniqueArraySizePerDictionary uint16
	NumberOfDictionary           uint32
}

func (d PayloadFixedData) Length() uint32 {
	return 24
}

func (d PayloadFixedData) String() string {
	return fmt.Sprintf(``+
		`"UniqueArrayOffset": %#x, `+
		`"StringArrayOffset": %#x, `+
		`"ByteArrayOffset": %#x, `+
		`"PayloadNameOffset": %#x, `+
		`"ItemsPerDictionary": %#x, `+
		`"UniqueArraySizePerDictionary": %#x, `+
		`"NumberOfDictionary": %#x`+
		"",
		d.UniqueArrayOffset, d.StringArrayOffset, d.ByteArrayOffset, d.PayloadNameOffset,
		d.ItemsPerDictionary, d.UniqueArraySizePerDictionary, d.NumberOfDictionary)
}

type PayloadFlexData struct {
	SharedArray []byte
	UniqueArray []byte
	StringArray []byte
	ByteArray   []byte
}

func (d PayloadFlexData) String() string {
	//return fmt.Sprintf(``+
	//	`"SharedArray": % x, `+
	//	`"UniqueArray": % x, `+
	//	`"StringArray": % x, `+
	//	`"ByteArray": % x`+
	//	``,
	//	d.SharedArray, d.UniqueArray, d.StringArray, d.ByteArray)
	return ""
}

type Subtitle struct {
	SubtitleHeader SubtitleHeader
	SubtitleString []byte
}

func (s Subtitle) String() string {
	return fmt.Sprintf(`{"Header": %v, "Text": %v}`,
		s.SubtitleHeader, string(s.SubtitleString))
}

// SubtitleHeader is encoded in LittleEndian
type SubtitleHeader struct {
	Language   uint32
	FrameRate  uint32
	FrameTime  uint32
	FrameEnd   uint32
	StringSize uint32
}

func (h SubtitleHeader) GetLang() string {
	switch h.Language {
	case 0:
		return Chinese
	case 1:
		return English
	case 2:
		return Thai
	case 3:
		return Vietnamese
	case 4:
		return French
	case 5:
		return German
	case 6:
		return Indonesian
	default:
		return Undefined
	}
}

func (h SubtitleHeader) String() string {
	return fmt.Sprintf(`{`+
		`"Language": %#x, `+
		`"FrameRate": %#x, `+
		`"FrameTime": %#x, `+
		`"FrameEnd": %#x, `+
		`"StringSize": %#x`+
		`}`,
		h.Language, h.FrameRate, h.FrameTime, h.FrameEnd, h.StringSize)
}
