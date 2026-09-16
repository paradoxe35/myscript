import { useActivePageStore } from "@/store/active-page";
import { useTranscriberStore } from "@/store/transcriber";
import { useEffect } from "react";
import { toast } from "sonner";

const STATE_TOAST = "transcriber-state";

export function TranscriberInit() {
  const transcriberStore = useTranscriberStore();
  const activePageStore = useActivePageStore();

  useEffect(() => {
    transcriberStore.getRecordingStatus();
  }, []);

  // Text still in flight belongs to the page being left.
  useEffect(() => {
    transcriberStore.cancelRecording();
  }, [activePageStore.getPageId()]);

  useEffect(() => {
    return transcriberStore.onTranscriberState((event) => {
      transcriberStore.setState(event);

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
    return transcriberStore.onMicLevel(transcriberStore.setMicLevel);
  }, []);

  useEffect(() => {
    return transcriberStore.onTranscribeError((error) => {
      console.log("Transcription error:", error);
      toast.error("Transcription error: " + error, { id: STATE_TOAST });
      transcriberStore.getRecordingStatus();
    });
  }, []);

  useEffect(() => {
    return transcriberStore.onRecordingStopped((autoStopped) => {
      transcriberStore.getRecordingStatus();
      console.log("Recording stopped:", autoStopped);

      if (autoStopped) {
        toast.info("30 seconds of silence detected, stopping the read mode");
      }
    });
  }, []);

  // Read mode outlives the take so the reader can restart or leave on their own.
  useEffect(() => {
    if (transcriberStore.state === "listening") activePageStore.setReadMode(true);
  }, [transcriberStore.state]);

  return null;
}
