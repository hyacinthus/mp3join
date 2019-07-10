package mp3join

import (
	"bytes"
	"io"

	"github.com/dmulholland/mp3lib"
)

// Joiner is a mp3 file joiner
type Joiner struct {
	totalFrames  uint32 // for gen vbr header
	totalFiles   int    // used in log only
	firstBitRate int
	isVBR        bool
	vbrHeader    []byte
	id3v2Tag     []byte
	buffer       *bytes.Buffer
}

// New Create a mp3 joiner
func New() *Joiner {
	return &Joiner{
		buffer: bytes.NewBuffer(nil),
	}
}

// Append join a mp3 file to the end
func (j *Joiner) Append(in io.Reader) error {
	isFirstFrame := true
LOOP:
	for {
		var frame *mp3lib.MP3Frame
		// Read the next object from the input file.
		obj := mp3lib.NextObject(in)
		switch obj := obj.(type) {
		case *mp3lib.MP3Frame:
			frame = obj
		case *mp3lib.ID3v1Tag:
			// ignore ID3 v1 tag
			continue
		case *mp3lib.ID3v2Tag:
			// Copy the first met ID3v2 tag
			if len(j.id3v2Tag) == 0 {
				j.id3v2Tag = obj.RawBytes
			}
			continue
		case nil:
			break LOOP
		}

		// Skip the first frame if it's a VBR header.
		if isFirstFrame {
			isFirstFrame = false
			if mp3lib.IsXingHeader(frame) || mp3lib.IsVbriHeader(frame) {
				continue
			}
		}

		// If we detect more than one bitrate we'll need to add a VBR
		// header to the output file.
		if j.firstBitRate == 0 {
			j.firstBitRate = frame.BitRate
		} else if frame.BitRate != j.firstBitRate {
			j.isVBR = true
		}

		// Write the frame to the output file.
		_, err := j.buffer.Write(frame.RawBytes)
		if err != nil {
			return err
		}

		j.totalFrames++
	}

	// If we detected multiple bitrates, prepend a VBR header to the file.
	if j.isVBR {
		j.vbrHeader = mp3lib.NewXingHeader(j.totalFrames, uint32(j.buffer.Len())).RawBytes
	}

	j.totalFiles++
	return nil
}

// Reader always return a new Reader from mp3 beginning
func (j *Joiner) Reader() *bytes.Reader {
	return bytes.NewReader(bytes.Join([][]byte{j.id3v2Tag, j.vbrHeader, j.buffer.Bytes()}, nil))
}

// FileCount is the count of files be joined
func (j *Joiner) FileCount() int {
	return j.totalFiles
}

// Len is the total size of the output mp3 file
func (j *Joiner) Len() int {
	return len(j.id3v2Tag) + len(j.vbrHeader) + j.buffer.Len()
}
