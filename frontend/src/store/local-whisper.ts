import { create } from "zustand";
import { EventsOff, EventsOn } from "~wails-runtime";
import {
  AreSomeLocalWhisperModelsDownloading,
  DownloadLocalWhisperModels,
  ExistsLocalWhisperModel,
  GetBestLocalWhisperModel,
  GetLocalWhisperModels,
  IsLocalWhisperModelDownloading,
} from "~wails/main/App";
import { local_whisper, whisper } from "~wails/models";

type LocalWhisperState = {
  models: Array<whisper.WhisperModel>;
  bestModel: string | undefined;

  getModels: () => void;
  getBestModel: () => void;

  existsModel: (model: local_whisper.LocalWhisperModel) => Promise<boolean>;

  downloadModels: (models: local_whisper.LocalWhisperModel[]) => Promise<void>;

  newModelObject: (
    name: string,
    englishOnly: boolean
  ) => local_whisper.LocalWhisperModel;

  isModelDownloading: (
    model: local_whisper.LocalWhisperModel
  ) => Promise<boolean>;
  areSomeModelsDownloading: () => Promise<boolean>;

  onWhisperModelDownloadProgress: (
    callback: (progress: local_whisper.DownloadProgress) => void
  ) => void;
  onWhisperModelDownloadError: (callback: (error: string) => void) => void;
};

const ON_WHISPER_MODEL_DOWNLOAD_PROGRESS = "on-whisper-model-download-progress";
const ON_WHISPER_MODEL_DOWNLOAD_ERROR = "on-whisper-model-download-error";

export const useLocalWhisperStore = create<LocalWhisperState>((set) => ({
  models: [],
  bestModel: undefined,

  getModels: () => {
    return GetLocalWhisperModels().then((models) => {
      set({ models });
    });
  },

  getBestModel: () => {
    return GetBestLocalWhisperModel().then((bestModel) => {
      set({ bestModel });
    });
  },

  downloadModels(models) {
    return DownloadLocalWhisperModels(models);
  },

  existsModel(model) {
    return ExistsLocalWhisperModel(model);
  },

  isModelDownloading(model) {
    return IsLocalWhisperModelDownloading(model);
  },

  areSomeModelsDownloading() {
    return AreSomeLocalWhisperModelsDownloading();
  },

  newModelObject(name, englishOnly) {
    return new local_whisper.LocalWhisperModel({
      Name: name,
      EnglishOnly: englishOnly,
    });
  },

  onWhisperModelDownloadProgress: (callback) => {
    EventsOff(ON_WHISPER_MODEL_DOWNLOAD_PROGRESS);
    EventsOn(ON_WHISPER_MODEL_DOWNLOAD_PROGRESS, callback);

    return () => {
      EventsOff(ON_WHISPER_MODEL_DOWNLOAD_PROGRESS);
    };
  },

  onWhisperModelDownloadError: (callback) => {
    EventsOff(ON_WHISPER_MODEL_DOWNLOAD_ERROR);
    EventsOn(ON_WHISPER_MODEL_DOWNLOAD_ERROR, callback);

    return () => {
      EventsOff(ON_WHISPER_MODEL_DOWNLOAD_ERROR);
    };
  },
}));
