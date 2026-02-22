package loader_slice

import (
	"bufio"
	"io"
	"log/slog"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// detectEncoding reads the beginning of the file to guess the encoding.
// It returns a reader that converts the content to UTF-8.
func detectEncoding(r io.Reader) (io.Reader, error) {
	slog.Debug("ENTER detectEncoding")

	bufReader := bufio.NewReader(r)
	peekBytes, err := bufReader.Peek(4096)
	if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
		return nil, err
	}

	if reader, ok := checkBOM(bufReader, peekBytes); ok {
		return reader, nil
	}

	if checkUTF8(peekBytes) {
		return bufReader, nil
	}

	if reader, ok := tryShiftJIS(bufReader, peekBytes); ok {
		return reader, nil
	}

	// Default fallback: Windows-1252 (CP1252) for European languages
	slog.Debug("detectEncoding: falling back to Windows-1252")
	return transform.NewReader(bufReader, charmap.Windows1252.NewDecoder()), nil
}

// checkBOM checks for a UTF-8 BOM (Byte Order Mark) and discards it if found.
func checkBOM(bufReader *bufio.Reader, peekBytes []byte) (io.Reader, bool) {
	slog.Debug("ENTER checkBOM")

	if len(peekBytes) >= 3 && peekBytes[0] == 0xEF && peekBytes[1] == 0xBB && peekBytes[2] == 0xBF {
		bufReader.Discard(3)
		return bufReader, true
	}
	return nil, false
}

// checkUTF8 returns true if the peeked bytes are valid UTF-8.
func checkUTF8(peekBytes []byte) bool {
	slog.Debug("ENTER checkUTF8")
	return utf8.Valid(peekBytes)
}

// tryShiftJIS checks if the peeked bytes look like Shift_JIS and returns a decoding reader.
func tryShiftJIS(bufReader *bufio.Reader, peekBytes []byte) (io.Reader, bool) {
	slog.Debug("ENTER tryShiftJIS")

	if isShiftJIS(peekBytes) {
		return transform.NewReader(bufReader, japanese.ShiftJIS.NewDecoder()), true
	}
	return nil, false
}

// isShiftJIS attempts to detect Shift_JIS byte sequences.
func isShiftJIS(b []byte) bool {
	// Check for characteristic Shift_JIS byte ranges
	// Lead byte: 0x81-0x9F, 0xE0-0xEF
	for i := 0; i < len(b); i++ {
		c := b[i]
		if (c >= 0x81 && c <= 0x9F) || (c >= 0xE0 && c <= 0xEF) {
			if i+1 < len(b) {
				c2 := b[i+1]
				// Trail byte: 0x40-0x7E, 0x80-0xFC
				if (c2 >= 0x40 && c2 <= 0x7E) || (c2 >= 0x80 && c2 <= 0xFC) {
					return true
				}
			}
		}
	}
	return false
}
