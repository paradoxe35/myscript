package openai

import (
	"bytes"
	"context"
	"myscript/internal/transcribe/whisper"

	"github.com/openai/openai-go" // imported as openai
	"github.com/openai/openai-go/option"
)

const WHISPER_MODEL = "whisper-1"

func validateLanguage(language string) error {
	valid := false
	for _, lang := range whisper.GetWhisperLanguages() {
		if lang.Code == language {
			valid = true
			break
		}
	}

	if !valid {
		return whisper.ErrInvalidLanguage
	}

	return nil
}

func TranscribeFromBuffer(buffer []byte, language, apiKey string) (string, error) {
	// It should have a valid language
	err := validateLanguage(language)
	if err != nil {
		return "", err
	}

	r := bytes.NewReader(buffer)

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithMaxRetries(2),
	)

	ctx := context.Background()

	reqFileName := "transcription.wav"
	reqContentType := "audio/wav"

	res, err := client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		File:           openai.FileParam(r, reqFileName, reqContentType),
		Model:          openai.String(WHISPER_MODEL),
		Language:       openai.String(language),
		ResponseFormat: openai.F(openai.AudioResponseFormatText),
	})

	if err != nil {
		return "", err
	}

	return res.Text, nil
}
