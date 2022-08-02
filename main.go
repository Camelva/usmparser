package parser

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// DumpFile tries to read file `path` and write result to file `outPath`
func DumpFile(path string, outPath string) {
	src, err := os.Open(path)
	if err != nil {
		log.Fatal("can't open source file: ", err)
	}

	out, err := os.Create(outPath)
	if err != nil {
		log.Fatal("can't create output file: ", err)
	}

	defer func() {
		_ = src.Close()
		_ = out.Close()
	}()

	err = DumpAllChunks(src, out)
	if err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
	}
}

func Kek(src io.Reader) {
	CRIDinfo, err := ReadChunk(src)
	if err != nil {
		panic("hi")
	}

	fmt.Println(PayloadType[CRIDinfo.PayloadType])

	result, err := ParsePayload(CRIDinfo.Payload)
	if err != nil {
		fmt.Println(err)
	}

	dict, name, err := BuildDict(result)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(name)
	fmt.Println(dict)
}

func DumpAllChunks(src io.Reader, out io.Writer) (err error) {
	// start json array
	if _, err = out.Write([]byte("[\n")); err != nil {
		return fmt.Errorf("can't write result: %w", err)
	}

	var i = 0
	for {
		i++
		chunkInfo, err := ReadChunk(src)
		if err != nil {
			if err == io.EOF {
				return err
			}
			return fmt.Errorf("read chunk: %w", err)
		}

		fmt.Printf("== Chunk #%d ==\n", i)

		j := map[string]string{
			"Chunk":         strconv.Itoa(i),
			"Header":        chunkInfo.Header.String(),
			"PayloadHeader": chunkInfo.Data.PayloadHeader.String(),
		}

		if chunkInfo.Data.PayloadHeader.PayloadType == PayloadTypeEnd {
			payload, err := ParsePayloadEnd(chunkInfo.Payload)
			if err != nil {
				fmt.Println("can't parse payload end: ", err)
			} else {
				j["Payload"] = payload
			}
		}

		if chunkInfo.Data.PayloadHeader.PayloadType == PayloadTypeHeader ||
			chunkInfo.Data.PayloadHeader.PayloadType == PayloadTypeSeek {
			payload, err := ParsePayload(chunkInfo.Payload)
			if err != nil {
				fmt.Println("can't parse payload: ", err)
			} else {
				j["Payload"] = payload.String()
			}

			name, dict, err := BuildDict(payload)
			if err != nil {
				fmt.Println("can't build dict: ", err)
			} else {
				j[name] = fmt.Sprintf("%v", dict)
			}
		}

		if chunkInfo.Data.PayloadHeader.PayloadType == PayloadTypeStream &&
			chunkInfo.Header.IDString() == "@SBT" {
			// we can parse subtitle stream data
			sub, err := ReadSubtitleData(chunkInfo.Data.Payload)
			if err != nil {
				fmt.Println("can't read subtitle data: ", err)
			} else {
				j["SubtitleInfo"] = sub.String()
			}
		}

		result, err := json.MarshalIndent(j, "", "\t")
		if err != nil {
			return fmt.Errorf("encoding err: %w", err)
		}

		result = append(result, []byte(",\n")...)

		_, err = out.Write(result)
		if err != nil {
			return fmt.Errorf("can't write result: %w", err)
		}
	}
}

func ReadSubtitleData(raw []byte) (result Subtitle, err error) {
	src := bytes.NewReader(raw)

	head := new(SubtitleHeader)
	if err = binary.Read(src, binary.LittleEndian, head); err != nil {
		return
	}

	var subText = make([]byte, head.GetStringSize())
	if err = safeRead(src, subText); err != nil {
		return
	}

	result.SubtitleHeader = *head
	result.SubtitleString = subText
	return
}

func ReadChunk(src io.Reader) (result Chunk, err error) {
	result.Header, err = ReadHeader(src)
	if err != nil {
		return result, err
	}
	result.Data, err = ReadChunkData(src, result.SizeUInt())
	//result.Data = *ReadPayloadHeader(src)

	//fmt.Println(result)
	return result, err
}

func ReadChunkData(src io.Reader, size uint32) (result Data, err error) {
	result.PayloadHeader, err = ReadPayloadHeader(src)

	_payload := make([]byte,
		int(size)-int(result.PayloadHeader.GetPaddingSize())-result.PayloadHeader.Len())
	err = safeRead(src, _payload)
	result.Payload = _payload

	// skip padding
	_ = safeRead(src, make([]byte, result.PayloadHeader.GetPaddingSize()))

	return result, err
}

func ParsePayloadEnd(raw []byte) (string, error) {
	src := bytes.NewReader(raw)

	return ReadStringAt(src, 0)
}

func ParsePayload(raw []byte) (result Payload, err error) {
	src := bytes.NewReader(raw)

	if err = binary.Read(src, binary.BigEndian, &result.Header); err != nil {
		return
	}

	fixedData := PayloadFixedData{}
	if err = binary.Read(src, binary.BigEndian, &fixedData); err != nil {
		return
	}

	_sharedArray := make([]byte, fixedData.GetUniqueArrayOffset()-fixedData.Len())
	_uniqueArray := make([]byte, fixedData.GetStringArrayOffset()-fixedData.GetUniqueArrayOffset())
	_stringArray := make([]byte, fixedData.GetByteArrayOffset()-fixedData.GetStringArrayOffset())
	_byteArray := make([]byte, result.Header.SizeUInt()-fixedData.GetByteArrayOffset())

	_ = safeRead(src, _sharedArray)
	_ = safeRead(src, _uniqueArray)
	_ = safeRead(src, _stringArray)
	_ = safeRead(src, _byteArray)

	flexData := PayloadFlexData{
		SharedArray: _sharedArray,
		UniqueArray: _uniqueArray,
		StringArray: _stringArray,
		ByteArray:   _byteArray,
	}

	result.PayloadData.PayloadFixedData = fixedData
	result.PayloadData.PayloadFlexData = flexData

	return result, nil
}

func ReadHeader(src io.Reader) (result Header, err error) {
	return result, binary.Read(src, binary.BigEndian, &result)
}

func ReadPayloadHeader(src io.Reader) (result PayloadHeader, err error) {
	return result, binary.Read(src, binary.BigEndian, &result)
}

func safeRead(src io.Reader, dst []byte) error {
	dstLength := len(dst)
	n, err := src.Read(dst)
	if n != dstLength {
		fmt.Printf("expected %d bytes but got %d", dstLength, n)
		return nil
	}

	return err
}

func BuildDict(src Payload) (name string, result []map[string]Entry, err error) {
	sharedArray := bytes.NewReader(src.PayloadData.SharedArray)
	uniqueArray := bytes.NewReader(src.PayloadData.UniqueArray)
	stringsArray := bytes.NewReader(src.PayloadData.StringArray)
	bytesArray := bytes.NewReader(src.PayloadData.ByteArray)

	result = make([]map[string]Entry, 0)

	for i := 1; i <= int(src.GetNumberOfDictionary()); i++ {
		var dict = make(map[string]Entry, 0)

		for ii := 1; ii <= int(src.GetItemsPerDictionary()); ii++ {
			var itemType byte
			itemType, err = sharedArray.ReadByte()
			if err != nil {
				return
			}

			valueType, isUnique := GetValue(itemType)

			keyAddr := make([]byte, 4)
			if err = safeRead(sharedArray, keyAddr); err != nil {
				return
			}

			var key string
			key, err = ReadStringAt(stringsArray, int(binary.BigEndian.Uint32(keyAddr)))
			if err != nil {
				return
			}

			valueInfo := make([]byte, valueType.Size)

			if isUnique {
				err = safeRead(uniqueArray, valueInfo)
			} else {
				err = safeRead(sharedArray, valueInfo)
			}
			if err != nil {
				return
			}

			if valueType.Name == "String" {
				// we have to find actual value in strings array
				var actualValue string
				actualValue, err = ReadStringAt(stringsArray, int(binary.BigEndian.Uint32(valueInfo)))
				if err != nil {
					return
				}

				dict[key] = Entry{valueType, []byte(actualValue)}
				continue
			}

			if valueType.Name == "Bytes" {
				// we need to find end of data as well
				valueInfoEnd := make([]byte, valueType.Size)
				if isUnique {
					err = safeRead(uniqueArray, valueInfoEnd)
				} else {
					err = safeRead(sharedArray, valueInfoEnd)
				}
				if err != nil {
					return
				}

				length := binary.BigEndian.Uint32(valueInfoEnd) - binary.BigEndian.Uint32(valueInfo)
				bytesVal := make([]byte, length)
				_, err = bytesArray.ReadAt(bytesVal, int64(binary.BigEndian.Uint32(valueInfo)))
				if err != nil {
					return
				}

				dict[key] = Entry{valueType, bytesVal}
				continue
			}

			dict[key] = Entry{valueType, valueInfo}
			continue
		}

		result = append(result, dict)

		_, err = sharedArray.Seek(0, io.SeekStart)
		if err != nil {
			return
		}
	}

	name, err = ReadStringAt(stringsArray, int(src.GetPayloadNameOffset()))
	return
}

func ReadStringAt(src *bytes.Reader, offset int) (string, error) {
	_, err := src.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return "", err
	}

	var result strings.Builder

	for {
		b, err := src.ReadByte()
		if err != nil {
			break
		}

		if b == 0 {
			break
		}

		_ = result.WriteByte(b)
	}

	return result.String(), nil
}
