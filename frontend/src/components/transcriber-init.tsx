import { useActivePageStore } from "@/store/active-page";
import { useTranscriberStore } from "@/store/transcriber";
import { useEffect } from "react";
import { toast } from "sonner";
import { useShallow } from "zustand/react/shallow";

const STATE_TOAST = "transcriber-state";

export function TranscriberInit() {
  // Actions only: this component must not re-render on every mic level tick.
  const {
    getRecordingStatus,
    cancelRecording,
    setState,
    setMicLevel,
    onTranscriberState,
    onMicLevel,
    onTranscribeError,
    onRecordingStopped,
  } = useTranscriberStore(
    useShallow((store) => ({
      getRecordingStatus: store.getRecordingStatus,
      cancelRecording: store.cancelRecording,
      setState: store.setState,
      setMicLevel: store.setMicLevel,
      onTranscriberState: store.onTranscriberState,
      onMicLevel: store.onMicLevel,
      onTranscribeError: store.onTranscribeError,
      onRecordingStopped: store.onRecordingStopped,
    }))
  );
  const state = useTranscriberStore((store) => store.state);

  const pageId = useActivePageStore((store) => store.getPageId());
  const setReadMode = useActivePageStore((store) => store.setReadMode);

  useEffect(() => {
    getRecordingStatus();
  }, []);

  // Text still in flight belongs to the page being left.
  useEffect(() => {
    cancelRecording();
  }, [pageId]);

  useEffect(() => {
    return onTranscriberState((event) => {
      setState(event);

      switch (event.State) {
        case "loading":
          toast.loading(`Loading ${event.ModelName}…`, {
            id: STATE_TOAST,
            duration: Infinity,
          });
          break;
        case "ready":
          toast.success(`${event.ModelName} is ready`, {
            id: STATE_TOAST,
            duration: 2000,
          });
          break;
        case "listening":
          toast.dismiss(STATE_TOAST);
          break;
        case "idle":
          toast.dismiss(STATE_TOAST);
          break;
      }
    });
  }, []);

  useEffect(() => {
    return onMicLevel(setMicLevel);
  }, []);

  useEffect(() => {
    return onTranscribeError((error) => {
      console.log("Transcription error:", error);
      toast.error("Transcription error: " + error, { id: STATE_TOAST });
      getRecordingStatus();
    });
  }, []);

  useEffect(() => {
    return onRecordingStopped((autoStopped) => {
      getRecordingStatus();
      console.log("Recording stopped:", autoStopped);

      if (autoStopped) {
        toast.info("30 seconds of silence detected, stopping the read mode");
      }
    });
  }, []);

  // Read mode outlives the take so the reader can restart or leave on their own.
  useEffect(() => {
    if (state === "listening") setReadMode(true);
  }, [state]);

  return null;
}
