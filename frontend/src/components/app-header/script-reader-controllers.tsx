import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { Button } from "../ui/button";
import { BookOpenText, Play } from "lucide-react";
import { useTranscriberStore } from "@/store/transcriber";
import { useEffect } from "react";
import { SRLanguagesModal } from "../sr-languages-modal";
import { toast } from "sonner";

export function ScriptReaderControllers(props: React.ComponentProps<"div">) {
  const transcriberStore = useTranscriberStore();

  const activePageStore = useActivePageStore();
  const activePage = activePageStore.page;
  const readMode = activePageStore.readMode;

  const handleLanguageSelected = (languageCode: string) => {
    transcriberStore.startRecording(languageCode);
  };

  useEffect(() => {
    transcriberStore.getRecordingStatus();
  }, []);

  useEffect(() => {
    return transcriberStore.setOnTranscribedText((text) => {
      console.log("Transcribed text:", text);
      // toast.success("Transcription saved successfully!");
    });
  }, []);

  // Stop recording when page changes
  useEffect(() => {
    transcriberStore.stopRecording();
  }, [activePageStore.getPageId()]);

  useEffect(() => {
    return transcriberStore.setOnTranscribeError((error) => {
      console.log("Transcription error:", error);
      toast.error("Transcription error: " + error);
    });
  }, []);

  useEffect(() => {
    return transcriberStore.setOnRecordingStopped((autoStopped) => {
      console.log("Recording stopped:", autoStopped);

      if (autoStopped) {
        toast.success("10 seconds of silence detected, stopping the read mode");
      }
    });
  }, []);

  useEffect(() => {
    activePageStore.setReadMode(transcriberStore.isRecording);
  }, [transcriberStore.isRecording]);

  if (!activePage) return null;

  const button = (
    <Button
      variant="outline"
      size="icon"
      onClick={
        transcriberStore.isRecording
          ? transcriberStore.stopRecording
          : undefined
      }
      className={cn(
        "bg-sidebar-accent hover:bg-sidebar-accent/40",
        readMode && "bg-red-300/40 hover:bg-red-300/60"
      )}
    >
      {readMode ? <BookOpenText /> : <Play />}
    </Button>
  );

  return (
    <div {...props} className={cn("flex gap-2 items-center", props.className)}>
      {readMode ? (
        button
      ) : (
        <SRLanguagesModal
          trigger={button}
          onLanguageSelected={handleLanguageSelected}
        />
      )}
    </div>
  );
}
