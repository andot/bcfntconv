package convert

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const HeaderSize = 0x80

// BcfntToShared converts a bcfnt byte slice to shared_font.bin format.
func BcfntToShared(b []byte) ([]byte, error) {
	if len(b) < 4 {
		return nil, errors.New("input too small to be a valid bcfnt file")
	}
	// Validate magic header "CFNT"
	if !bytes.Equal(b[:4], []byte("CFNT")) {
		return nil, errors.New("bcfnt validation failed: missing CFNT header")
	}

	// modify copy
	mod := make([]byte, len(b))
	copy(mod, b)
	mod[3] = 0x55

	header := make([]byte, HeaderSize)
	header[0] = 0x02
	header[4] = 0x01
	binary.LittleEndian.PutUint32(header[8:], uint32(len(b)))

	out := make([]byte, 0, HeaderSize+len(mod))
	out = append(out, header...)
	out = append(out, mod...)
	return out, nil
}

// SharedToBcfnt converts a shared_font.bin byte slice to bcfnt format.
func SharedToBcfnt(b []byte) ([]byte, error) {
	if len(b) < HeaderSize+4 {
		return nil, errors.New("input too small to be a valid shared_font.bin")
	}
	header := b[:HeaderSize]
	if header[0] != 0x02 {
		return nil, errors.New("shared_font.bin validation failed: header[0] != 0x02")
	}
	if header[4] != 0x01 {
		return nil, errors.New("shared_font.bin validation failed: header[4] != 0x01")
	}
	origLen := binary.LittleEndian.Uint32(header[8:12])
	if int(origLen) != len(b)-HeaderSize {
		return nil, errors.New("shared_font.bin validation failed: original length mismatch")
	}

	body := make([]byte, len(b)-HeaderSize)
	copy(body, b[HeaderSize:])
	if len(body) < 4 {
		return nil, errors.New("body too small after header removal")
	}
	body[3] = 0x54
	return body, nil
}
