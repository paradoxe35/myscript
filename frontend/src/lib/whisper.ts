export const LOCAL_WHISPER_MODELS = [
  {
    name: "tiny",
    englishOnly: true,
  },

  {
    name: "base",
    englishOnly: true,
  },

  {
    name: "small",
    englishOnly: true,
  },

  {
    name: "medium",
    englishOnly: true,
  },

  {
    name: "large",
    englishOnly: false,
    versions: ["v1", "v2", "v3"],
  },

  {
    name: "turbo",
    englishOnly: false,
  },
] as const;
