import { create } from "zustand";
import {
  GetSpeechServiceAPIKey,
  GetSpeechServices,
  SaveSpeechServiceAPIKey,
} from "~wails/main/App";
import { main } from "~wails/models";

export type SpeechService = main.SpeechService;

type SpeechServicesStore = {
  services: SpeechService[];
  fetch: () => Promise<SpeechService[]>;
  apiKey: (service: string) => Promise<string>;
  saveAPIKey: (service: string, apiKey: string) => Promise<void>;
};

export const useSpeechServicesStore = create<SpeechServicesStore>((set) => ({
  services: [],

  fetch: async () => {
    const services = (await GetSpeechServices()) || [];
    set({ services });
    return services;
  },

  apiKey: (service) => GetSpeechServiceAPIKey(service),
  saveAPIKey: (service, apiKey) => SaveSpeechServiceAPIKey(service, apiKey),
}));
