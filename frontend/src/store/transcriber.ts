import { create } from "zustand";
import {
  GetCache,
  GetLanguages,
  GetMicInputDevices,
  IsRecording,
  SaveCache,
  StartRecording,
  StopRecording,
} from "~wails/main/App";
import { microphone, structs } from "~wails/models";
import { EventsOn } from "~wails-runtime";
import { EventClear } from "@/types";

type TranscriberState = {
  isRecording: boolean;
  languages: Array<structs.Language>;
  micInputDevices: Array<microphone.MicInputDevice>;

  startRecording: (
    languageCode: string,
    micInputDeviceID: number[]
  ) => Promise<void>;
  stopRecording: () => Promise<void>;
  getRecordingStatus: () => void;
  getMicInputDevices: () => Promise<Array<microphone.MicInputDevice>>;
  getLanguages: () => void;

  // Page language
  getPageLanguage: (pageId: string | number) => Promise<string | null>;
  setPageLanguage: (pageId: string | number, language: string) => Promise<void>;

  // Default mic input device
  setDefaultMicInput: (micInputDevice: microphone.MicInputDevice) => void;
  getDefaultMicInput: () => Promise<microphone.MicInputDevice | undefined>;

  onTranscribedText: (callback: (text: string) => void) => EventClear;
  onTranscribeError: (callback: (error: string) => void) => EventClear;
  onRecordingStopped: (callback: (autoStopped: boolean) => void) => EventClear;
};

const ON_TRANSCRIBED_TEXT = "on-transcribed-text";
const ON_TRANSCRIBE_ERROR = "on-transcribe-error";
const ON_RECORDING_STOPPED = "on-recording-stopped";

export const useTranscriberStore = create<TranscriberState>((set, get) => ({
  isRecording: false,
  languages: [],
  micInputDevices: [],

  async startRecording(languageCode, micInputDeviceID) {
    if (!get().isRecording) {
      const micDeviceID = JSON.stringify(micInputDeviceID);

      return StartRecording(languageCode, micDeviceID).finally(() => {
        get().getRecordingStatus();
      });
    }
  },

  async stopRecording() {
    if (get().isRecording) {
      return StopRecording().finally(() => {
        get().getRecordingStatus();
      });
    }
  },

  getRecordingStatus: () => {
    return IsRecording().then((isRecording) => {
      set({ isRecording });

      return isRecording;
    });
  },

  getMicInputDevices: () => {
    return GetMicInputDevices().then((micInputDevices) => {
      set({ micInputDevices });

      return micInputDevices;
    });
  },

  getLanguages: () => {
    GetLanguages().then((languages) => {
      set({ languages });
    });
  },

  getPageLanguage: async (pageId) => {
    const language = await GetCache(`page-${pageId}-language`);
    return language?.value;
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
}));
