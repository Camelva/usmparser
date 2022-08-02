package parser

import (
	"encoding/binary"
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
	Header
	PayloadData
}

func (d Payload) String() string {
	return fmt.Sprintf(`{`+
		`"Header": %v, `+
		`"Data": {%v}`+
		`}`,
		d.Header, d.PayloadData)
}

type PayloadData struct {
	PayloadFixedData
	PayloadFlexData
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
	UniqueArrayOffset            [4]byte
	StringArrayOffset            [4]byte
	ByteArrayOffset              [4]byte
	PayloadNameOffset            [4]byte
	ItemsPerDictionary           [2]byte
	UniqueArraySizePerDictionary [2]byte
	NumberOfDictionary           [4]byte
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

func (d PayloadFixedData) Len() uint32 {
	//return unsafe.Sizeof(d)
	// Size is always fixed
	return 24
}

func (d PayloadFixedData) GetUniqueArrayOffset() uint32 {
	return binary.BigEndian.Uint32(d.UniqueArrayOffset[:])
}

func (d PayloadFixedData) GetStringArrayOffset() uint32 {
	return binary.BigEndian.Uint32(d.StringArrayOffset[:])
}

func (d PayloadFixedData) GetByteArrayOffset() uint32 {
	return binary.BigEndian.Uint32(d.ByteArrayOffset[:])
}

func (d PayloadFixedData) GetPayloadNameOffset() uint32 {
	return binary.BigEndian.Uint32(d.PayloadNameOffset[:])
}

func (d PayloadFixedData) GetItemsPerDictionary() uint16 {
	return binary.BigEndian.Uint16(d.ItemsPerDictionary[:])
}

func (d PayloadFixedData) GetUniqueArraySizePerDictionary() uint16 {
	return binary.BigEndian.Uint16(d.UniqueArraySizePerDictionary[:])
}

func (d PayloadFixedData) GetNumberOfDictionary() uint32 {
	return binary.BigEndian.Uint32(d.NumberOfDictionary[:])
}

type PayloadFlexData struct {
	SharedArray []byte
	UniqueArray []byte
	StringArray []byte
	ByteArray   []byte
}

func (d PayloadFlexData) String() string {
	return fmt.Sprintf(``+
		`"SharedArray": % x, `+
		`"UniqueArray": % x, `+
		`"StringArray": % x, `+
		`"ByteArray": % x`+
		``,
		d.SharedArray, d.UniqueArray, d.StringArray, d.ByteArray)
}

type Subtitle struct {
	SubtitleHeader
	SubtitleString []byte
}

func (s Subtitle) String() string {
	return fmt.Sprintf(`{"Header": %v, "Text": %v}`,
		s.SubtitleHeader, string(s.SubtitleString))
}

// SubtitleHeader is encoded in LittleEndian
type SubtitleHeader struct {
	Language   [4]byte
	FrameRate  [4]byte
	FrameTime  [4]byte
	FrameEnd   [4]byte
	StringSize [4]byte
}

func (h SubtitleHeader) GetLanguage() uint32 {
	return binary.LittleEndian.Uint32(h.Language[:])
}

func (h SubtitleHeader) GetFrameRate() uint32 {
	return binary.LittleEndian.Uint32(h.FrameRate[:])
}

func (h SubtitleHeader) GetFrameTime() uint32 {
	return binary.LittleEndian.Uint32(h.FrameTime[:])
}

func (h SubtitleHeader) GetFrameEnd() uint32 {
	return binary.LittleEndian.Uint32(h.FrameEnd[:])
}

func (h SubtitleHeader) GetStringSize() uint32 {
	return binary.LittleEndian.Uint32(h.StringSize[:])
}

func (h SubtitleHeader) String() string {
	return fmt.Sprintf(`{`+
		`"Language": %v, `+
		`"FrameRate": %v, `+
		`"FrameTime": %v, `+
		`"FrameEnd": %#x, `+
		`"StringSize": %#x`+
		`}`,
		h.Language, h.FrameRate, h.FrameTime, h.FrameEnd, h.StringSize)
}
