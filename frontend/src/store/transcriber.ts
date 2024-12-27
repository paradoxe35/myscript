import { create } from "zustand";
import {
  GetLanguages,
  IsRecording,
  StartRecording,
  StopRecording,
} from "~wails/main/App";
import { structs } from "~wails/models";
import { EventsOn, EventsOff } from "~wails-runtime";

type TranscriberState = {
  isRecording: boolean;
  languages: Array<structs.Language>;

  startRecording: (languageCode: string) => void;
  stopRecording: () => void;
  getRecordingStatus: () => void;
  getLanguages: () => void;

  onTranscribedText: (callback: (text: string) => void) => void;
  onTranscribeError: (callback: (error: string) => void) => void;
  onRecordingStopped: (callback: (autoStopped: boolean) => void) => void;
};

const ON_TRANSCRIBED_TEXT = "on-transcribed-text";
const ON_TRANSCRIBE_ERROR = "on-transcribe-error";
const ON_RECORDING_STOPPED = "on-recording-stopped";

export const useTranscriberStore = create<TranscriberState>((set, get) => ({
  isRecording: false,
  languages: [],

  async startRecording(languageCode) {
    if (!get().isRecording) {
      await StartRecording(languageCode);
      get().getRecordingStatus();
    }
  },

  async stopRecording() {
    if (get().isRecording) {
      await StopRecording();
      get().getRecordingStatus();
    }
  },

  getRecordingStatus: () => {
    return IsRecording().then((isRecording) => {
      set({ isRecording });

      return isRecording;
    });
  },

  getLanguages: () => {
    GetLanguages().then((languages) => {
      set({ languages });
    });
  },

  onRecordingStopped(callback) {
    EventsOff(ON_RECORDING_STOPPED);
    EventsOn(ON_RECORDING_STOPPED, callback);

    return () => {
      EventsOff(ON_RECORDING_STOPPED);
    };
  },

  onTranscribedText(callback) {
    EventsOff(ON_TRANSCRIBED_TEXT);
    EventsOn(ON_TRANSCRIBED_TEXT, callback);

    return () => {
      EventsOff(ON_TRANSCRIBED_TEXT);
    };
  },

  onTranscribeError(callback) {
    EventsOff(ON_TRANSCRIBE_ERROR);
    EventsOn(ON_TRANSCRIBE_ERROR, callback);

    return () => {
      EventsOff(ON_TRANSCRIBE_ERROR);
    };
  },
}));
