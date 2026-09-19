import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import { AIModel } from "@/store/ai-providers";
import { Check, Search } from "lucide-react";
import { useMemo, useState } from "react";

type ModelPickerProps = {
  open: boolean;
  models: AIModel[];
  selected?: string;
  onSelect: (model: AIModel) => void;
  onOpenChange: (open: boolean) => void;
};

export function ModelPicker({
  open,
  models,
  selected,
  onSelect,
  onOpenChange,
}: ModelPickerProps) {
  const [query, setQuery] = useState("");

  const visible = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return models;

    return models.filter(
      (model) =>
        model.ID.toLowerCase().includes(needle) ||
        model.Name?.toLowerCase().includes(needle),
    );
  }, [models, query]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[520px] gap-3">
        <DialogHeader>
          <DialogTitle>Choose a model</DialogTitle>
          <DialogDescription className="text-xs">
            {models.length} models available from this provider.
          </DialogDescription>
        </DialogHeader>

        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            autoFocus
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search models"
            className="pl-8"
          />
        </div>

        <ScrollArea className="h-[320px] -mx-1 px-1">
          <div className="flex flex-col gap-1">
            {visible.map((model) => (
              <Button
                key={model.ID}
                variant="ghost"
                onClick={() => {
                  onSelect(model);
                  onOpenChange(false);
                }}
                className={cn(
                  "h-auto justify-between gap-3 px-3 py-2 font-normal",
                  model.ID === selected && "border border-primary bg-primary/5",
                )}
              >
                <span className="flex min-w-0 flex-col items-start">
                  <span className="truncate font-medium">{model.ID}</span>
                  {describe(model) && (
                    <span className="truncate text-xs text-muted-foreground">
                      {describe(model)}
                    </span>
                  )}
                </span>

                {model.ID === selected && (
                  <Check className="h-4 w-4 shrink-0 text-primary" />
                )}
              </Button>
            ))}

            {visible.length === 0 && (
              <p className="px-3 py-6 text-center text-sm text-muted-foreground">
                No model matches "{query}".
              </p>
            )}
          </div>
        </ScrollArea>

        <Button variant="secondary" onClick={() => onOpenChange(false)}>
          Cancel
        </Button>
      </DialogContent>
    </Dialog>
  );
}

// Providers usually name a model after its id; a gateway is where the name adds something.
function describe(model: AIModel) {
  const normalize = (value: string) =>
    value.toLowerCase().replace(/[^a-z0-9]/g, "");
  return !model.Name || normalize(model.Name) === normalize(model.ID)
    ? ""
    : model.Name;
}
