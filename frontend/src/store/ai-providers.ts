import { create } from "zustand";
import {
  DeleteAIProvider,
  HasConfiguredAIProvider,
  GetAIProviderAPIKey,
  GetAIProviders,
  GetActiveAIProvider,
  ListAIModels,
  SaveAIProvider,
  SetActiveAIProvider,
  TestAIProvider,
} from "~wails/main/App";
import { ai, main } from "~wails/models";

export type AIProvider = main.AIProvider;
export type AIModel = ai.ModelInfo;

export const OPENAI = "openai";
export const ANTHROPIC = "anthropic";
export const GEMINI = "gemini";

export const PROVIDER_LABELS: Record<string, string> = {
  [OPENAI]: "OpenAI",
  [ANTHROPIC]: "Claude",
  [GEMINI]: "Gemini",
};

export function providerLabel(provider: Pick<AIProvider, "Name">) {
  return PROVIDER_LABELS[provider.Name] ?? provider.Name;
}

type AIProvidersStore = {
  providers: AIProvider[];
  active: string;
  loading: boolean;
  enabled: boolean;

  fetch: () => Promise<AIProvider[]>;
  checkEnabled: () => Promise<boolean>;
  save: (provider: AIProvider, apiKey: string) => Promise<void>;
  remove: (name: string) => Promise<void>;
  activate: (name: string) => Promise<void>;
  apiKey: (name: string) => Promise<string>;
  listModels: (provider: AIProvider, apiKey: string) => Promise<AIModel[]>;
  test: (provider: AIProvider, apiKey: string) => Promise<void>;
};

export const useAIProvidersStore = create<AIProvidersStore>((set, get) => ({
  providers: [],
  active: OPENAI,
  loading: false,
  enabled: false,

  fetch: async () => {
    set({ loading: true });
    try {
      const [providers, active] = await Promise.all([
        GetAIProviders(),
        GetActiveAIProvider(),
      ]);
      set({ providers: providers || [], active });
      get().checkEnabled();
      return providers || [];
    } finally {
      set({ loading: false });
    }
  },

  checkEnabled: async () => {
    const enabled = await HasConfiguredAIProvider();
    set({ enabled });
    return enabled;
  },

  save: async (provider, apiKey) => {
    await SaveAIProvider(provider, apiKey);
    await get().fetch();
  },

  remove: async (name) => {
    await DeleteAIProvider(name);
    await get().fetch();
  },

  activate: async (name) => {
    await SetActiveAIProvider(name);
    set({ active: name });
    await get().fetch();
  },

  apiKey: (name) => GetAIProviderAPIKey(name),

  listModels: async (provider, apiKey) => {
    const models = await ListAIModels(provider, apiKey);
    return models || [];
  },

  test: (provider, apiKey) => TestAIProvider(provider, apiKey),
}));
