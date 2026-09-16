// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

// Package languages is the one list of spoken languages shared by every
// transcriber, local or remote.
package languages

import (
	"errors"
	"sort"
	"strings"
)

type Language struct {
	Name string
	Code string
}

var ErrInvalid = errors.New("invalid language")

// Whisper is what the OpenAI and Groq Whisper endpoints accept, as ISO 639-1.
// Filipino stays "tl": Whisper's own vocabulary predates "fil".
var Whisper = Named([]string{
	"en", "ar", "hy", "az", "eu", "be", "bn", "bg", "ca", "zh", "hr", "cs", "da", "nl", "et",
	"tl", "fi", "fr", "gl", "ka", "de", "el", "gu", "he", "hi", "hu", "is", "id", "ga", "it",
	"ja", "kn", "ko", "la", "lv", "lt", "mk", "ms", "mt", "no", "fa", "pl", "pt", "ro", "ru",
	"sr", "sk", "sl", "es", "sw", "sv", "ta", "te", "th", "tr", "uk", "ur", "vi", "cy", "yi",
})

// Validate accepts only what the Whisper endpoints take.
func Validate(code string) error {
	for _, language := range Whisper {
		if language.Code == code {
			return nil
		}
	}
	return ErrInvalid
}

// Name resolves a code to its English name, or the code itself when unknown.
func Name(code string) string {
	if name, ok := names[strings.ToLower(code)]; ok {
		return name
	}
	return code
}

// Named turns model language codes into a list sorted by name, English first.
func Named(codes []string) []Language {
	languages := make([]Language, 0, len(codes))
	for _, code := range codes {
		languages = append(languages, Language{Code: code, Name: Name(code)})
	}
	sort.SliceStable(languages, func(i, j int) bool {
		if (languages[i].Code == "en") != (languages[j].Code == "en") {
			return languages[i].Code == "en"
		}
		return languages[i].Name < languages[j].Name
	})
	return languages
}

// names covers the codes the model catalogue uses; unknown codes display as-is.
var names = map[string]string{
	"en": "English", "fr": "French", "es": "Spanish", "de": "German",
	"it": "Italian", "pt": "Portuguese", "nl": "Dutch", "pl": "Polish",
	"ru": "Russian", "uk": "Ukrainian", "cs": "Czech", "sk": "Slovak",
	"hu": "Hungarian", "ro": "Romanian", "bg": "Bulgarian", "el": "Greek",
	"tr": "Turkish", "ar": "Arabic", "he": "Hebrew", "iw": "Hebrew", "fa": "Persian",
	"ur": "Urdu", "hi": "Hindi", "bn": "Bengali", "ta": "Tamil",
	"te": "Telugu", "mr": "Marathi", "gu": "Gujarati", "pa": "Punjabi",
	"th": "Thai", "vi": "Vietnamese", "id": "Indonesian", "ms": "Malay",
	"tl": "Filipino", "fil": "Filipino", "zh": "Chinese", "yue": "Cantonese",
	"ja": "Japanese", "ko": "Korean",
	"sv": "Swedish", "da": "Danish", "no": "Norwegian", "nb": "Norwegian",
	"nn": "Norwegian Nynorsk", "fi": "Finnish",
	"is": "Icelandic", "ca": "Catalan", "eu": "Basque", "gl": "Galician",
	"cy": "Welsh", "ga": "Irish", "af": "Afrikaans", "sw": "Swahili",
	"am": "Amharic", "ha": "Hausa", "yo": "Yoruba", "ig": "Igbo",
	"zu": "Zulu", "rw": "Kinyarwanda", "ln": "Lingala", "mg": "Malagasy",
	"hy": "Armenian", "ka": "Georgian", "az": "Azerbaijani", "kk": "Kazakh",
	"uz": "Uzbek", "mn": "Mongolian", "ne": "Nepali", "si": "Sinhala",
	"km": "Khmer", "lo": "Lao", "my": "Burmese", "sq": "Albanian",
	"sr": "Serbian", "hr": "Croatian", "sl": "Slovenian", "mk": "Macedonian",
	"bs": "Bosnian", "et": "Estonian", "lv": "Latvian", "lt": "Lithuanian",
	"ml": "Malayalam", "kn": "Kannada", "or": "Odia", "so": "Somali",
	"om": "Oromo", "jv": "Javanese", "jw": "Javanese", "su": "Sundanese",
	"eo": "Esperanto",
	"as": "Assamese", "ba": "Bashkir", "be": "Belarusian", "bo": "Tibetan",
	"br": "Breton", "fo": "Faroese", "haw": "Hawaiian", "ht": "Haitian Creole",
	"la": "Latin", "lb": "Luxembourgish", "mi": "Maori", "mt": "Maltese",
	"oc": "Occitan", "ps": "Pashto", "sa": "Sanskrit", "sd": "Sindhi",
	"sn": "Shona", "tg": "Tajik", "tk": "Turkmen", "tt": "Tatar",
	"yi": "Yiddish",
}
