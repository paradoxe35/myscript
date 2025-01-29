package groq

import (
	"bytes"
	"context"
	"myscript/internal/transcribe/whisper"

	"github.com/conneroisu/groq-go"
)

var (
	GROQ_TRANSCRIBE_MODEL = groq.ModelWhisperLargeV3Turbo
)

func GetGroqTranscribeModel() string {
	return string(GROQ_TRANSCRIBE_MODEL)
}

func TranscribeFromBuffer(buffer []byte, language, apiKey string) (string, error) {
	// Since it uses the whisper model, it should have a valid language
	err := whisper.ValidateWhisperLanguage(language)
	if err != nil {
		return "", err
	}

	client, err := groq.NewClient(apiKey)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	response, err := client.Transcribe(ctx, groq.AudioRequest{
		Model:    GROQ_TRANSCRIBE_MODEL,
		Language: language,
		FilePath: "stt.wav",
		Reader:   bytes.NewReader(buffer),
	})

	if err != nil {
		return "", err
	}

	return response.Text, nil
}
