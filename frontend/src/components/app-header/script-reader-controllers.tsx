import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { Button } from "../ui/button";
import { BookOpenText, Play } from "lucide-react";
import { useTranscriberStore } from "@/store/transcriber";
import SRInputsModal from "../script-reader-inputs-modal";
import { toast } from "sonner";

export function ScriptReaderControllers(props: React.ComponentProps<"div">) {
  const transcriberStore = useTranscriberStore();
  const activePageStore = useActivePageStore();

  const activePage = activePageStore.page;
  const readMode = activePageStore.readMode;

  const handleLanguageSelected = (languageCode: string) => {
    transcriberStore.startRecording(languageCode).catch((err) => {
      console.error("Error starting recording:", err);
      toast.error(err || "Error starting recording");
    });
  };

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
        <SRInputsModal
          trigger={button}
          onLanguageSelected={handleLanguageSelected}
        />
      )}
    </div>
  );
}
