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
// Returns buffer of bytes for each language and error if any
func GetSubs(src io.Reader) (subs map[string]bytes.Buffer, err error) {
	subs = make(map[string]bytes.Buffer, 0)
	var counter = make(map[uint32]int, 0)

	for {
		var c Chunk
		c, err = ReadChunk(src, 0)
		if err != nil {
			break
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
			break
		}

		var b = subs[sub.SubtitleHeader.GetLang()]

		counter[sub.SubtitleHeader.Language] += 1

		b.WriteString(strconv.Itoa(counter[sub.SubtitleHeader.Language]))
		b.Write([]byte{0x0D, 0x0A})
		b.WriteString(parseTime(sub.SubtitleHeader.FrameTime))
		b.WriteString(" --> ")
		b.WriteString(parseTime(sub.SubtitleHeader.FrameTime + sub.SubtitleHeader.FrameEnd))
		b.Write([]byte{0x0D, 0x0A})
		b.Write(sub.SubtitleString)
		b.Write([]byte{0x0D, 0x0A})

		subs[sub.SubtitleHeader.GetLang()] = b
	}

	return
}

func parseTime(t uint32) string {
	ms := t % 1000
	s := (t / 1000) % 60
	m := (t / 1000 / 60) % 60
	h := t / 1000 / 60 / 60

	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
