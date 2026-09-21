// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package stt

import (
	"bytes"
	"encoding/binary"
)

const (
	SampleRate = 16000
	channels   = 1
	bitDepth   = 16
)

// 16-bit mono PCM at SampleRate.
func WAV(pcm []byte) []byte {
	var out bytes.Buffer
	out.Grow(44 + len(pcm))

	out.WriteString("RIFF")
	binary.Write(&out, binary.LittleEndian, uint32(36+len(pcm)))
	out.WriteString("WAVE")

	out.WriteString("fmt ")
	binary.Write(&out, binary.LittleEndian, uint32(16))
	binary.Write(&out, binary.LittleEndian, uint16(1))
	binary.Write(&out, binary.LittleEndian, uint16(channels))
	binary.Write(&out, binary.LittleEndian, uint32(SampleRate))
	binary.Write(&out, binary.LittleEndian, uint32(SampleRate*channels*bitDepth/8))
	binary.Write(&out, binary.LittleEndian, uint16(channels*bitDepth/8))
	binary.Write(&out, binary.LittleEndian, uint16(bitDepth))

	out.WriteString("data")
	binary.Write(&out, binary.LittleEndian, uint32(len(pcm)))
	out.Write(pcm)

	return out.Bytes()
}
