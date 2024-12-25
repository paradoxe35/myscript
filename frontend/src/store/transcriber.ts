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

  setOnTranscribedText: (callback: (text: string) => void) => void;
  setOnTranscribeError: (callback: (error: Error) => void) => void;
  setOnRecordingStopped: (callback: (autoStopped: boolean) => void) => void;
};

const ON_TRANSCRIBED_TEXT = "on-transcribed-text";
const ON_TRANSCRIBE_ERROR = "on-transcribe-error";
const ON_RECORDING_STOPPED = "on-recording-stopped";

export const useTranscriberStore = create<TranscriberState>((set, get) => ({
  isRecording: false,
  languages: [],

  async startRecording(languageCode: string) {
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

  setOnRecordingStopped(callback: (autoStopped: boolean) => void) {
    EventsOff(ON_RECORDING_STOPPED);
    EventsOn(ON_RECORDING_STOPPED, callback);

    return () => {
      EventsOff(ON_RECORDING_STOPPED);
    };
  },

  setOnTranscribedText(callback: (text: string) => void) {
    EventsOff(ON_TRANSCRIBED_TEXT);
    EventsOn(ON_TRANSCRIBED_TEXT, callback);

    return () => {
      EventsOff(ON_TRANSCRIBED_TEXT);
    };
  },

  setOnTranscribeError(callback: (error: Error) => void) {
    EventsOff(ON_TRANSCRIBE_ERROR);
    EventsOn(ON_TRANSCRIBE_ERROR, callback);

    return () => {
      EventsOff(ON_TRANSCRIBE_ERROR);
    };
  },
}));
