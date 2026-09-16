import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { Button } from "../ui/button";
import { BookOpenText, Loader2, Play } from "lucide-react";
import { useTranscriberStore } from "@/store/transcriber";
import SRInputsModal from "../script-reader-inputs-modal";
import { toast } from "sonner";

export function ScriptReaderControllers(props: React.ComponentProps<"div">) {
  const transcriberStore = useTranscriberStore();
  const activePageStore = useActivePageStore();

  const activePage = activePageStore.page;
  const state = transcriberStore.state;

  const handleStartReading = (languageCode: string, micInputDevice: string) => {
    transcriberStore.startRecording(languageCode, micInputDevice).catch((err) => {
      console.error("Error starting recording:", err);
      toast.error(String(err || "Error starting recording"));
    });
  };

  if (!activePage) return null;

  const preparing = state === "loading" || state === "ready";
  const listening = state === "listening";
  // The ring grows with the microphone level so the reader can see it hears them.
  const ring = listening ? Math.min(8, Math.round(transcriberStore.micLevel * 40)) : 0;

  const button = (
    <Button
      variant="outline"
      size="icon"
      title={
        preparing
          ? `Loading ${transcriberStore.modelName || "model"}… click to cancel`
          : listening
          ? "Stop reading"
          : "Start reading"
      }
      onClick={state !== "idle" ? transcriberStore.stopRecording : undefined}
      className={cn(
        "bg-sidebar-accent hover:bg-sidebar-accent/40 transition-shadow",
        preparing && "bg-amber-300/40 hover:bg-amber-300/60",
        listening && "bg-red-300/40 hover:bg-red-300/60"
      )}
      style={ring ? { boxShadow: `0 0 0 ${ring}px rgba(239, 68, 68, 0.25)` } : undefined}
    >
      {preparing ? (
        <Loader2 className="animate-spin" />
      ) : listening ? (
        <BookOpenText />
      ) : (
        <Play />
      )}
    </Button>
  );

  return (
    <div {...props} className={cn("flex gap-2 items-center", props.className)}>
      {state === "idle" ? (
        <SRInputsModal trigger={button} onStartReading={handleStartReading} />
      ) : (
        button
      )}
    </div>
  );
}
