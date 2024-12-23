package witai

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

var logger = log.New(log.Writer(), "speech: ", log.LstdFlags)

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
		logger.Printf("Error creating request: %v", err)
		return ""
	}

	// Add headers
	for key, value := range w.headers {
		req.Header.Add(key, value)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("verbose", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := w.client.Do(req)
	if err != nil {
		logger.Printf("Error making request: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("Error reading response body: %v", err)
		return ""
	}

	logger.Printf("Request response %s", string(body))

	var witResp WitResponse
	if err := json.Unmarshal(body, &witResp); err != nil {
		logger.Printf("Error unmarshaling response: %v", err)
		return ""
	}

	// Check for text in both new and legacy fields
	if witResp.Text != "" {
		return witResp.Text
	}
	return witResp.Legacy
}

func (w *WitTranscriber) Close() {
	w.client.CloseIdleConnections()
}

func splitAudioBytes(audioData []byte, sampleRate int, channels int, bytesPerSample int, chunkDurationSeconds int) ([][]byte, error) {
	// Validate input parameters
	if len(audioData) == 0 {
		return nil, fmt.Errorf("empty audio data")
	}
	if sampleRate <= 0 || channels <= 0 || bytesPerSample <= 0 || chunkDurationSeconds <= 0 {
		return nil, fmt.Errorf("invalid audio parameters")
	}

	// Calculate bytes per chunk
	bytesPerSecond := sampleRate * channels * bytesPerSample
	bytesPerChunk := bytesPerSecond * chunkDurationSeconds

	// Calculate total number of chunks
	totalChunks := (len(audioData) + bytesPerChunk - 1) / bytesPerChunk // Round up division

	// Create slice to hold all chunks
	chunks := make([][]byte, 0, totalChunks)

	// Split audio data into chunks
	for start := 0; start < len(audioData); start += bytesPerChunk {
		end := start + bytesPerChunk
		if end > len(audioData) {
			end = len(audioData)
		}

		// Create a new slice for this chunk
		chunk := make([]byte, end-start)
		copy(chunk, audioData[start:end])
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func preprocessAudio(file string) ([]byte, error) {
	// Open the input file
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Create a new decoder
	decoder := wav.NewDecoder(f)

	// Read the full buffer
	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, err
	}

	// Calculate the number of samples in the output
	outputSamples := len(buf.Data) * SampleRate / (buf.Format.SampleRate * buf.Format.NumChannels)

	// Create a slice to hold the output samples
	outputData := make([]int16, outputSamples)

	// Resample and convert the audio
	for i := 0; i < len(buf.Data); i += buf.Format.NumChannels {
		var sample int
		if buf.Format.NumChannels > 1 {
			// Convert to mono by averaging channels
			sample = 0
			for ch := 0; ch < buf.Format.NumChannels; ch++ {
				sample += buf.Data[i+ch]
			}
			sample /= buf.Format.NumChannels
		} else {
			sample = buf.Data[i]
		}

		// Resample
		targetIndex := i * SampleRate / (buf.Format.SampleRate * buf.Format.NumChannels)
		if targetIndex < len(outputData) {
			// Convert to int16 and apply any necessary scaling
			outputData[targetIndex] = int16(sample)
		}
	}

	// Convert int16 slice to bytes
	outputBytes := make([]byte, len(outputData)*2)
	for i, sample := range outputData {
		binary.LittleEndian.PutUint16(outputBytes[i*2:], uint16(sample))
	}

	return outputBytes, nil
}

func WitAITranscribe(file string, apiKey string) (chan string, error) {
	textChan := make(chan string)

	go func() {
		defer close(textChan)

		processedAudio, err := preprocessAudio(file)
		if err != nil {
			logger.Printf("Error preprocessing audio: %v", err)
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
