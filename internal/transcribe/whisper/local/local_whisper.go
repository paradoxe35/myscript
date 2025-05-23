// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package local_whisper

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-audio/wav"
	whisper "github.com/paradoxe35/whisper.cpp-go/stt"

	whisper_model "myscript/internal/transcribe/whisper"
)

type LocalWhisperTranscriber struct {
	model        whisper.Model
	transcribing bool
	closing      bool
	mu           sync.Mutex
}

func NewLocalWhisperTranscriber() *LocalWhisperTranscriber {
	return &LocalWhisperTranscriber{}
}

// If the language is English, we use the English-only model if it is already downloaded
// Otherwise, we use the multilingual model
func (l *LocalWhisperTranscriber) getBestModelPath(modelName string, language string) (string, error) {
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
		if localWhisperModel.EnglishOnly && language != whisper_model.ENGLISH_LANG_CODE {
			return "", fmt.Errorf("You are using an English-only model on a non-English script. Please consider downloading the multilingual model as well.")
		}

		return getModelPath(localWhisperModel)
	}

	return "", fmt.Errorf("no model found for %s. Please ensure you have downloaded it.", modelName)
}

func (l *LocalWhisperTranscriber) LoadModel(modelName string, language string) error {
	if l.transcribing && l.model != nil {
		return fmt.Errorf("another transcription is in progress; please wait.")
	}

	// Unload model if it is already loaded
	if l.model != nil {
		l.model.Close()
		l.model = nil
	}

	modelPath, err := l.getBestModelPath(modelName, language)
	if err != nil {
		return err
	}

	// Load model
	model, err := whisper.New(modelPath)
	if err != nil {
		return err
	}

	l.model = model

	slog.Debug("Loaded local whisper model", "model", modelPath)

	return nil
}

func (l *LocalWhisperTranscriber) Transcribe(buffer []byte, language string) (string, error) {
	if err := l.validateTranscribeInput(buffer, language); err != nil {
		return "", err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.model == nil {
		return "", fmt.Errorf("no model loaded")
	}

	l.transcribing = true
	l.closing = false
	defer func() {
		l.transcribing = false
	}()

	// Create processing context
	context, err := l.model.NewContext()
	if err != nil {
		return "", err
	}

	context.SetLanguage(language)
	context.ResetTimings()

	slog.Debug("Local transcribing with language", "language", language)

	samples, err := bytesToFloat32Buffer(buffer)
	if err != nil {
		return "", err
	}

	if err := context.Process(samples, nil); err != nil {
		return "", err
	}

	// Get the result
	text := ""

	// Print out the results
	for {
		segment, err := context.NextSegment()
		if err != nil {
			break
		}

		text += segment.Text + " "
	}

	return text, nil
}

func (l *LocalWhisperTranscriber) Close() error {
	l.mu.Lock()
	l.closing = true
	l.mu.Unlock()

	if l.model != nil && l.canClose() {
		slog.Debug("Unloading local whisper model")
		l.model.Close()
		l.model = nil
	}

	return nil
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

func (l *LocalWhisperTranscriber) canClose() bool {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !l.closing {
				return false
			}
			if !l.transcribing {
				return true
			}
		}
	}
}

func bytesToFloat32Buffer(b []byte) ([]float32, error) {
	// Validate minimum WAV file size
	if len(b) < 44 { // 44 bytes is the minimum size for a valid WAV header
		return nil, fmt.Errorf("invalid WAV file: too small (%d bytes)", len(b))
	}

	fh := bytes.NewReader(b)

	// Decode the WAV file - load the full buffer
	dec := wav.NewDecoder(fh)
	buf, err := dec.FullPCMBuffer()

	if err != nil {
		return nil, err
	}

	if dec.SampleRate != whisper.SampleRate {
		return nil, fmt.Errorf("unsupported sample rate: %d", dec.SampleRate)
	}

	if dec.NumChans != 1 {
		return nil, fmt.Errorf("expected mono audio, got %d channels", dec.NumChans)
	}

	return buf.AsFloat32Buffer().Data, nil

}
