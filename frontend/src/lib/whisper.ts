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
    version: "v3",
  },

  {
    name: "large-turbo",
    englishOnly: false,
    version: "v3",
  },
] as const;
