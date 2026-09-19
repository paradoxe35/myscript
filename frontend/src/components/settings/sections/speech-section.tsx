import { cn } from "@/lib/utils";
import { useSpeechModelsStore } from "@/store/speech-models";
import { useEffect, useState } from "react";
import { IsWitAIAvailable } from "~wails/main/App";
import {
  TranscriberSource,
  TRANSCRIBER_SOURCES,
  useSettings,
} from "../context";
import { Hint, SettingsCard, SettingsGroup, SettingsPanel } from "../fields";
import { SpeechModelsInputs } from "../settings-speech-models";
import { HostedSpeech } from "./hosted-speech";

export function SpeechSection() {
  const { config, updateConfig } = useSettings();
  const fetchModels = useSpeechModelsStore((store) => store.fetchModels);

  // Wit.ai keys are embedded at build time; without them the option would only
  // fail once someone tried to record.
  const [witAIAvailable, setWitAIAvailable] = useState(false);

  const source = (config?.TranscriberSource || "local") as TranscriberSource;

  useEffect(() => {
    IsWitAIAvailable().then(setWitAIAvailable);
  }, []);

  useEffect(() => {
    if (source === "local") fetchModels();
  }, [source]);

  const sources = TRANSCRIBER_SOURCES.filter(
    (option) => option.key !== "witai" || witAIAvailable,
  );

  return (
    <SettingsPanel>
      <SettingsGroup
        title="Transcription"
        description="How MyScript follows along while you read."
      >
        <div className="flex flex-col gap-2">
          {sources.map((option) => (
            <button
              key={option.key}
              type="button"
              onClick={() => updateConfig({ TranscriberSource: option.key })}
              className={cn(
                "flex flex-col items-start gap-0.5 rounded-lg border px-3 py-2.5 text-left transition",
                "hover:bg-accent",
                source === option.key &&
                  "border-primary bg-primary/5 hover:bg-primary/5",
              )}
            >
              <span className="text-sm font-medium">{option.name}</span>
              <span className="text-xs text-muted-foreground">
                {option.description}
              </span>
            </button>
          ))}
        </div>
      </SettingsGroup>

      <SettingsGroup title="Setup">
        <SettingsCard>
          <SourceSetup source={source} />
        </SettingsCard>
      </SettingsGroup>
    </SettingsPanel>
  );
}

function SourceSetup({ source }: { source: TranscriberSource }) {
  switch (source) {
    case "local":
      return <SpeechModelsInputs />;

    case "remote":
      return <HostedSpeech />;

    case "witai":
      return (
        <Hint>
          Wit.ai needs no key. It works over the internet and is less accurate
          than the other options.
        </Hint>
      );
  }
}
