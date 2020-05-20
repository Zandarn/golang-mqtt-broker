// +build linux

package mqtt

import (
	"bytes"
	"testing"
)

func TestDecoding(t *testing.T) {

	buffer = []byte{16, 0, 0, 4, 77, 81, 84, 84, 4, 204}

	if decodeLength() != 0 {
		t.Errorf("decodeLength([0]) вернула не (0)\n")
	}

	buffer[1] = 127
	if decodeLength() != 127 {
		t.Errorf("decodeLength([127]) вернула не  (127)\n")
	}

	buffer[1] = 128
	buffer[2] = 1
	if decodeLength() != 128 {
		t.Errorf("decodeLength([128]) вернула не  (128)\n")
	}

	buffer[1] = 255
	buffer[2] = 127
	if decodeLength() != 16383 {
		t.Errorf("decodeLength([16383]) вернула не  (16383)\n")
	}

	buffer[1] = 128
	buffer[2] = 128
	buffer[3] = 1
	if decodeLength() != 16384 {
		t.Errorf("decodeLength([16384]) вернула не  (16384)\n")
	}

	buffer[1] = 255
	buffer[2] = 255
	buffer[3] = 127
	if decodeLength() != 2097151 {
		t.Errorf("decodeLength([2097151]) вернула не  (2097151)\n")
	}

	buffer[1] = 128
	buffer[2] = 128
	buffer[3] = 128
	buffer[4] = 1
	if decodeLength() != 2097152 {
		t.Errorf("decodeLength([2097152]) вернула не  (2097152)\n")
	}

	buffer[1] = 255
	buffer[2] = 255
	buffer[3] = 255
	buffer[4] = 127

	if decodeLength() != 268435455 {
		t.Errorf("decodeLength([268435455]) вернула не  (268435455)\n")
	}
}

func TestEncoding(t *testing.T) {

	lengths := map[int][]byte{
		0:         {0x00},
		127:       {0x7F},
		128:       {0x80, 0x01},
		16383:     {0xFF, 0x7F},
		16384:     {0x80, 0x80, 0x01},
		2097151:   {0xFF, 0xFF, 0x7F},
		2097152:   {0x80, 0x80, 0x80, 0x01},
		268435455: {0xFF, 0xFF, 0xFF, 0x7F},
	}

	for length, encoded := range lengths {
		if !bytes.Equal(encodeLength(length), encoded) {
			t.Errorf("encodeLength(%d) вернул не [0x%X]", length, encoded)
		}
	}
}
