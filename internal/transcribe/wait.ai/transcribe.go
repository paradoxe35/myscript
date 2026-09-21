// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package witai

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-audio/wav"
)

const (
	SampleRate           = 8000
	Channels             = 1
	BytesPerSample       = 2
	ChunkDurationSeconds = 20
)

type WitTranscriber struct {
	speechURL string
	client    *http.Client
	headers   map[string]string
}

func NewWitTranscriber(apiKey string) *WitTranscriber {
	return &WitTranscriber{
		speechURL: "https://api.wit.ai/speech",
		client:    &http.Client{Timeout: time.Second * 30},
		headers: map[string]string{
			"Authorization": "Bearer " + apiKey,
			"Accept":        "application/vnd.wit.20180705+json",
			"Content-Type":  "audio/raw;encoding=signed-integer;bits=16;rate=8000;endian=little",
		},
	}
}

type WitResponse struct {
	Text   string `json:"text,omitempty"`
	Legacy string `json:"_text,omitempty"`
}

func (w *WitTranscriber) Transcribe(chunk []byte) string {
	req, err := http.NewRequest("POST", w.speechURL, bytes.NewReader(chunk))
	if err != nil {
		slog.Error("SPEECH | Error creating request", "error", err)
		return ""
	}

	for key, value := range w.headers {
		req.Header.Add(key, value)
	}

	q := req.URL.Query()
	q.Add("verbose", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := w.client.Do(req)
	if err != nil {
		slog.Error("SPEECH | Error making request", "error", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("SPEECH | Error reading response body", "error", err)
		return ""
	}

	var witResp WitResponse
	if err := json.Unmarshal(body, &witResp); err != nil {
		slog.Error("SPEECH | Error unmarshaling response", "error", err)
		return ""
	}

	if witResp.Text != "" {
		return witResp.Text
	}
	return witResp.Legacy
}

func (w *WitTranscriber) Close() {
	w.client.CloseIdleConnections()
}

func splitAudioBytes(audioData []byte, sampleRate int, channels int, bytesPerSample int, chunkDurationSeconds int) ([][]byte, error) {
	if len(audioData) == 0 {
		return nil, fmt.Errorf("empty audio data")
	}
	if sampleRate <= 0 || channels <= 0 || bytesPerSample <= 0 || chunkDurationSeconds <= 0 {
		return nil, fmt.Errorf("invalid audio parameters")
	}

	bytesPerSecond := sampleRate * channels * bytesPerSample
	bytesPerChunk := bytesPerSecond * chunkDurationSeconds

	totalChunks := (len(audioData) + bytesPerChunk - 1) / bytesPerChunk
	chunks := make([][]byte, 0, totalChunks)

	for start := 0; start < len(audioData); start += bytesPerChunk {
		end := start + bytesPerChunk
		if end > len(audioData) {
			end = len(audioData)
		}

		chunk := make([]byte, end-start)
		copy(chunk, audioData[start:end])
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func preprocessAudio(r io.ReadSeeker) ([]byte, error) {
	decoder := wav.NewDecoder(r)

	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, err
	}

	outputSamples := len(buf.Data) * SampleRate / (buf.Format.SampleRate * buf.Format.NumChannels)

	outputData := make([]int16, outputSamples)

	for i := 0; i < len(buf.Data); i += buf.Format.NumChannels {
		var sample int
		if buf.Format.NumChannels > 1 {
			sample = 0
			for ch := 0; ch < buf.Format.NumChannels; ch++ {
				sample += buf.Data[i+ch]
			}
			sample /= buf.Format.NumChannels
		} else {
			sample = buf.Data[i]
		}

		targetIndex := i * SampleRate / (buf.Format.SampleRate * buf.Format.NumChannels)
		if targetIndex < len(outputData) {
			outputData[targetIndex] = int16(sample)
		}
	}

	outputBytes := make([]byte, len(outputData)*2)
	for i, sample := range outputData {
		binary.LittleEndian.PutUint16(outputBytes[i*2:], uint16(sample))
	}

	return outputBytes, nil
}

func witAITranscribe(r io.ReadSeeker, apiKey string) (chan string, error) {
	textChan := make(chan string)

	go func() {
		defer close(textChan)

		processedAudio, err := preprocessAudio(r)
		if err != nil {
			slog.Error("SPEECH | Error preprocessing audio", "error", err)
			return
		}

		chunks, _ := splitAudioBytes(processedAudio, SampleRate, Channels, BytesPerSample, ChunkDurationSeconds)

		transcriber := NewWitTranscriber(apiKey)
		for _, chunk := range chunks {
			text := transcriber.Transcribe(chunk)
			if text != "" {
				textChan <- text
			}
		}
		transcriber.Close()
	}()

	return textChan, nil
}

func WitAITranscribeFromFile(file string, apiKey string) (chan string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return witAITranscribe(f, apiKey)
}

func WitAITranscribeFromBuffer(buffer []byte, apiKey string) (string, error) {
	r := bytes.NewReader(buffer)

	resultChan, err := witAITranscribe(r, apiKey)
	if err != nil {
		return "", err
	}

	return <-resultChan, nil
}
