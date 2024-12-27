package whisper

import (
	"errors"
	"myscript/internal/transcribe/structs"
)

type WhisperModel struct {
	Name                      string
	HasAlsoAnEnglishOnlyModel bool
	RAMRequired               float64 // in GB
	Enabled                   bool
}

var ErrInvalidLanguage = errors.New("invalid language")
var ErrInvalidModelName = errors.New("invalid model name")

var localWhisperModels = []WhisperModel{
	{"tiny", true, 6, true},   // ~1GB VRAM -> 6GB RAM
	{"base", true, 6, true},   // ~1GB VRAM -> 6GB RAM
	{"small", true, 12, true}, // ~2GB VRAM -> 12GB RAM
	// Disabled models
	{"medium", true, 30, false},       // ~5GB VRAM -> 30GB RAM
	{"large", false, 60, false},       // ~10GB VRAM -> 60GB RAM
	{"large-turbo", false, 36, false}, // ~6GB VRAM -> 36GB RAM
}

var LANGUAGES = []structs.Language{
	{Code: "en", Name: "English"},
	{Code: "ar", Name: "Arabic"},
	{Code: "hy", Name: "Armenian"},
	{Code: "az", Name: "Azerbaijani"},
	{Code: "eu", Name: "Basque"},
	{Code: "be", Name: "Belarusian"},
	{Code: "bn", Name: "Bengali"},
	{Code: "bg", Name: "Bulgarian"},
	{Code: "ca", Name: "Catalan"},
	{Code: "zh", Name: "Chinese"},
	{Code: "hr", Name: "Croatian"},
	{Code: "cs", Name: "Czech"},
	{Code: "da", Name: "Danish"},
	{Code: "nl", Name: "Dutch"},
	{Code: "et", Name: "Estonian"},
	{Code: "tl", Name: "Filipino"},
	{Code: "fi", Name: "Finnish"},
	{Code: "fr", Name: "French"},
	{Code: "gl", Name: "Galician"},
	{Code: "ka", Name: "Georgian"},
	{Code: "de", Name: "German"},
	{Code: "el", Name: "Greek"},
	{Code: "gu", Name: "Gujarati"},
	{Code: "iw", Name: "Hebrew"},
	{Code: "hi", Name: "Hindi"},
	{Code: "hu", Name: "Hungarian"},
	{Code: "is", Name: "Icelandic"},
	{Code: "id", Name: "Indonesian"},
	{Code: "ga", Name: "Irish"},
	{Code: "it", Name: "Italian"},
	{Code: "ja", Name: "Japanese"},
	{Code: "kn", Name: "Kannada"},
	{Code: "ko", Name: "Korean"},
	{Code: "la", Name: "Latin"},
	{Code: "lv", Name: "Latvian"},
	{Code: "lt", Name: "Lithuanian"},
	{Code: "mk", Name: "Macedonian"},
	{Code: "ms", Name: "Malay"},
	{Code: "mt", Name: "Maltese"},
	{Code: "no", Name: "Norwegian"},
	{Code: "fa", Name: "Persian"},
	{Code: "pl", Name: "Polish"},
	{Code: "pt", Name: "Portuguese"},
	{Code: "ro", Name: "Romanian"},
	{Code: "ru", Name: "Russian"},
	{Code: "sr", Name: "Serbian"},
	{Code: "sk", Name: "Slovak"},
	{Code: "sl", Name: "Slovenian"},
	{Code: "es", Name: "Spanish"},
	{Code: "sw", Name: "Swahili"},
	{Code: "sv", Name: "Swedish"},
	{Code: "ta", Name: "Tamil"},
	{Code: "te", Name: "Telugu"},
	{Code: "th", Name: "Thai"},
	{Code: "tr", Name: "Turkish"},
	{Code: "uk", Name: "Ukrainian"},
	{Code: "ur", Name: "Urdu"},
	{Code: "vi", Name: "Vietnamese"},
	{Code: "cy", Name: "Welsh"},
	{Code: "yi", Name: "Yiddish"},
}

// Get whisper from the predefined models
//
// The modelName value should be only name of the model without the extension
// e.g. tiny, base, small, medium, large, turbo
func GetWhisperModel(modelName string) (WhisperModel, error) {
	for _, model := range localWhisperModels {
		if model.Name == modelName {
			return model, nil
		}
	}

	return WhisperModel{}, ErrInvalidModelName
}

func GetWhisperLanguages() []structs.Language {
	return LANGUAGES
}

func GetWhisperModels() (models []WhisperModel) {
	for _, model := range localWhisperModels {
		if model.Enabled {
			models = append(models, model)
		}
	}

	return models
}

func SuggestWhisperModel(availableRAM float64) string {
	var bestModel string
	var highestRAMUsage float64

	for _, model := range localWhisperModels {
		if availableRAM >= model.RAMRequired && model.RAMRequired >= highestRAMUsage {
			highestRAMUsage = model.RAMRequired
			bestModel = model.Name
		}
	}

	if bestModel == "" {
		return "tiny" // Default to the smallest model if no suitable model is found
	}

	return bestModel
}
