package parser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

func WriteChunk(src Chunk, out io.Writer) (n int64, err error) {
	err = binary.Write(out, binary.BigEndian, src.Header)
	if err != nil {
		return
	}

	n += 8 // Header

	err = binary.Write(out, binary.BigEndian, src.Data.PayloadHeader)
	if err != nil {
		return
	}

	n += 24 // PayloadHeader

	var chunkPayloadSize = int(src.Header.Size) - src.Data.PayloadHeader.Len() -
		int(src.Data.PayloadHeader.PaddingSize)

	err = safeWriter(out, src.Data.Payload[:chunkPayloadSize])
	if err != nil {
		return
	}

	n += int64(chunkPayloadSize)

	err = safeWriter(out, bytes.Repeat([]byte{0}, int(src.Data.PayloadHeader.PaddingSize)))
	if err != nil {
		return
	}

	n += int64(src.Data.PayloadHeader.PaddingSize)

	return
}

func safeWriter(out io.Writer, data []byte) error {
	dataLen := len(data)

	n, err := out.Write(data)
	if n != dataLen {
		fmt.Printf("expected to write %#x bytes but only did %#x\n", dataLen, n)
	}

	return err
}

func compressDict(dictName string, raw [][]Entry) PayloadData {
	fixed := PayloadFixedData{
		NumberOfDictionary: uint32(len(raw)),
		ItemsPerDictionary: uint16(len(raw[0])),
		PayloadNameOffset:  0x7, // always 0x7
	}

	flex := PayloadFlexData{}

	// Always starts with <NULL>\u0000
	flex.StringArray = append(flex.StringArray, []byte{0x3C, 0x4E, 0x55, 0x4C, 0x4C, 0x3E, 0x00}...)

	// First element always name of dictionary
	flex.StringArray = append(flex.StringArray, stringToC(dictName)...)

	for i, dict := range raw {
		for _, v := range dict {
			if i == 0 {
				// add keys only once
				valueType := v.ToByte()
				flex.SharedArray = append(flex.SharedArray, valueType)

				var keyOffset = make([]byte, 4)
				binary.BigEndian.PutUint32(keyOffset, uint32(len(flex.StringArray)))
				flex.SharedArray = append(flex.SharedArray, keyOffset...)
				flex.StringArray = append(flex.StringArray, stringToC(v.Key)...)
			}

			// if its string we add value to string array and then save pointer to that string to shared/unique array
			if v.Type.Name == "String" {
				var valueOffset = make([]byte, 4)
				binary.BigEndian.PutUint32(valueOffset, uint32(len(flex.StringArray)))
				flex.StringArray = append(flex.StringArray, stringToC(string(v.Value))...)

				if v.Recurring && i == 0 { // if its recurring - add only at first time
					flex.SharedArray = append(flex.SharedArray, valueOffset...)
				} else if !v.Recurring {
					flex.UniqueArray = append(flex.UniqueArray, valueOffset...)
				}

				continue
			}

			// if its Bytes we add value to bytes array and then save pointer to start and end to shared/unique array
			if v.Type.Name == "Bytes" {
				var valueOffset = make([]byte, 4)
				var endOffset = make([]byte, 4)
				binary.BigEndian.PutUint32(valueOffset, uint32(len(flex.ByteArray)))
				flex.ByteArray = append(flex.ByteArray, v.Value...)
				binary.BigEndian.PutUint32(endOffset, uint32(len(flex.ByteArray)))

				if v.Recurring && i == 0 { // if its recurring - add only at first time
					flex.SharedArray = append(flex.SharedArray, valueOffset...)
					flex.SharedArray = append(flex.SharedArray, endOffset...)
				} else if !v.Recurring {
					flex.UniqueArray = append(flex.UniqueArray, valueOffset...)
					flex.UniqueArray = append(flex.UniqueArray, endOffset...)
				}

				continue
			}

			if v.Recurring && i == 0 { // if its recurring - add only at first time
				flex.SharedArray = append(flex.SharedArray, v.Value[:v.Type.Size]...)
			} else if !v.Recurring {
				flex.UniqueArray = append(flex.UniqueArray, v.Value[:v.Type.Size]...)
			}
		}
	}

	fixed.UniqueArraySizePerDictionary = uint16(len(flex.UniqueArray) / len(raw))
	// 24 is the length of fixed part of payload
	fixed.UniqueArrayOffset = uint32(len(flex.SharedArray) + 24)
	fixed.StringArrayOffset = uint32(len(flex.UniqueArray)) + fixed.UniqueArrayOffset
	fixed.ByteArrayOffset = uint32(len(flex.StringArray)) + fixed.StringArrayOffset

	return PayloadData{fixed, flex}
}

// adds \u0000 to the end of string
func stringToC(s string) []byte {
	result := []byte(s)
	result = append(result, 0x00)
	return result
}

func compressPayload(src Payload) ([]byte, error) {
	var result bytes.Buffer

	err := binary.Write(&result, binary.BigEndian, src.Header)
	if err != nil {
		return nil, err
	}

	if err = binary.Write(&result, binary.BigEndian, src.PayloadData.PayloadFixedData); err != nil {
		return nil, err
	}

	if err = safeWriter(&result, src.PayloadData.PayloadFlexData.SharedArray); err != nil {
		return nil, err
	}
	if err = safeWriter(&result, src.PayloadData.PayloadFlexData.UniqueArray); err != nil {
		return nil, err
	}
	if err = safeWriter(&result, src.PayloadData.PayloadFlexData.StringArray); err != nil {
		return nil, err
	}
	if err = safeWriter(&result, src.PayloadData.PayloadFlexData.ByteArray); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}
