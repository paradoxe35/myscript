import {
  Check,
  Cpu,
  Download,
  Gauge,
  HardDrive,
  Languages,
  Loader2,
  Radio,
  Search,
  Sparkles,
  Trash2,
  X,
  Zap,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import { cn } from "@/lib/utils";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Progress } from "../ui/progress";
import { ScrollArea } from "../ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../ui/tooltip";
import { useSettings } from "./context";
import { useSpeechModelsStore } from "@/store/speech-models";
import { main } from "~wails/models";

type Filter = "all" | "downloaded" | "multilingual" | "english";

const FILTERS: Array<{ key: Filter; label: string }> = [
  { key: "all", label: "All models" },
  { key: "downloaded", label: "Downloaded" },
  { key: "multilingual", label: "Multilingual" },
  { key: "english", label: "English only" },
];

export function SpeechModelsInputs() {
  const { state, dispatch } = useSettings();
  const models = useSpeechModelsStore((store) => store.models);
  const machine = useSpeechModelsStore((store) => store.machine);
  const fetchMachine = useSpeechModelsStore((store) => store.fetchMachine);
  const hasLegacyFiles = useSpeechModelsStore((store) => store.hasLegacyFiles);
  const checkLegacyFiles = useSpeechModelsStore(
    (store) => store.checkLegacyFiles
  );

  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<Filter>("all");

  useEffect(() => {
    fetchMachine();
    checkLegacyFiles();
  }, []);

  const browsing = query.trim() === "" && filter === "all";
  const visible = useMemo(
    () => models.filter((model) => matches(model, query, filter)),
    [models, query, filter]
  );
  const suggested = models.find((model) => model.Suggested);
  const downloaded = visible.filter((model) => model.Downloaded);
  const available = visible.filter((model) => !model.Downloaded);

  const select = (model: main.SpeechModel) =>
    dispatch({ SpeechModelID: model.ID });

  return (
    <TooltipProvider delayDuration={200}>
      <div className="flex flex-col gap-3">
        <div className="flex items-start justify-between gap-3">
          <div className="flex flex-col gap-1">
            <Label className="text-xs dark:text-white/70 text-slate-900/70">
              Speech models
            </Label>
            <p className="text-xs dark:text-white/50 text-slate-900/50">
              Run on this computer. No internet needed once downloaded.
            </p>
          </div>
          {machine && <MachineChip machine={machine} />}
        </div>

        {hasLegacyFiles && <LegacyFilesNotice />}

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
        </div>

        <ScrollArea className="h-80 -mx-2 px-2">
          <div className="flex flex-col gap-4 pb-2">
            {browsing && suggested && (
              <SuggestedCard
                model={suggested}
                active={state.SpeechModelID === suggested.ID}
                onSelect={() => select(suggested)}
              />
            )}

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
                  active={state.SpeechModelID === model.ID}
                  onSelect={() => select(model)}
                />
              ))}
            </Section>

            <Section
              title={browsing ? "More models" : "Available"}
              count={available.length}
            >
              {available.map((model) => (
                <ModelRow
                  key={model.ID}
                  model={model}
                  active={false}
                  onSelect={() => select(model)}
                />
              ))}
            </Section>
          </div>
        </ScrollArea>
      </div>
    </TooltipProvider>
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

function MachineChip({ machine }: { machine: main.MachineInfo }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] whitespace-nowrap dark:text-white/70 text-slate-900/70">
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

type ModelProps = {
  model: main.SpeechModel;
  active: boolean;
  onSelect: () => void;
};

function SuggestedCard({ model, active, onSelect }: ModelProps) {
  return (
    <div className="rounded-lg border border-primary/40 bg-primary/5 p-3 flex flex-col gap-2">
      <div className="flex items-center gap-1.5 text-[10px] uppercase tracking-wider text-primary">
        <Sparkles className="h-3 w-3" />
        Suggested for this computer
      </div>
      <ModelBody model={model} active={active} onSelect={onSelect} />
    </div>
  );
}

function ModelRow({ model, active, onSelect }: ModelProps) {
  return (
    <li
      className={cn(
        "rounded-md border p-3 transition-colors",
        active && "border-primary bg-primary/5"
      )}
    >
      <ModelBody model={model} active={active} onSelect={onSelect} />
    </li>
  );
}

function ModelBody({ model, active, onSelect }: ModelProps) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-start gap-3">
        <div className="flex flex-col gap-1 min-w-0 flex-1">
          <div className="flex items-center gap-1.5 flex-wrap">
            <span className="text-sm font-medium">{model.Name}</span>
            {active && <Tag tone="green">Active</Tag>}
            {model.Streaming && (
              <Tag tone="muted" icon={<Radio className="h-2.5 w-2.5" />}>
                Live
              </Tag>
            )}
            {model.LanguageDetect && <Tag tone="muted">Auto-detect</Tag>}
            {model.Custom && <Tag tone="muted">Your file</Tag>}
          </div>
          {model.Description && (
            <p className="text-xs dark:text-white/55 text-slate-900/55">
              {model.Description}
            </p>
          )}
        </div>
        <Actions model={model} active={active} onSelect={onSelect} />
      </div>

      <Facts model={model} />
    </div>
  );
}

function Facts({ model }: { model: main.SpeechModel }) {
  return (
    <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] dark:text-white/50 text-slate-900/50">
      <Fact icon={<HardDrive className="h-3 w-3" />}>
        {formatSize(model.SizeMB)}
      </Fact>
      <Fact icon={<Languages className="h-3 w-3" />}>
        {languagesSummary(model)}
      </Fact>
      {model.Accuracy > 0 && (
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="inline-flex items-center gap-1">
              <Gauge className="h-3 w-3" />
              <Meter value={model.Accuracy} />
            </span>
          </TooltipTrigger>
          <TooltipContent className="text-xs">
            Accuracy score {model.Accuracy} / 100
          </TooltipContent>
        </Tooltip>
      )}
      <FitFact model={model} />
    </div>
  );
}

function Fact({
  icon,
  children,
}: {
  icon: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <span className="inline-flex items-center gap-1 whitespace-nowrap">
      {icon}
      {children}
    </span>
  );
}

function Meter({ value }: { value: number }) {
  const filled = Math.round((value / 100) * 5);
  return (
    <span className="inline-flex items-center gap-0.5" aria-label={`Accuracy ${value}`}>
      {Array.from({ length: 5 }, (_, i) => (
        <span
          key={i}
          className={cn(
            "h-1.5 w-1.5 rounded-full",
            i < filled ? "bg-primary" : "bg-primary/20"
          )}
        />
      ))}
    </span>
  );
}

function FitFact({ model }: { model: main.SpeechModel }) {
  const tone =
    model.Fit === "too-large"
      ? "text-destructive"
      : model.Fit === "slow"
        ? "text-amber-600 dark:text-amber-400"
        : model.Fit === "unknown"
          ? ""
          : "text-green-600 dark:text-green-400";

  const label =
    model.Speed > 0 && model.Fit === "comfortable"
      ? `${model.Speed}× realtime here`
      : model.FitLabel;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className={cn("inline-flex items-center gap-1 whitespace-nowrap", tone)}>
          <Zap className="h-3 w-3" />
          {label}
        </span>
      </TooltipTrigger>
      <TooltipContent className="text-xs">{model.FitLabel}</TooltipContent>
    </Tooltip>
  );
}

function Tag({
  tone,
  icon,
  children,
}: {
  tone: "green" | "muted";
  icon?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] uppercase tracking-wide",
        tone === "green" &&
          "bg-green-500/15 text-green-700 dark:text-green-400",
        tone === "muted" && "bg-sidebar-accent dark:text-white/60 text-slate-900/60"
      )}
    >
      {icon}
      {children}
    </span>
  );
}

function Actions({ model, active, onSelect }: ModelProps) {
  const store = useSpeechModelsStore();
  const progress = store.progress[model.ID];
  const downloading = !model.Downloaded && (model.Downloading || !!progress);

  const pct =
    progress && progress.Total
      ? Math.round((progress.Downloaded / progress.Total) * 100)
      : 0;

  const download = () =>
    store.downloadModel(model.ID).catch((err) => toast.error(String(err)));
  const remove = () =>
    store.deleteModel(model.ID).catch((err) => toast.error(String(err)));

  if (model.Downloaded) {
    return (
      <div className="flex items-center gap-1 shrink-0">
        {active ? (
          <span className="inline-flex h-8 items-center gap-1 px-2 text-xs text-primary">
            <Check className="h-4 w-4" />
            In use
          </span>
        ) : (
          <Button size="sm" variant="outline" className="h-8" onClick={onSelect}>
            Use
          </Button>
        )}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              size="icon"
              variant="ghost"
              className="h-8 w-8"
              onClick={remove}
              disabled={active}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent className="text-xs">
            {active ? "Choose another model before deleting" : "Delete from disk"}
          </TooltipContent>
        </Tooltip>
      </div>
    );
  }

  if (downloading) {
    return (
      <div className="flex items-center gap-2 shrink-0 w-36">
        {progress?.Stage === "verifying" ? (
          <span className="text-xs flex items-center gap-1">
            <Loader2 className="h-3 w-3 animate-spin" />
            Verifying
          </span>
        ) : (
          <>
            <Progress value={pct} className="h-1.5 flex-1" />
            <span className="text-xs tabular-nums w-8 text-right">{pct}%</span>
          </>
        )}
        <Button
          size="icon"
          variant="ghost"
          className="h-8 w-8"
          title="Cancel download"
          onClick={() => store.cancelDownload(model.ID)}
        >
          <X className="h-4 w-4" />
        </Button>
      </div>
    );
  }

  const tooLarge = model.Fit === "too-large";
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="shrink-0">
          <Button
            size="sm"
            variant="secondary"
            className="h-8"
            onClick={download}
            disabled={tooLarge}
          >
            <Download className="h-4 w-4" />
            Download
          </Button>
        </span>
      </TooltipTrigger>
      {tooLarge && (
        <TooltipContent className="text-xs">
          Needs more memory than this computer has to spare.
        </TooltipContent>
      )}
    </Tooltip>
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

function languagesSummary(model: main.SpeechModel) {
  const languages = model.Languages || [];
  if (languages.length === 0) return "Languages unknown";
  if (languages.length === 1) return languages[0].Name;
  if (languages.length <= 3) return languages.map((l) => l.Name).join(", ");
  return `${languages.length} languages`;
}

function formatSize(mb: number) {
  return mb >= 1000 ? `${(mb / 1024).toFixed(1)} GB` : `${Math.round(mb)} MB`;
}

function formatMemory(mb: number) {
  if (mb <= 0) return "memory unknown";
  return `${Math.round(mb / 1024)} GB RAM`;
}
