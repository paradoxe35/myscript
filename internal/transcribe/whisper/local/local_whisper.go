package local_whisper

import (
	"encoding/binary"
	"fmt"
	"math"

	whisper "github.com/kardianos/whisper.cpp/stt"

	whisper_model "myscript/internal/transcribe/whisper"
)

type LocalWhisperTranscriber struct {
	model whisper.Model
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
	if l.model == nil {
		return "", fmt.Errorf("no model loaded")
	}

	// Create processing context
	context, err := l.model.NewContext()
	if err != nil {
		return "", err
	}

	context.SetLanguage(language)
	context.ResetTimings()

	fmt.Printf("Local transcribing with language: %s\n", language)

	if err := context.Process(GetFloatArray(buffer), nil); err != nil {
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
	if l.model != nil {
		l.model.Close()
		l.model = nil
	}

	return nil
}

func GetFloatArray(aBytes []byte) []float32 {
	aArr := make([]float32, 3)
	for i := 0; i < 3; i++ {
		aArr[i] = BytesFloat32(aBytes[i*4:])
	}
	return aArr
}

func BytesFloat32(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}
