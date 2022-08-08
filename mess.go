package parser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
)

type USMInfo struct {
	CRID            Chunk
	HDRInfo         map[[4]byte]Chunk
	Metadata        map[[4]byte]Chunk
	AudioStreams    []Chunk
	VideoStreams    []Chunk
	SubtitleStreams []Chunk
}

var customOrder = map[int][4]byte{
	0: _SFV,
	1: _SFA,
	2: _SBT,
}

var (
	CRID = [4]byte{0x43, 0x52, 0x49, 0x44}
	_SFV = [4]byte{0x40, 0x53, 0x46, 0x56}
	_SFA = [4]byte{0x40, 0x53, 0x46, 0x41}
	_SBT = [4]byte{0x40, 0x53, 0x42, 0x54}
	_UTF = [4]byte{0x40, 0x55, 0x54, 0x46}

	HCA_ = [4]byte{0x48, 0x43, 0x41, 0x00}

	HeaderEnd = [32]byte{
		0x23, 0x48, 0x45, 0x41, 0x44, 0x45, 0x52, 0x20, 0x45, 0x4E, 0x44, 0x20,
		0x20, 0x20, 0x20, 0x20, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D,
		0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x00,
	} // #HEADER END     ===============\u0000

	MetadataEnd = [32]byte{
		0x23, 0x4D, 0x45, 0x54, 0x41, 0x44, 0x41, 0x54, 0x41, 0x20, 0x45, 0x4E,
		0x44, 0x20, 0x20, 0x20, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D,
		0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x00,
	} // #METADATA END   ===============\u0000

	ContentsEnd = [32]byte{
		0x23, 0x43, 0x4F, 0x4E, 0x54, 0x45, 0x4E, 0x54, 0x53, 0x20, 0x45, 0x4E,
		0x44, 0x20, 0x20, 0x20, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D,
		0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x3D, 0x00,
	} // #CONTENTS END   ===============\u0000
)

func ParseFile(src *os.File) (*USMInfo, error) {
	var result USMInfo
	result.HDRInfo = make(map[[4]byte]Chunk, 0)
	result.Metadata = make(map[[4]byte]Chunk, 0)

	var pos int
	for {
		chunkInfo, err := ReadChunk(src, pos)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read chunk: %w", err)
		}

		if chunkInfo.Header.ID == CRID {
			result.CRID = chunkInfo
			continue
		}

		if chunkInfo.Data.PayloadHeader.PayloadType == PayloadTypeHeader {
			result.HDRInfo[chunkInfo.Header.ID] = chunkInfo
			continue
		}

		if chunkInfo.Data.PayloadHeader.PayloadType == PayloadTypeSeek {
			result.Metadata[chunkInfo.Header.ID] = chunkInfo
			continue
		}

		if chunkInfo.Data.PayloadHeader.PayloadType == PayloadTypeStream {
			switch chunkInfo.Header.ID {
			case _SFV:
				result.VideoStreams = append(result.VideoStreams, chunkInfo)
			case _SFA:
				result.AudioStreams = append(result.AudioStreams, chunkInfo)
			case _SBT:
				result.SubtitleStreams = append(result.SubtitleStreams, chunkInfo)
			}
			continue
		}
	}

	return &result, nil
}

func (s *USMInfo) PrepareStreams() *USMInfo {
	// videos don't have chunks with same frame time
	sort.SliceStable(s.VideoStreams, func(i, j int) bool {
		return s.VideoStreams[i].Data.PayloadHeader.FrameTime <
			s.VideoStreams[j].Data.PayloadHeader.FrameTime
	})

	s.VideoStreams = addContentsEnd(s.VideoStreams)

	sort.SliceStable(s.AudioStreams, func(i, j int) bool {
		// audio streams include additional HCA header, which should be first
		if bytes.HasPrefix(s.AudioStreams[i].Data.Payload, HCA_[:]) {
			return true
		}

		return s.AudioStreams[i].Data.PayloadHeader.FrameTime <
			s.AudioStreams[j].Data.PayloadHeader.FrameTime
	})

	s.AudioStreams = addContentsEnd(s.AudioStreams)

	sort.SliceStable(s.SubtitleStreams, func(i, j int) bool {
		if s.SubtitleStreams[i].Data.PayloadHeader.FrameTime ==
			s.SubtitleStreams[j].Data.PayloadHeader.FrameTime {
			switch bytes.Compare(s.SubtitleStreams[i].Data.Payload[:1], s.SubtitleStreams[j].Data.Payload[:1]) {
			case -1:
				return true
			default:
				return false
			}
		}

		return s.SubtitleStreams[i].Data.PayloadHeader.FrameTime <
			s.SubtitleStreams[j].Data.PayloadHeader.FrameTime
	})

	s.SubtitleStreams = addContentsEnd(s.SubtitleStreams)

	return s
}

func addContentsEnd(src []Chunk) []Chunk {
	end := ContentsEndChunk(src[len(src)-1].Header.ID)
	// make FrameTime just a bit higher so it goes next after last element
	end.Data.PayloadHeader.FrameTime = src[len(src)-1].Data.PayloadHeader.FrameTime + 1

	return append(src, end)
}

func (s *USMInfo) WriteTo(seeker io.WriteSeeker) error {
	var videoSeekPos int64

	videoOffsets := make([]int64, 0)

	var pos int64
	n, err := WriteChunk(s.CRID, seeker)
	if err != nil {
		return err
	}
	pos += n

	for i := 0; i < len(s.HDRInfo); i++ {
		n, err = WriteChunk(s.HDRInfo[customOrder[i]], seeker)
		if err != nil {
			return err
		}
		pos += n
	}
	for i := 0; i < len(s.HDRInfo); i++ {
		n, err = WriteChunk(HeaderEndChunk(customOrder[i]), seeker)
		if err != nil {
			return err
		}
		pos += n
	}

	for i := 0; i < len(s.Metadata); i++ {
		id := customOrder[i]

		if id == _SFV {
			// skip video seek data for now
			videoSeekPos = pos

			// we need 12 (0xC) bytes per each video entry + 144 (0x90) bytes for other data
			pos, err = seeker.Seek(getSizeForVideoSeek(len(s.VideoStreams)), io.SeekCurrent)
			//pos, err = seeker.Seek(int64(s.Metadata[id].Header.Size+8), io.SeekCurrent)
			if err != nil {
				return err
			}
			continue
		}

		n, err = WriteChunk(s.Metadata[id], seeker)
		if err != nil {
			return err
		}
		pos += n
	}
	for i := 0; i < len(s.Metadata); i++ {
		n, err = WriteChunk(MetadataEndChunk(customOrder[i]), seeker)
		if err != nil {
			return err
		}
		pos += n
	}

	var c Chunk
	// write first video chunk
	c, s.VideoStreams = pop(s.VideoStreams)
	n, err = WriteChunk(c, seeker)
	if err != nil {
		return err
	}

	videoOffsets = append(videoOffsets, pos)
	pos += n

	// then write 2 audio chunks (HCA header + 1st chunk)
	c, s.AudioStreams = pop(s.AudioStreams)
	n, err = WriteChunk(c, seeker)
	if err != nil {
		return err
	}
	pos += n

	c, s.AudioStreams = pop(s.AudioStreams)
	n, err = WriteChunk(c, seeker)
	if err != nil {
		return err
	}
	pos += n

	// After this write chunks based on their frame time

	chunks := append(s.VideoStreams, s.AudioStreams...)
	chunks = append(chunks, s.SubtitleStreams...)

	sort.SliceStable(chunks, func(i, j int) bool {
		iFrame := getFrameSeconds(chunks[i].Data.PayloadHeader)
		jFrame := getFrameSeconds(chunks[j].Data.PayloadHeader)

		//if iFrame != jFrame {
		return iFrame < jFrame
		//}
	})

	for _, c = range chunks {
		if c.Data.PayloadHeader.PayloadType == PayloadTypeEnd {
			c.Data.PayloadHeader.FrameTime = 0
		}

		if c.Header.ID == _SFV {
			// store offsets for video chunks
			videoOffsets = append(videoOffsets, pos)
		}

		n, err = WriteChunk(c, seeker)
		if err != nil {
			return err
		}
		pos += n

	}

	c, err = generateVideoSeek(videoOffsets)
	if err != nil {
		return err
	}

	_, err = seeker.Seek(videoSeekPos, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = WriteChunk(c, seeker)
	if err != nil {
		return err
	}

	return nil
}

func getSizeForVideoSeek(videos int) int64 {
	c := int64(videos/30 + 1)
	size := 12*c + 144

	if remainder := size % 0x10; remainder != 0 {
		size += remainder
	}

	return size
}

func pop(src []Chunk) (Chunk, []Chunk) {
	return src[0], src[1:]
}

func generateVideoSeek(videoOffsets []int64) (Chunk, error) {
	c := Chunk{
		Header: Header{
			ID: _SFV,
			//Size: 0,
		},
		Data: Data{
			PayloadHeader: PayloadHeader{
				Offset:        0x18,
				PaddingSize:   0,
				ChannelNumber: 0,
				PayloadType:   PayloadTypeSeek,
				FrameTime:     0,
				FrameRate:     0x1e,
			},
			//Payload: content,
		},
	}

	data := make([][]Entry, 0)

	for k, v := range videoOffsets {
		if k != 0 && k%30 != 0 {
			continue
		}

		var elData = make([]byte, 8)
		var elFrame = make([]byte, 4)
		binary.BigEndian.PutUint64(elData, uint64(v))
		binary.BigEndian.PutUint32(elFrame, uint32(k))

		el := []Entry{
			{Key: "ofs_byte", Type: values[0x16], Recurring: false, Value: elData},
			{Key: "ofs_frmid", Type: values[0x15], Recurring: false, Value: elFrame},
			{Key: "num_skip", Type: values[0x13], Recurring: true, Value: []byte{0x00, 0x00}},
			{Key: "resv", Type: values[0x13], Recurring: true, Value: []byte{0x00, 0x00}},
		}

		data = append(data, el)
	}

	payloadContent := compressDict("VIDEO_SEEKINFO", data)

	if remainder := (payloadContent.Size() + 8) % 0x10; remainder != 0 {
		c.Data.PayloadHeader.PaddingSize = uint16(0x10 - remainder)
	}

	compressedPayload, err := compressPayload(Payload{
		Header:      Header{ID: _UTF, Size: int32(payloadContent.Size())},
		PayloadData: payloadContent,
	})

	if err != nil {
		return Chunk{}, fmt.Errorf("can't compress payload: %w", err)
	}

	c.Data.Payload = compressedPayload

	c.Header.Size = int32(payloadContent.Size()) +
		int32(c.Data.PayloadHeader.PaddingSize) +
		8 + // UTF header inside payload
		24 // 24 - PayloadHeader

	return c, nil
}

func ContentsEndChunk(id [4]byte) Chunk {
	return makeEndChunk(id, ContentsEnd[:])
}

func HeaderEndChunk(id [4]byte) Chunk {
	return makeEndChunk(id, HeaderEnd[:])
}

func MetadataEndChunk(id [4]byte) Chunk {
	return makeEndChunk(id, MetadataEnd[:])
}

func makeEndChunk(id [4]byte, content []byte) Chunk {
	return Chunk{
		Header: Header{
			ID:   id,
			Size: 0x38,
		},
		Data: Data{
			PayloadHeader: PayloadHeader{
				Offset:        0x18,
				PaddingSize:   0,
				ChannelNumber: 0,
				PayloadType:   PayloadTypeEnd,
				FrameTime:     0,
				FrameRate:     0x1e,
			},
			Payload: content,
		},
	}
}

func getFrameSeconds(src PayloadHeader) int32 {
	return src.FrameTime * 1000 / src.FrameRate
}

func ReplaceAudio(in1, in2 *USMInfo) *USMInfo {

	// ignore CRID for now kek
	in1.HDRInfo[_SFA] = in2.HDRInfo[_SFA]
	in1.Metadata[_SFA] = in2.Metadata[_SFA]
	in1.AudioStreams = in2.AudioStreams

	return in1
}
