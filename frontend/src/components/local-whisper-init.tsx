import { Progress } from "@/components/ui/progress";
import { useLocalWhisperStore } from "@/store/local-whisper";
import { useEffect } from "react";
import { toast } from "sonner";

export function LocalWhisperInit() {
  const localWhisperStore = useLocalWhisperStore();

  useEffect(() => {
    return localWhisperStore.onWhisperModelDownloadProgress((progress) => {
      const pct = Math.round((progress.Size / progress.Total) * 100);
      console.log(`Downloading model: ${progress.Name} (${pct}%)`);

      toast.info(
        <div className="flex flex-col">
          <span className="text-sm mb-1">
            Downloading model: {progress.Name} ({pct}%)
          </span>

          <Progress value={pct} />
        </div>,

        { id: progress.Name, duration: 6000 }
      );
    });
  }, []);

  useEffect(() => {
    return localWhisperStore.onWhisperModelDownloadError((error) => {
      console.log("Error downloading model:", error);
      toast.error("Error downloading model: " + error, {
        id: "model-download",
      });
    });
  }, []);

  return null;
}
