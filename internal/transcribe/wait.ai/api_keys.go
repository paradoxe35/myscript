// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package witai

import "myscript/internal/transcribe/structs"

type ApiKey struct {
	Key      string
	Language string
}

var LANGUAGES = []structs.Language{
	{Name: "English", Code: "en"},
	{Name: "French", Code: "fr"},
	{Name: "Arabic", Code: "ar"},
	{Name: "Bengali", Code: "bn"},
	{Name: "Burmese", Code: "my"},
	{Name: "Chinese", Code: "zh"},
	{Name: "Dutch", Code: "nl"},
	{Name: "Finnish", Code: "fi"},
	{Name: "German", Code: "de"},
	{Name: "Hindi", Code: "hi"},
	{Name: "Indonesian", Code: "id"},
	{Name: "Italian", Code: "it"},
	{Name: "Japanese", Code: "ja"},
	{Name: "Kannada", Code: "kn"},
	{Name: "Korean", Code: "ko"},
	{Name: "Malay", Code: "ms"},
	{Name: "Malayalam", Code: "ml"},
	{Name: "Marathi", Code: "mr"},
	{Name: "Polish", Code: "pl"},
	{Name: "Portuguese", Code: "pt"},
	{Name: "Russian", Code: "ru"},
	{Name: "Sinhalese", Code: "si"},
	{Name: "Spanish", Code: "es"},
	{Name: "Swedish", Code: "sv"},
	{Name: "Tamil", Code: "ta"},
	{Name: "Telugu", Code: "te"},
	{Name: "Thai", Code: "th"},
	{Name: "Turkish", Code: "tr"},
	{Name: "Urdu", Code: "ur"},
	{Name: "Vietnamese", Code: "vi"},
}

func GetSupportedLanguages() []structs.Language {
	var languages []structs.Language

	for _, lang := range LANGUAGES {
		for _, key := range API_KEYS {
			if key.Language == lang.Code {
				languages = append(languages, lang)
				continue
			}
		}
	}

	return languages
}

// Create keys.go where you list all the API keys
// for the different languages
// Example:
//
//	var API_KEYS = []ApiKey{
//		{"XXXXXXXXXXXXXXXXXXX", "en"},
//	}
func GetAPIKey(lan string) *ApiKey {
	for _, key := range API_KEYS {
		if key.Language == lan {
			return &key
		}
	}

	return nil
}
