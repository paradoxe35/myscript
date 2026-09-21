import { Cpu, Loader2, RefreshCw, Search } from "lucide-react";
import { PropsWithChildren, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import { cn } from "@/lib/utils";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Input } from "../ui/input";
import { ScrollArea } from "../ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import { useSpeechModelsStore } from "@/store/speech-models";
import { main } from "~wails/models";
import { ModelBody, ModelProps } from "./speech-model-card";

type Filter = "all" | "downloaded" | "multilingual" | "english";

const FILTERS: Array<{ key: Filter; label: string }> = [
  { key: "all", label: "All models" },
  { key: "downloaded", label: "Downloaded" },
  { key: "multilingual", label: "Multilingual" },
  { key: "english", label: "English only" },
];

type BrowserProps = PropsWithChildren<{
  selectedID: string | undefined;
  onSelect: (model: main.SpeechModel) => void;
}>;

export function SpeechModelBrowser({
  selectedID,
  onSelect,
  children,
}: BrowserProps) {
  const [open, setOpen] = useState(false);

  const choose = (model: main.SpeechModel) => {
    onSelect(model);
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>

      <DialogContent className="sm:max-w-[560px] gap-3">
        <DialogHeader>
          <DialogTitle>Speech models</DialogTitle>
          <DialogDescription className="text-xs">
            Run on this computer. No internet needed once downloaded.
          </DialogDescription>
        </DialogHeader>

        <ModelList selectedID={selectedID} onSelect={choose} />
      </DialogContent>
    </Dialog>
  );
}

function ModelList({ selectedID, onSelect }: Omit<BrowserProps, "children">) {
  const models = useSpeechModelsStore((store) => store.models);
  const machine = useSpeechModelsStore((store) => store.machine);
  const fetchMachine = useSpeechModelsStore((store) => store.fetchMachine);

  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<Filter>("all");

  useEffect(() => {
    fetchMachine();
  }, []);

  const visible = useMemo(
    () => models.filter((model) => matches(model, query, filter)),
    [models, query, filter],
  );
  const downloaded = visible.filter((model) => model.Downloaded);
  const available = visible.filter((model) => !model.Downloaded);

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 opacity-50" />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search models or languages"
            className="h-8 pl-8 text-xs"
          />
        </div>
        <Select value={filter} onValueChange={(v) => setFilter(v as Filter)}>
          <SelectTrigger className="h-8 w-[140px] text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {FILTERS.map((item) => (
              <SelectItem key={item.key} value={item.key}>
                {item.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {machine && <MachineChip machine={machine} />}
        <RefreshButton />
      </div>

      <ScrollArea className="h-[55vh] max-h-[460px] -mx-2 px-2">
        <div className="flex flex-col gap-4 pb-2">
          {visible.length === 0 && (
            <p className="text-xs text-center py-8 dark:text-white/50 text-slate-900/50">
              No model matches.
            </p>
          )}

          <Section title="Downloaded" count={downloaded.length}>
            {downloaded.map((model) => (
              <ModelRow
                key={model.ID}
                model={model}
                active={selectedID === model.ID}
                onSelect={() => onSelect(model)}
              />
            ))}
          </Section>

          <Section title="Available" count={available.length}>
            {available.map((model) => (
              <ModelRow
                key={model.ID}
                model={model}
                active={false}
                onSelect={() => onSelect(model)}
              />
            ))}
          </Section>
        </div>
      </ScrollArea>
    </div>
  );
}

function matches(model: main.SpeechModel, query: string, filter: Filter) {
  const languages = model.Languages || [];
  switch (filter) {
    case "downloaded":
      if (!model.Downloaded) return false;
      break;
    case "multilingual":
      if (languages.length < 2) return false;
      break;
    case "english":
      if (languages.length !== 1 || languages[0].Code !== "en") return false;
      break;
  }

  const needle = query.trim().toLowerCase();
  if (!needle) return true;
  return (
    model.Name.toLowerCase().includes(needle) ||
    model.Description.toLowerCase().includes(needle) ||
    languages.some((l) => l.Name.toLowerCase().includes(needle))
  );
}

function Section({
  title,
  count,
  children,
}: {
  title: string;
  count: number;
  children: React.ReactNode;
}) {
  if (count === 0) return null;
  return (
    <section className="flex flex-col gap-2">
      <h4 className="text-[10px] uppercase tracking-wider dark:text-white/40 text-slate-900/40">
        {title} · {count}
      </h4>
      <ul className="flex flex-col gap-2">{children}</ul>
    </section>
  );
}

function RefreshButton() {
  const refreshModels = useSpeechModelsStore((store) => store.refreshModels);
  const [refreshing, setRefreshing] = useState(false);

  const refresh = () => {
    setRefreshing(true);
    refreshModels()
      .catch((err) => toast.error(`Could not refresh the models: ${err}`))
      .finally(() => setRefreshing(false));
  };

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="outline"
          size="icon"
          className="h-8 w-8 shrink-0"
          aria-label="Refresh models"
          disabled={refreshing}
          onClick={refresh}
        >
          {refreshing ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <RefreshCw className="h-3.5 w-3.5" />
          )}
        </Button>
      </TooltipTrigger>
      <TooltipContent className="text-xs">
        Check for new or updated models
      </TooltipContent>
    </Tooltip>
  );
}

function MachineChip({ machine }: { machine: main.MachineInfo }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="inline-flex h-8 items-center gap-1.5 rounded-md border px-2.5 text-[11px] whitespace-nowrap dark:text-white/70 text-slate-900/70">
          <Cpu className="h-3 w-3" />
          {machine.Cores} cores · {formatMemory(machine.MemoryMB)}
        </span>
      </TooltipTrigger>
      <TooltipContent className="max-w-60 text-xs">
        Models are ranked for this computer: speed is estimated from the core
        count and a model needs to fit well within memory.
      </TooltipContent>
    </Tooltip>
  );
}

function ModelRow({ model, active, onSelect }: ModelProps) {
  return (
    <li
      className={cn(
        "rounded-md border p-3 transition-colors",
        active && "border-primary bg-primary/5",
      )}
    >
      <ModelBody model={model} active={active} onSelect={onSelect} />
    </li>
  );
}

function formatMemory(mb: number) {
  if (mb <= 0) return "memory unknown";
  return `${Math.round(mb / 1024)} GB RAM`;
}
