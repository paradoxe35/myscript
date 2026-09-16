import { Sparkles } from "lucide-react";
import { useEffect } from "react";
import { toast } from "sonner";

import { Button } from "../ui/button";
import { Label } from "../ui/label";
import { TooltipProvider } from "../ui/tooltip";
import { useSettings } from "./context";
import { useSpeechModelsStore } from "@/store/speech-models";
import { ModelBody } from "./speech-model-card";
import { SpeechModelBrowser } from "./speech-model-browser";

export function SpeechModelsInputs() {
  const { state, dispatch } = useSettings();
  const models = useSpeechModelsStore((store) => store.models);
  const hasLegacyFiles = useSpeechModelsStore((store) => store.hasLegacyFiles);
  const checkLegacyFiles = useSpeechModelsStore(
    (store) => store.checkLegacyFiles
  );

  useEffect(() => {
    checkLegacyFiles();
  }, []);

  const selected = models.find((model) => model.ID === state.SpeechModelID);
  const suggested = models.find((model) => model.Suggested);

  return (
    <TooltipProvider delayDuration={200}>
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between gap-3">
          <Label className="text-xs dark:text-white/70 text-slate-900/70">
            Speech model
          </Label>
          <SpeechModelBrowser
            selectedID={state.SpeechModelID}
            onSelect={(model) => dispatch({ SpeechModelID: model.ID })}
          >
            <Button size="sm" variant="outline" className="h-8">
              {selected ? "Change model" : "Choose a model"}
            </Button>
          </SpeechModelBrowser>
        </div>

        {selected ? (
          <div className="rounded-md border border-primary bg-primary/5 p-3">
            <ModelBody model={selected} active onSelect={() => {}} />
          </div>
        ) : (
          <NoModel
            suggested={suggested?.Name}
            onSuggested={() =>
              suggested && dispatch({ SpeechModelID: suggested.ID })
            }
          />
        )}

        {hasLegacyFiles && <LegacyFilesNotice />}
      </div>
    </TooltipProvider>
  );
}

function NoModel({
  suggested,
  onSuggested,
}: {
  suggested: string | undefined;
  onSuggested: () => void;
}) {
  return (
    <div className="rounded-md border border-dashed p-3 flex items-center justify-between gap-3">
      <p className="text-xs dark:text-white/55 text-slate-900/55">
        No model selected. Speech runs on this computer once a model is
        downloaded.
      </p>
      {suggested && (
        <Button size="sm" variant="secondary" className="h-8 shrink-0" onClick={onSuggested}>
          <Sparkles className="h-3.5 w-3.5" />
          Use {suggested}
        </Button>
      )}
    </div>
  );
}

function LegacyFilesNotice() {
  const removeLegacyFiles = useSpeechModelsStore(
    (store) => store.removeLegacyFiles
  );

  const confirm = () => {
    toast("Remove the old Whisper model files?", {
      description:
        "They were used by earlier versions and cannot be read any more.",
      action: {
        label: "Remove",
        onClick: () => {
          removeLegacyFiles()
            .then((removed) =>
              toast.success(`Removed ${removed} old model file(s)`)
            )
            .catch((err) => toast.error(String(err)));
        },
      },
    });
  };

  return (
    <div className="flex items-center justify-between gap-3 rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2">
      <p className="text-xs dark:text-white/70 text-slate-900/70">
        Model files from an earlier version are taking up space.
      </p>
      <Button size="sm" variant="secondary" onClick={confirm}>
        Remove old files
      </Button>
    </div>
  );
}
