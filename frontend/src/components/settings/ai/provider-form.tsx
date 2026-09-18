import { ApiKeyInput } from "@/components/ui/api-key-input";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import {
  AIModel,
  providerLabel,
  useAIProvidersStore,
} from "@/store/ai-providers";
import { Check, List, Loader2, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Field, Hint } from "../fields";
import { ModelPicker } from "./model-picker";
import { ProviderDraft } from "./use-provider-draft";

type ProviderFormProps = {
  draft: ProviderDraft;
  dirty: boolean;
  onChange: (patch: Partial<ProviderDraft>) => void;
  onSave: () => Promise<void>;
  onDelete: () => void;
};

export function ProviderForm({
  draft,
  dirty,
  onChange,
  onSave,
  onDelete,
}: ProviderFormProps) {
  const listModels = useAIProvidersStore((store) => store.listModels);
  const test = useAIProvidersStore((store) => store.test);

  const [models, setModels] = useState<AIModel[]>([]);
  const [browsing, setBrowsing] = useState(false);
  const [loadingModels, setLoadingModels] = useState(false);
  const [testing, setTesting] = useState(false);
  const [saving, setSaving] = useState(false);

  const needsKey = !draft.NoAPIKey;

  const browseModels = async () => {
    setLoadingModels(true);
    try {
      const available = await listModels(draft, draft.apiKey);
      if (available.length === 0) {
        toast.warning("This provider listed no models");
        return;
      }
      setModels(available);
      setBrowsing(true);
    } catch (error) {
      toast.error(String(error));
    } finally {
      setLoadingModels(false);
    }
  };

  const testConnection = async () => {
    setTesting(true);
    try {
      await test(draft, draft.apiKey);
      toast.success(`${providerLabel(draft)} answered`);
    } catch (error) {
      toast.error(String(error));
    } finally {
      setTesting(false);
    }
  };

  const save = async () => {
    setSaving(true);
    try {
      await onSave();
      toast.success("Provider saved");
    } catch (error) {
      toast.error(String(error));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col gap-4">
      {needsKey && (
        <Field
          label="API key"
          hint="Stored encrypted on this computer and never uploaded."
        >
          <ApiKeyInput
            value={draft.apiKey}
            placeholder={`Your ${providerLabel(draft)} API key`}
            onChange={(event) => onChange({ apiKey: event.target.value })}
          />
        </Field>
      )}

      {draft.Custom && (
        <Field label="Base URL" hint="Any OpenAI-compatible endpoint.">
          <Input
            value={draft.BaseURL}
            placeholder="http://localhost:11434/v1"
            onChange={(event) => onChange({ BaseURL: event.target.value })}
          />
        </Field>
      )}

      <Field label="Model">
        <div className="flex gap-2">
          <Input
            value={draft.Model}
            placeholder="Model name"
            onChange={(event) => onChange({ Model: event.target.value })}
          />
          <Button
            variant="outline"
            size="icon"
            className="shrink-0"
            title="Browse models"
            disabled={loadingModels}
            onClick={browseModels}
          >
            {loadingModels ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <List className="h-4 w-4" />
            )}
          </Button>
        </div>
      </Field>

      <div className="flex flex-col gap-2">
        {draft.Custom && (
          <Toggle
            checked={draft.NoAPIKey}
            onChange={(checked) => onChange({ NoAPIKey: checked })}
            label="This provider needs no API key"
            hint="For a model running on this computer."
          />
        )}

        <Toggle
          checked={draft.LowReasoning}
          onChange={(checked) => onChange({ LowReasoning: checked })}
          label="Ask reasoning models to think less"
          hint="Rewriting a sentence rarely needs it, and the thinking is billed."
        />
      </div>

      <div className="flex items-center gap-2 border-t pt-4">
        <Button onClick={save} disabled={!dirty || saving}>
          {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
          Save
        </Button>

        <Button variant="outline" onClick={testConnection} disabled={testing}>
          {testing ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Check className="h-4 w-4" />
          )}
          Test connection
        </Button>

        {draft.Custom && (
          <Button
            variant="ghost"
            size="icon"
            className="ml-auto text-destructive hover:text-destructive"
            title="Remove this provider"
            onClick={onDelete}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        )}
      </div>

      <ModelPicker
        open={browsing}
        models={models}
        selected={draft.Model}
        onSelect={(model) => onChange({ Model: model.ID })}
        onOpenChange={setBrowsing}
      />
    </div>
  );
}

type ToggleProps = {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label: string;
  hint?: string;
};

function Toggle({ checked, onChange, label, hint }: ToggleProps) {
  return (
    <label className="flex cursor-pointer items-start gap-2.5 rounded-md py-1">
      <Checkbox
        checked={checked}
        onCheckedChange={(value) => onChange(value === true)}
        className="mt-0.5"
      />
      <span className="flex flex-col gap-0.5">
        <span className="text-sm leading-none">{label}</span>
        {hint && <Hint>{hint}</Hint>}
      </span>
    </label>
  );
}
