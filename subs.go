package parser

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

const (
	Chinese    = "cn"
	English    = "en"
	Thai       = "th"
	Vietnamese = "vn"
	French     = "fr"
	German     = "de"
	Indonesian = "id"

	Undefined = "xx"
)

// GetSubs will try to read subtitles from provided reader.
// Returns array of Subtitles for each language, ready for further processing and error if any
func GetSubs(src io.Reader) (subs map[string][]Subtitle, err error) {
	subs = make(map[string][]Subtitle, 0)

	for {
		var c Chunk
		c, err = ReadChunk(src, 0)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return
			}
		}

		if c.Header.IDString() != "@SBT" {
			continue
		}

		if c.Data.PayloadHeader.PayloadType != PayloadTypeStream {
			// skip headers
			continue
		}

		var sub Subtitle
		sub, err = ReadSubtitleData(c.Data.Payload)
		if err != nil {
			return
		}

		subs[sub.SubtitleHeader.GetLang()] = append(subs[sub.SubtitleHeader.GetLang()], sub)
	}

	return subs, nil
}

func SubsToSrt(src map[string][]Subtitle) map[string]bytes.Buffer {
	result := make(map[string]bytes.Buffer, 0)

	for lang, subs := range src {
		var b = result[lang]
		for i, sub := range subs {
			b.WriteString(strconv.Itoa(i + 1))
			b.Write([]byte{0x0D, 0x0A})
			b.WriteString(parseTime(sub.SubtitleHeader.FrameTime))
			b.WriteString(" --> ")
			b.WriteString(parseTime(sub.SubtitleHeader.FrameTime + sub.SubtitleHeader.FrameEnd))
			b.Write([]byte{0x0D, 0x0A})
			b.Write(sub.SubtitleString)
			b.Write([]byte{0x0D, 0x0A})
		}
		result[lang] = b
	}

	return result
}

func SubsToTxt(src map[string][]Subtitle) map[string]bytes.Buffer {
	result := make(map[string]bytes.Buffer, 0)

	for lang, subs := range src {
		var b = result[lang]
		for i, sub := range subs {
			if i == 0 {
				// first line is display interval - always 1000ms
				b.WriteString(strconv.Itoa(1000))
				b.Write([]byte{0x0D, 0x0A})
			}
			b.WriteString(strconv.FormatUint(uint64(sub.SubtitleHeader.FrameTime), 10))
			b.WriteString(", ")
			b.WriteString(strconv.FormatUint(uint64(sub.SubtitleHeader.FrameTime+sub.SubtitleHeader.FrameEnd), 10))
			b.WriteString(", ")
			b.Write(sub.SubtitleString)
		}
		result[lang] = b
	}

	return result
}

func parseTime(t uint32) string {
	ms := t % 1000
	s := (t / 1000) % 60
	m := (t / 1000 / 60) % 60
	h := t / 1000 / 60 / 60

	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
