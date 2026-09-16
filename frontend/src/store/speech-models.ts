import { EventClear } from "@/types";
import { create } from "zustand";
import { EventsOn } from "~wails-runtime";
import {
  CancelSpeechModelDownload,
  DeleteSpeechModel,
  DownloadSpeechModel,
  GetMachine,
  GetSpeechModels,
  HasLegacyWhisperFiles,
  RemoveLegacyWhisperFiles,
} from "~wails/main/App";
import { main } from "~wails/models";

export type ModelDownloadEvent = {
  ID: string;
  Name: string;
  Downloaded: number;
  Total: number;
  Stage: "downloading" | "verifying" | "done" | "";
  Error: string;
};

type SpeechModelsStore = {
  models: Array<main.SpeechModel>;
  machine: main.MachineInfo | null;
  progress: Record<string, ModelDownloadEvent>;

  fetchModels: () => Promise<Array<main.SpeechModel>>;
  fetchMachine: () => Promise<void>;
  downloadModel: (id: string) => Promise<void>;
  cancelDownload: (id: string) => Promise<void>;
  deleteModel: (id: string) => Promise<void>;
  setProgress: (event: ModelDownloadEvent) => void;
  clearProgress: (id: string) => void;

  downloaded: () => Array<main.SpeechModel>;

  hasLegacyFiles: boolean;
  checkLegacyFiles: () => Promise<void>;
  removeLegacyFiles: () => Promise<number>;

  onDownloadProgress: (
    callback: (event: ModelDownloadEvent) => void
  ) => EventClear;
  onDownloadDone: (callback: (event: ModelDownloadEvent) => void) => EventClear;
  onDownloadError: (callback: (event: ModelDownloadEvent) => void) => EventClear;
  onDownloadCancelled: (
    callback: (event: ModelDownloadEvent) => void
  ) => EventClear;
};

const ON_MODEL_DOWNLOAD_PROGRESS = "on-model-download-progress";
const ON_MODEL_DOWNLOAD_DONE = "on-model-download-done";
const ON_MODEL_DOWNLOAD_ERROR = "on-model-download-error";
const ON_MODEL_DOWNLOAD_CANCELLED = "on-model-download-cancelled";

export const useSpeechModelsStore = create<SpeechModelsStore>((set, get) => ({
  models: [],
  machine: null,
  progress: {},

  fetchModels: async () => {
    const models = (await GetSpeechModels()) || [];
    set({ models });
    return models;
  },

  fetchMachine: async () => {
    set({ machine: await GetMachine() });
  },

  downloadModel: async (id) => {
    await DownloadSpeechModel(id);
    set((state) => ({
      models: state.models.map((model) =>
        model.ID === id
          ? main.SpeechModel.createFrom({ ...model, Downloading: true })
          : model
      ),
    }));
  },

  cancelDownload: async (id) => {
    await CancelSpeechModelDownload(id);
  },

  deleteModel: async (id) => {
    await DeleteSpeechModel(id);
    await get().fetchModels();
  },

  setProgress: (event) => {
    set((state) => ({ progress: { ...state.progress, [event.ID]: event } }));
  },

  clearProgress: (id) => {
    set((state) => {
      const { [id]: _, ...progress } = state.progress;
      return { progress };
    });
  },

  downloaded: () => get().models.filter((model) => model.Downloaded),

  hasLegacyFiles: false,

  checkLegacyFiles: async () => {
    set({ hasLegacyFiles: await HasLegacyWhisperFiles() });
  },

  removeLegacyFiles: async () => {
    const removed = await RemoveLegacyWhisperFiles();
    await get().checkLegacyFiles();
    return removed;
  },

  onDownloadProgress: (callback) => {
    return EventsOn(ON_MODEL_DOWNLOAD_PROGRESS, callback);
  },

  onDownloadDone: (callback) => {
    return EventsOn(ON_MODEL_DOWNLOAD_DONE, callback);
  },

  onDownloadError: (callback) => {
    return EventsOn(ON_MODEL_DOWNLOAD_ERROR, callback);
  },

  onDownloadCancelled: (callback) => {
    return EventsOn(ON_MODEL_DOWNLOAD_CANCELLED, callback);
  },
}));
