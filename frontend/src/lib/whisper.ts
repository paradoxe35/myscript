export const LOCAL_WHISPER_MODELS = [
  {
    name: "tiny",
    englishOnly: true,
    requiredRAM: 6,
  },

  {
    name: "base",
    englishOnly: true,
    requiredRAM: 6,
  },

  {
    name: "small",
    englishOnly: true,
    requiredRAM: 12,
  },

  {
    name: "medium",
    englishOnly: true,
    requiredRAM: 30,
  },

  {
    name: "large",
    englishOnly: false,
    version: "v3",
    requiredRAM: 60,
  },

  {
    name: "large-turbo",
    englishOnly: false,
    version: "v3",
    requiredRAM: 36,
  },
] as const;
