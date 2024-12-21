package whisper

type WhisperModel struct {
	Name        string
	RAMRequired float64 // in GB
}

var LOCAL_WHISPER_MODELS = []WhisperModel{
	{"tiny", 6},    // ~1GB VRAM -> 6GB RAM
	{"base", 6},    // ~1GB VRAM -> 6GB RAM
	{"small", 12},  // ~2GB VRAM -> 12GB RAM
	{"medium", 30}, // ~5GB VRAM -> 30GB RAM
	{"large", 60},  // ~10GB VRAM -> 60GB RAM
	{"turbo", 36},  // ~6GB VRAM -> 36GB RAM
}

func SuggestWhisperModel(availableRAM float64) string {
	var bestModel string
	var highestRAMUsage float64

	for _, model := range LOCAL_WHISPER_MODELS {
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
