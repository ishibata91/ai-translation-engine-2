package loader

import (
	"bufio"
	"io"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// detectEncoding reads the beginning of the file to guess the encoding.
// It returns a reader that converts the content to UTF-8.
func detectEncoding(r io.Reader) (io.Reader, error) {
	// Read a small chunk to detect encoding
	// 4KB should be enough to detect BOM or common characters
	bufReader := bufio.NewReader(r)
	peekBytes, err := bufReader.Peek(4096)
	if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
		return nil, err
	}

	// 1. Check for BOM (Byte Order Mark)
	if len(peekBytes) >= 3 && peekBytes[0] == 0xEF && peekBytes[1] == 0xBB && peekBytes[2] == 0xBF {
		// UTF-8 with BOM: Discard BOM and read as UTF-8
		bufReader.Discard(3)
		return bufReader, nil
	}

	// 2. Check if valid UTF-8
	if utf8.Valid(peekBytes) {
		return bufReader, nil
	}

	// 3. Check for specific Japanese characters (Shift_JIS)
	// Simple heuristic: check for common kana ranges in SJIS
	// This is not perfect but sufficient for Skyrim data context
	if isShiftJIS(peekBytes) {
		return transform.NewReader(bufReader, japanese.ShiftJIS.NewDecoder()), nil
	}

	// 4. Default fallback: Windows-1252 (CP1252) for European languages
	return transform.NewReader(bufReader, charmap.Windows1252.NewDecoder()), nil
}

// isShiftJIS attempts to detect Shift_JIS byte sequences.
func isShiftJIS(b []byte) bool {
	// Very basic check. Real implementation might need more robust analysis
	// or use a detection library like `chunyun` or `saintfish/chardet` if needed.
	// For now, we rely on the fact that if it's NOT UTF-8, and contains high bytes,
	// and we are expecting Skyrim data (often SJIS for JP output), we try SJIS.

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
