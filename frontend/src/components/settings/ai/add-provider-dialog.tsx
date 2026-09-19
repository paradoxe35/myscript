import { ApiKeyInput } from "@/components/ui/api-key-input";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAIProvidersStore } from "@/store/ai-providers";
import { Loader2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { main } from "~wails/models";
import { Field, Hint } from "../fields";

type AddProviderDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdded: (name: string) => void;
};

const empty = { name: "", baseURL: "", model: "", apiKey: "", noAPIKey: false };

export function AddProviderDialog({
  open,
  onOpenChange,
  onAdded,
}: AddProviderDialogProps) {
  const save = useAIProvidersStore((store) => store.save);

  const [form, setForm] = useState(empty);
  const [saving, setSaving] = useState(false);

  const update = (patch: Partial<typeof empty>) =>
    setForm((current) => ({ ...current, ...patch }));

  const close = (next: boolean) => {
    if (!next) setForm(empty);
    onOpenChange(next);
  };

  const add = async () => {
    setSaving(true);
    try {
      await save(
        main.AIProvider.createFrom({
          Name: form.name.trim(),
          BaseURL: form.baseURL.trim(),
          Model: form.model.trim(),
          NoAPIKey: form.noAPIKey,
          Custom: true,
        }),
        form.apiKey,
      );

      onAdded(form.name.trim());
      close(false);
      toast.success(`${form.name.trim()} added`);
    } catch (error) {
      toast.error(String(error));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={close}>
      <DialogContent className="sm:max-w-[460px]">
        <DialogHeader>
          <DialogTitle>Add a provider</DialogTitle>
          <DialogDescription className="text-xs">
            Anything that speaks the OpenAI API: Ollama, LM Studio, OpenRouter,
            vLLM, a company gateway.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4 py-2">
          <Field label="Name" hint="Letters, numbers, hyphens and underscores.">
            <Input
              autoFocus
              value={form.name}
              placeholder="ollama"
              onChange={(event) => update({ name: event.target.value })}
            />
          </Field>

          <Field label="Base URL">
            <Input
              value={form.baseURL}
              placeholder="http://localhost:11434/v1"
              onChange={(event) => update({ baseURL: event.target.value })}
            />
          </Field>

          <Field label="Model">
            <Input
              value={form.model}
              placeholder="llama3.1"
              onChange={(event) => update({ model: event.target.value })}
            />
          </Field>

          {!form.noAPIKey && (
            <Field label="API key">
              <ApiKeyInput
                value={form.apiKey}
                onChange={(event) => update({ apiKey: event.target.value })}
              />
            </Field>
          )}

          <Label className="flex cursor-pointer items-start gap-2.5 font-normal">
            <Checkbox
              checked={form.noAPIKey}
              onCheckedChange={(value) => update({ noAPIKey: value === true })}
              className="mt-0.5"
            />
            <span className="flex flex-col gap-0.5">
              <span className="text-sm leading-none">Needs no API key</span>
              <Hint>For a model running on this computer.</Hint>
            </span>
          </Label>
        </div>

        <DialogFooter>
          <Button variant="secondary" onClick={() => close(false)}>
            Cancel
          </Button>
          <Button
            onClick={add}
            disabled={saving || !form.name.trim() || !form.baseURL.trim()}
          >
            {saving && <Loader2 className="h-4 w-4 animate-spin" />}
            Add provider
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
