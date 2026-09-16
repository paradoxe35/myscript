package stt

import (
	"bytes"
	"testing"

	"github.com/go-audio/wav"
)

func TestWAVIsReadableMono16k(t *testing.T) {
	pcm := []byte{0x00, 0x00, 0xff, 0x7f, 0x00, 0x80}
	decoder := wav.NewDecoder(bytes.NewReader(WAV(pcm)))

	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoder.SampleRate != SampleRate || decoder.NumChans != 1 || decoder.BitDepth != 16 {
		t.Errorf("format = %d Hz, %d ch, %d bit", decoder.SampleRate, decoder.NumChans, decoder.BitDepth)
	}
	if len(buf.Data) != 3 || buf.Data[1] != 32767 || buf.Data[2] != -32768 {
		t.Errorf("samples = %v", buf.Data)
	}
}
