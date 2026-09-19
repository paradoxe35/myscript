import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import { providerLabel, useAIProvidersStore } from "@/store/ai-providers";
import { Plus, Sparkles } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { AddProviderDialog } from "../ai/add-provider-dialog";
import { ProviderForm } from "../ai/provider-form";
import { useProviderDraft } from "../ai/use-provider-draft";
import {
  Field,
  Hint,
  SettingsCard,
  SettingsGroup,
  SettingsPanel,
} from "../fields";

export function AISection() {
  const providers = useAIProvidersStore((store) => store.providers);
  const active = useAIProvidersStore((store) => store.active);
  const fetchProviders = useAIProvidersStore((store) => store.fetch);
  const activate = useAIProvidersStore((store) => store.activate);
  const remove = useAIProvidersStore((store) => store.remove);

  const [selectedName, setSelectedName] = useState<string>();
  const [adding, setAdding] = useState(false);

  useEffect(() => {
    fetchProviders();
  }, []);

  useEffect(() => {
    if (!selectedName && active) setSelectedName(active);
  }, [active, selectedName]);

  const selected = providers.find((provider) => provider.Name === selectedName);

  const { draft, update, dirty, save } = useProviderDraft(selected);

  const deleteProvider = () => {
    if (!selected) return;

    toast(`Remove ${selected.Name}?`, {
      description: "Its API key is deleted with it.",
      action: {
        label: "Remove",
        onClick: () =>
          remove(selected.Name)
            .then(() => {
              setSelectedName(undefined);
              toast.success(`${selected.Name} removed`);
            })
            .catch((error) => toast.error(String(error))),
      },
    });
  };

  return (
    <SettingsPanel>
      <SettingsGroup
        title="Writing assistant"
        description="The provider the editor asks when you use Ask AI."
      >
        <Field label="Active provider">
          <Select value={active} onValueChange={activate}>
            <SelectTrigger>
              <SelectValue placeholder="Choose a provider" />
            </SelectTrigger>
            <SelectContent>
              {providers.map((provider) => (
                <SelectItem key={provider.Name} value={provider.Name}>
                  {providerLabel(provider)}
                  {!provider.Configured && " — not set up"}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </Field>
      </SettingsGroup>

      <SettingsGroup
        title="Providers"
        description="Each keeps its own key and model."
        action={
          <Button size="sm" variant="outline" onClick={() => setAdding(true)}>
            <Plus className="h-3.5 w-3.5" />
            Add
          </Button>
        }
      >
        <div className="flex flex-wrap gap-2">
          {providers.map((provider) => (
            <Button
              key={provider.Name}
              variant={provider.Name === selectedName ? "secondary" : "outline"}
              size="sm"
              onClick={() => setSelectedName(provider.Name)}
              className={cn(
                "gap-2 rounded-full font-normal",
                provider.Name === selectedName && "border-primary",
              )}
            >
              <span
                className={cn(
                  "h-1.5 w-1.5 rounded-full",
                  provider.Configured
                    ? "bg-emerald-500"
                    : "bg-muted-foreground/40",
                )}
              />
              {providerLabel(provider)}
              {provider.Name === active && (
                <Sparkles className="h-3 w-3 text-primary" />
              )}
            </Button>
          ))}
        </div>

        <SettingsCard>
          {draft ? (
            <ProviderForm
              draft={draft}
              dirty={dirty}
              onChange={update}
              onSave={save}
              onDelete={deleteProvider}
            />
          ) : (
            <Hint>Choose a provider above to set it up.</Hint>
          )}
        </SettingsCard>
      </SettingsGroup>

      <AddProviderDialog
        open={adding}
        onOpenChange={setAdding}
        onAdded={setSelectedName}
      />
    </SettingsPanel>
  );
}
