import { Progress } from "@/components/ui/progress";
import { useSpeechModelsStore } from "@/store/speech-models";
import { useEffect } from "react";
import { toast } from "sonner";

export function SpeechModelsInit() {
  const store = useSpeechModelsStore();

  useEffect(() => {
    store.fetchModels();
  }, []);

  useEffect(() => {
    return store.onDownloadProgress((event) => {
      store.setProgress(event);

      const pct = event.Total
        ? Math.round((event.Downloaded / event.Total) * 100)
        : 0;
      const label =
        event.Stage === "verifying"
          ? `Verifying ${event.Name}`
          : `Downloading ${event.Name} (${pct}%)`;

      toast.info(
        <div className="flex flex-col w-full">
          <span className="text-sm mb-1">{label}</span>
          <Progress value={pct} />
        </div>,
        { id: event.ID, duration: 6000 }
      );
    });
  }, []);

  useEffect(() => {
    return store.onDownloadDone((event) => {
      store.clearProgress(event.ID);
      store.fetchModels();
      toast.success(`${event.Name} is ready to use`, { id: event.ID });
    });
  }, []);

  useEffect(() => {
    return store.onDownloadCancelled((event) => {
      store.clearProgress(event.ID);
      store.fetchModels();
      toast.dismiss(event.ID);
    });
  }, []);

  useEffect(() => {
    return store.onDownloadError((event) => {
      store.clearProgress(event.ID);
      store.fetchModels();
      toast.error(`Could not download ${event.Name}: ${event.Error}`, {
        id: event.ID,
      });
    });
  }, []);

  return null;
}
