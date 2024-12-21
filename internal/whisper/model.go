package whisper

type WhisperModel struct {
	Name          string
	RAMRequired   float64 // in GB
	RelativeSpeed float64
}

var LOCAL_WHISPER_MODELS = []WhisperModel{
	{"tiny", 1, 10},
	{"base", 1, 7},
	{"small", 2, 4},
	{"medium", 5, 2},
	{"large", 10, 1},
	{"turbo", 6, 8},
}

func SuggestWhisperModel(availableRAM float64) string {
	var bestModel string
	var bestScore float64

	for _, model := range LOCAL_WHISPER_MODELS {
		if availableRAM < model.RAMRequired {
			continue
		}

		// Calculate a score based on RAM efficiency and speed
		ramEfficiency := availableRAM / model.RAMRequired
		score := ramEfficiency * model.RelativeSpeed

		if score > bestScore {
			bestScore = score
			bestModel = model.Name
		}
	}

	if bestModel == "" {
		return "tiny" // Default to the smallest model if no suitable model is found
	}

	return bestModel
}
