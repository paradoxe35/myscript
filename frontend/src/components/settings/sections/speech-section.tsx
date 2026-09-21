import { Card } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { TranscriberSource, transcriberSource } from "@/store/device-settings";
import { useSpeechModelsStore } from "@/store/speech-models";
import { useEffect, useState } from "react";
import { IsWitAIAvailable } from "~wails/main/App";
import { TRANSCRIBER_SOURCES, useSettings } from "../context";
import { Hint, SettingsCard, SettingsGroup, SettingsPanel } from "../fields";
import { SpeechModelsInputs } from "../settings-speech-models";
import { HostedSpeech } from "./hosted-speech";

export function SpeechSection() {
  const { deviceSettings, updateDeviceSettings } = useSettings();
  const fetchModels = useSpeechModelsStore((store) => store.fetchModels);

  // Wit.ai keys are embedded at build time; without them the option would only
  // fail once someone tried to record.
  const [witAIAvailable, setWitAIAvailable] = useState(false);

  const source = transcriberSource(deviceSettings);

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
            <Card
              key={option.key}
              role="radio"
              tabIndex={0}
              aria-checked={source === option.key}
              onClick={() =>
                updateDeviceSettings({ TranscriberSource: option.key })
              }
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  updateDeviceSettings({ TranscriberSource: option.key });
                }
              }}
              className={cn(
                "cursor-pointer px-3 py-2.5 shadow-none transition",
                "hover:bg-accent focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
                source === option.key &&
                  "border-primary bg-primary/5 hover:bg-primary/5",
              )}
            >
              <p className="text-sm font-medium">{option.name}</p>
              <Hint>{option.description}</Hint>
            </Card>
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
