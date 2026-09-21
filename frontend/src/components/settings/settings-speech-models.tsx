import { Sparkles } from "lucide-react";

import { Button } from "../ui/button";
import { Label } from "../ui/label";
import { TooltipProvider } from "../ui/tooltip";
import { useSettings } from "./context";
import { useSpeechModelsStore } from "@/store/speech-models";
import { ModelBody } from "./speech-model-card";
import { SpeechModelBrowser } from "./speech-model-browser";

export function SpeechModelsInputs() {
  const { deviceSettings, updateDeviceSettings } = useSettings();
  const models = useSpeechModelsStore((store) => store.models);

  const selectedID = deviceSettings?.SpeechModelID;
  const selected = models.find((model) => model.ID === selectedID);
  const suggested = models.find((model) => model.Suggested);

  return (
    <TooltipProvider delayDuration={200}>
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between gap-3">
          <Label className="text-xs dark:text-white/70 text-slate-900/70">
            Speech model
          </Label>
          <SpeechModelBrowser
            selectedID={selectedID}
            onSelect={(model) =>
              updateDeviceSettings({ SpeechModelID: model.ID })
            }
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
              suggested && updateDeviceSettings({ SpeechModelID: suggested.ID })
            }
          />
        )}
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
        <Button
          size="sm"
          variant="secondary"
          className="h-8 shrink-0"
          onClick={onSuggested}
        >
          <Sparkles className="h-3.5 w-3.5" />
          Use {suggested}
        </Button>
      )}
    </div>
  );
}
