import { create } from "zustand";
import {
  CancelRecording,
  GetCache,
  GetLanguages,
  GetMicInputDevices,
  IsRecording,
  SaveCache,
  StartRecording,
  StopRecording,
} from "~wails/main/App";
import { languages, stt } from "~wails/models";
import { EventsOn } from "~wails-runtime";
import { EventClear } from "@/types";

export type TranscriberState = "idle" | "loading" | "ready" | "listening";

export type TranscriberStateEvent = {
  State: TranscriberState;
  ModelName: string;
};

/** Offered when the model can detect the language; sent to Go as an empty code. */
export const AUTO_DETECT_LANGUAGE: languages.Language = {
  Code: "auto",
  Name: "Auto-detect",
};

type TranscriberStore = {
  state: TranscriberState;
  modelName: string;
  micLevel: number;
  isRecording: boolean;
  languages: Array<languages.Language>;
  micInputDevices: Array<stt.Device>;

  startRecording: (languageCode: string, micInputDevice: string) => Promise<void>;
  stopRecording: () => Promise<void>;
  cancelRecording: () => Promise<void>;
  setState: (event: TranscriberStateEvent) => void;
  setMicLevel: (level: number) => void;
  getRecordingStatus: () => Promise<boolean>;
  getMicInputDevices: () => Promise<Array<stt.Device>>;
  getLanguages: () => void;

  getPageLanguage: (pageId: string | number) => Promise<string | null>;
  setPageLanguage: (pageId: string | number, language: string) => Promise<void>;

  setDefaultMicInput: (micInputDevice: stt.Device) => void;
  getDefaultMicInput: () => Promise<stt.Device | undefined>;

  onTranscribedText: (callback: (text: string) => void) => EventClear;
  onTranscribeError: (callback: (error: string) => void) => EventClear;
  onRecordingStopped: (callback: (autoStopped: boolean) => void) => EventClear;
  onTranscriberState: (
    callback: (event: TranscriberStateEvent) => void
  ) => EventClear;
  onMicLevel: (callback: (level: number) => void) => EventClear;
};

// Pages remembered before the Whisper list moved to ISO codes.
const LEGACY_LANGUAGE_CODES: Record<string, string> = { iw: "he" };

const ON_TRANSCRIBED_TEXT = "on-transcribed-text";
const ON_TRANSCRIBE_ERROR = "on-transcribe-error";
const ON_RECORDING_STOPPED = "on-recording-stopped";
const ON_TRANSCRIBER_STATE = "on-transcriber-state";
const ON_MIC_LEVEL = "on-mic-level";

export const useTranscriberStore = create<TranscriberStore>((set, get) => ({
  state: "idle",
  modelName: "",
  micLevel: 0,
  isRecording: false,
  languages: [],
  micInputDevices: [],

  async startRecording(languageCode, micInputDevice) {
    if (get().isRecording) return;

    const code = languageCode === AUTO_DETECT_LANGUAGE.Code ? "" : languageCode;
    set({ state: "loading", isRecording: true });

    return StartRecording(code, micInputDevice).catch((err) => {
      set({ state: "idle", isRecording: false });
      throw err;
    });
  },

  async stopRecording() {
    if (!get().isRecording) return;

    return StopRecording().finally(() => {
      get().getRecordingStatus();
    });
  },

  async cancelRecording() {
    if (!get().isRecording) return;

    return CancelRecording().finally(() => {
      get().getRecordingStatus();
    });
  },

  setState(event) {
    set({
      state: event.State,
      modelName: event.ModelName,
      isRecording: event.State !== "idle",
      micLevel: event.State === "listening" ? get().micLevel : 0,
    });
  },

  setMicLevel(level) {
    set({ micLevel: level });
  },

  getRecordingStatus: () => {
    return IsRecording().then((isRecording) => {
      set({
        isRecording,
        state: isRecording ? get().state : "idle",
        micLevel: isRecording ? get().micLevel : 0,
      });
      return isRecording;
    });
  },

  getMicInputDevices: () => {
    return GetMicInputDevices().then((micInputDevices) => {
      set({ micInputDevices: micInputDevices || [] });
      return micInputDevices || [];
    });
  },

  getLanguages: () => {
    GetLanguages().then((languages) => {
      set({ languages: languages || [] });
    });
  },

  getPageLanguage: async (pageId) => {
    const language = await GetCache(`page-${pageId}-language`);
    return LEGACY_LANGUAGE_CODES[language?.value] ?? language?.value ?? null;
  },

  setPageLanguage: async (pageId, language) => {
    SaveCache(`page-${pageId}-language`, language);
  },

  setDefaultMicInput: (micInputDevice) => {
    SaveCache("mic-input-device", micInputDevice.Name);
  },

  getDefaultMicInput: async () => {
    const micInputs = get().micInputDevices;
    const micInput = await GetCache("mic-input-device");

    if (micInput?.value) {
      return micInputs.find((device) => device.Name === micInput.value);
    }
    return;
  },

  onRecordingStopped(callback) {
    return EventsOn(ON_RECORDING_STOPPED, callback);
  },

  onTranscribedText(callback) {
    return EventsOn(ON_TRANSCRIBED_TEXT, callback);
  },

  onTranscribeError(callback) {
    return EventsOn(ON_TRANSCRIBE_ERROR, callback);
  },

  onTranscriberState(callback) {
    return EventsOn(ON_TRANSCRIBER_STATE, callback);
  },

  onMicLevel(callback) {
    return EventsOn(ON_MIC_LEVEL, callback);
  },
}));
