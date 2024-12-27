package local_whisper

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/go-audio/wav"
	whisper "github.com/kardianos/whisper.cpp/stt"

	whisper_model "myscript/internal/transcribe/whisper"
)

type LocalWhisperTranscriber struct {
	model        whisper.Model
	transcribing bool
	closing      bool
	mutex        sync.Mutex
}

func NewLocalWhisperTranscriber() *LocalWhisperTranscriber {
	return &LocalWhisperTranscriber{}
}

// If the language is English, we use the English-only model if it is already downloaded
// Otherwise, we use the multilingual model
func (l *LocalWhisperTranscriber) getBestModelPath(modelName string, language string) (string, error) {
	whisperModel, err := whisper_model.GetWhisperModel(modelName)
	if err != nil {
		return "", err
	}

	localWhisperModel := LocalWhisperModel{
		Name:        modelName,
		EnglishOnly: whisperModel.HasAlsoAnEnglishOnlyModel && language == whisper_model.ENGLISH_LANG_CODE,
	}

	if ModelExists(localWhisperModel) {
		return getModelPath(localWhisperModel)
	}

	localWhisperModel.EnglishOnly = !localWhisperModel.EnglishOnly

	if ModelExists(localWhisperModel) {
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

	fmt.Printf("Loaded local whisper model: %s\n", modelPath)

	return nil
}

func (l *LocalWhisperTranscriber) Transcribe(buffer []byte, language string) (string, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

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

	fmt.Printf("Local transcribing with language: %s\n", language)

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
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.closing = true

	if l.model != nil && l.canClose() {
		fmt.Printf("Unloading local whisper model\n")
		l.model.Close()
		l.model = nil
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
