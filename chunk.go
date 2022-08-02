package parser

import (
	"fmt"
)

type Chunk struct {
	offset int

	Header Header
	Data   Data
}

func (c Chunk) String() string {
	return fmt.Sprintf(`{`+
		`"Header": %v, `+
		`"Data": %v`+
		`}`,
		c.Header, c.Data)
}

type Header struct {
	ID   [4]byte
	Size int32
}

func (h Header) String() string {
	return fmt.Sprintf(`{"ID": %s, "Size": %#x}`, h.ID, h.Size)
}

func (h Header) IDString() string {
	return string(h.ID[:])
}

type Data struct {
	PayloadHeader PayloadHeader
	Payload       []byte
	//Padding []byte
}

func (d Data) String() string {
	return fmt.Sprintf(`{`+
		`"Header": %v, `+
		`"Payload": % x `+
		`}`,
		d.PayloadHeader, d.Payload,
	)
}

type PayloadHeader struct {
	// skip 1 byte
	_             byte
	Offset        byte
	PaddingSize   uint16
	ChannelNumber byte
	// skip 2 bytes
	_           [2]byte
	PayloadType byte
	FrameTime   int32
	FrameRate   int32
	// skip 8 bytes
	_ [8]byte
}

func (h PayloadHeader) Len() int {
	return 24
}

func (h PayloadHeader) String() string {
	return fmt.Sprintf(`{`+
		`"Offset": %#x, `+
		`"PaddingSize": %#x, `+
		`"ChannelNumber": %#x, `+
		`"PayloadType": %#x, `+
		`"FrameTime": %#x, `+
		`"FrameRate": %#x`+
		`}`,
		h.Offset, h.PaddingSize, h.ChannelNumber, h.PayloadType, h.FrameTime, h.FrameRate)
}

func (h PayloadHeader) GetPayloadType() int {
	return int(h.PayloadType)
}
