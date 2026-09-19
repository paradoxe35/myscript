import { ApiKeyInput } from "@/components/ui/api-key-input";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { SpeechService, useSpeechServicesStore } from "@/store/speech-services";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { useSettings } from "../context";
import { Field, Hint } from "../fields";

export function HostedSpeech() {
  const { config, updateConfig } = useSettings();

  const services = useSpeechServicesStore((store) => store.services);
  const fetchServices = useSpeechServicesStore((store) => store.fetch);
  const loadAPIKey = useSpeechServicesStore((store) => store.apiKey);
  const saveAPIKey = useSpeechServicesStore((store) => store.saveAPIKey);

  const [apiKey, setApiKey] = useState("");
  const [savedKey, setSavedKey] = useState("");
  const [savingKey, setSavingKey] = useState(false);

  const selectedID = config?.RemoteProvider || services[0]?.ID || "";
  const service = services.find((item) => item.ID === selectedID);

  useEffect(() => {
    fetchServices();
  }, []);

  useEffect(() => {
    if (!selectedID) return;

    loadAPIKey(selectedID).then((stored) => {
      setApiKey(stored);
      setSavedKey(stored);
    });
  }, [selectedID]);

  const selectService = (id: string) => {
    const next = services.find((item) => item.ID === id);
    if (!next) return;

    updateConfig({
      RemoteProvider: id,
      // A known service owns its endpoint, and its models are its own: Gemini
      // and Whisper share no names, so carrying one over would only fail.
      RemoteBaseURL: next.Custom ? "" : next.BaseURL,
      RemoteModel: next.Models?.[0] ?? "",
    });
  };

  const storeAPIKey = async () => {
    setSavingKey(true);
    try {
      await saveAPIKey(selectedID, apiKey.trim());
      setSavedKey(apiKey.trim());
      setApiKey(apiKey.trim());
      toast.success("API key saved");
    } catch (error) {
      toast.error(String(error));
    } finally {
      setSavingKey(false);
    }
  };

  return (
    <div className="flex flex-col gap-4">
      <Field label="Service">
        <Select value={selectedID} onValueChange={selectService}>
          <SelectTrigger>
            <SelectValue placeholder="Choose a service" />
          </SelectTrigger>
          <SelectContent>
            {services.map((item) => (
              <SelectItem key={item.ID} value={item.ID}>
                {item.Name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </Field>

      <ModelField service={service} />

      <Field
        label="Endpoint"
        hint={
          service?.Custom
            ? "Any OpenAI-compatible endpoint."
            : "Set by the service."
        }
      >
        <Input
          value={config?.RemoteBaseURL || service?.BaseURL || ""}
          disabled={!service?.Custom}
          placeholder="https://api.example.com/v1"
          onChange={(event) =>
            updateConfig({ RemoteBaseURL: event.target.value })
          }
        />
      </Field>

      <Field label="API key" hint={service?.KeyHint}>
        <div className="flex gap-2">
          <ApiKeyInput
            value={apiKey}
            onChange={(event) => setApiKey(event.target.value)}
          />
          <Button
            variant="outline"
            className="shrink-0"
            disabled={savingKey || apiKey.trim() === savedKey}
            onClick={storeAPIKey}
          >
            {savingKey && <Loader2 className="h-4 w-4 animate-spin" />}
            Save
          </Button>
        </div>
      </Field>

      <Hint>Audio is sent to this service. Nothing is downloaded.</Hint>
    </div>
  );
}

function ModelField({ service }: { service: SpeechService | undefined }) {
  const { config, updateConfig } = useSettings();
  const suggestions = service?.Models ?? [];

  return (
    <Field
      label="Model"
      hint={
        suggestions.length === 0 ? "Whatever the endpoint serves." : undefined
      }
    >
      <Input
        list={suggestions.length > 0 ? "speech-model-suggestions" : undefined}
        value={config?.RemoteModel || ""}
        placeholder={suggestions[0] ?? "Model name"}
        onChange={(event) => updateConfig({ RemoteModel: event.target.value })}
      />
      {suggestions.length > 0 && (
        <datalist id="speech-model-suggestions">
          {suggestions.map((model) => (
            <option key={model} value={model} />
          ))}
        </datalist>
      )}
    </Field>
  );
}
