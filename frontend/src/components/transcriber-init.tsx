import { useActivePageStore } from "@/store/active-page";
import { useTranscriberStore } from "@/store/transcriber";
import { useEffect } from "react";
import { toast } from "sonner";

export function TranscriberInit() {
  const transcriberStore = useTranscriberStore();
  const activePageStore = useActivePageStore();

  useEffect(() => {
    transcriberStore.getRecordingStatus();
  }, []);

  useEffect(() => {
    return transcriberStore.onTranscribedText((text) => {
      console.log("Transcribed text:", text);
      // toast.success("Transcription saved successfully!");
    });
  }, []);

  // Stop recording when page changes
  useEffect(() => {
    transcriberStore.stopRecording();
  }, [activePageStore.getPageId()]);

  useEffect(() => {
    return transcriberStore.onTranscribeError((error) => {
      console.log("Transcription error:", error);
      toast.error("Transcription error: " + error);
    });
  }, []);

  useEffect(() => {
    return transcriberStore.onRecordingStopped((autoStopped) => {
      console.log("Recording stopped:", autoStopped);

      if (autoStopped) {
        toast.success("10 seconds of silence detected, stopping the read mode");
      }
    });
  }, []);

  useEffect(() => {
    activePageStore.setReadMode(transcriberStore.isRecording);
  }, [transcriberStore.isRecording]);

  return null;
}
