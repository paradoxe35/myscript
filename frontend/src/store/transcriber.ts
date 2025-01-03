import { create } from "zustand";
import {
  GetLanguages,
  GetMicInputDevices,
  IsRecording,
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

  startRecording: (languageCode: string) => Promise<void>;
  stopRecording: () => Promise<void>;
  getRecordingStatus: () => void;
  getMicInputDevices: () => Promise<Array<microphone.MicInputDevice>>;
  getLanguages: () => void;

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

  async startRecording(languageCode) {
    if (!get().isRecording) {
      return StartRecording(languageCode).finally(() => {
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
