import { create } from "zustand";
import { GetSecret, SaveSecret } from "~wails/main/App";

export type SecretKey = "notion";

type SecretsStore = {
  values: Partial<Record<SecretKey, string>>;
  load: (key: SecretKey) => Promise<string>;
  save: (key: SecretKey, value: string) => Promise<void>;
};

export const useSecretsStore = create<SecretsStore>((set, get) => ({
  values: {},

  load: async (key) => {
    const value = await GetSecret(key);
    set({ values: { ...get().values, [key]: value } });
    return value;
  },

  save: async (key, value) => {
    await SaveSecret(key, value);
    set({ values: { ...get().values, [key]: value } });
  },
}));
