package local_whisper

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	whisper_model "myscript/internal/transcribe/whisper"

	"github.com/go-audio/wav"
	whisper "github.com/kardianos/whisper.cpp/stt"
)

// Config holds configuration options for LocalWhisperTranscriber
type Config struct {
	CloseTimeout time.Duration
	BufferSize   int
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		CloseTimeout: 10 * time.Minute,
		BufferSize:   44, // Minimum WAV header size
	}
}

// LocalWhisperTranscriber handles local speech-to-text transcription using Whisper
type LocalWhisperTranscriber struct {
	model        whisper.Model
	transcribing bool
	closing      bool
	mutex        sync.RWMutex
	config       *Config
}

// NewLocalWhisperTranscriber creates a new transcriber with the given configuration
func NewLocalWhisperTranscriber(config *Config) *LocalWhisperTranscriber {
	if config == nil {
		config = DefaultConfig()
	}
	return &LocalWhisperTranscriber{
		config: config,
	}
}

// getBestModelPath determines the most appropriate model path based on language
func (l *LocalWhisperTranscriber) getBestModelPath(modelName, language string) (string, error) {
	if modelName == "" || language == "" {
		return "", fmt.Errorf("modelName and language cannot be empty")
	}

	whisperModel, err := whisper_model.GetWhisperModel(modelName)
	if err != nil {
		return "", fmt.Errorf("failed to get whisper model: %w", err)
	}

	localWhisperModel := LocalWhisperModel{
		Name:        modelName,
		EnglishOnly: whisperModel.HasAlsoAnEnglishOnlyModel && language == whisper_model.ENGLISH_LANG_CODE,
	}

	// Try preferred model first
	if ModelExists(localWhisperModel) {
		return getModelPath(localWhisperModel)
	}

	// Fall back to alternative model
	localWhisperModel.EnglishOnly = !localWhisperModel.EnglishOnly
	if ModelExists(localWhisperModel) {
		return getModelPath(localWhisperModel)
	}

	return "", fmt.Errorf("no model found for %s (language: %s)", modelName, language)
}

// LoadModel loads the specified model for transcription
func (l *LocalWhisperTranscriber) LoadModel(modelName, language string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.transcribing && l.model != nil {
		return fmt.Errorf("transcription in progress; please wait")
	}

	if err := l.unloadCurrentModel(); err != nil {
		return fmt.Errorf("failed to unload current model: %w", err)
	}

	modelPath, err := l.getBestModelPath(modelName, language)
	if err != nil {
		return fmt.Errorf("failed to get model path: %w", err)
	}

	model, err := whisper.New(modelPath)
	if err != nil {
		return fmt.Errorf("failed to create new whisper model: %w", err)
	}

	l.model = model
	return nil
}

// Transcribe performs the audio transcription
func (l *LocalWhisperTranscriber) Transcribe(buffer []byte, language string) (string, error) {
	if err := l.validateTranscribeInput(buffer, language); err != nil {
		return "", err
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.model == nil {
		return "", fmt.Errorf("no model loaded")
	}

	l.setTranscribing(true)
	defer l.setTranscribing(false)

	return l.performTranscription(buffer, language)
}

func (l *LocalWhisperTranscriber) performTranscription(buffer []byte, language string) (string, error) {
	context, err := l.model.NewContext()
	if err != nil {
		return "", fmt.Errorf("failed to create context: %w", err)
	}

	context.SetLanguage(language)
	context.ResetTimings()

	samples, err := bytesToFloat32Buffer(buffer, l.config.BufferSize)
	if err != nil {
		return "", fmt.Errorf("failed to convert audio buffer: %w", err)
	}

	if err := context.Process(samples, nil); err != nil {
		return "", fmt.Errorf("failed to process audio: %w", err)
	}

	return l.collectTranscription(context)
}

func (l *LocalWhisperTranscriber) collectTranscription(context whisper.Context) (string, error) {
	var text string
	for {
		segment, err := context.NextSegment()
		if err != nil {
			break
		}
		text += segment.Text + " "
	}
	return text, nil
}

// Close safely closes the transcriber and releases resources
func (l *LocalWhisperTranscriber) Close() error {
	l.mutex.Lock()
	l.closing = true
	l.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), l.config.CloseTimeout)
	defer cancel()

	if err := l.waitForTranscriptionCompletion(ctx); err != nil {
		return fmt.Errorf("failed to close transcriber: %w", err)
	}

	return l.unloadCurrentModel()
}

func (l *LocalWhisperTranscriber) waitForTranscriptionCompletion(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for transcription to complete")
		case <-ticker.C:
			l.mutex.RLock()
			if !l.transcribing {
				l.mutex.RUnlock()
				return nil
			}
			l.mutex.RUnlock()
		}
	}
}

func (l *LocalWhisperTranscriber) unloadCurrentModel() error {
	if l.model != nil {
		l.model.Close()
		l.model = nil
	}
	return nil
}

func (l *LocalWhisperTranscriber) setTranscribing(state bool) {
	l.mutex.Lock()
	l.transcribing = state
	l.mutex.Unlock()
}

func (l *LocalWhisperTranscriber) validateTranscribeInput(buffer []byte, language string) error {
	if len(buffer) == 0 {
		return fmt.Errorf("empty audio buffer")
	}
	if language == "" {
		return fmt.Errorf("language cannot be empty")
	}
	return nil
}

// bytesToFloat32Buffer converts WAV bytes to float32 samples
func bytesToFloat32Buffer(b []byte, minSize int) ([]float32, error) {
	if len(b) < minSize {
		return nil, fmt.Errorf("invalid WAV file: too small (%d bytes)", len(b))
	}

	fh := bytes.NewReader(b)
	dec := wav.NewDecoder(fh)
	buf, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to decode WAV file: %w", err)
	}

	if dec.SampleRate != whisper.SampleRate {
		return nil, fmt.Errorf("unsupported sample rate: %d (expected %d)", dec.SampleRate, whisper.SampleRate)
	}

	if dec.NumChans != 1 {
		return nil, fmt.Errorf("expected mono audio, got %d channels", dec.NumChans)
	}

	return buf.AsFloat32Buffer().Data, nil
}
