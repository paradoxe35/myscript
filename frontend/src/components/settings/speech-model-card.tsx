import {
  Check,
  Download,
  Gauge,
  HardDrive,
  Languages,
  Loader2,
  Radio,
  Trash2,
  X,
  Zap,
} from "lucide-react";
import { toast } from "sonner";

import { cn } from "@/lib/utils";
import { Button } from "../ui/button";
import { Progress } from "../ui/progress";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import { useSpeechModelsStore } from "@/store/speech-models";
import { main } from "~wails/models";

export type ModelProps = {
  model: main.SpeechModel;
  active: boolean;
  onSelect: () => void;
};

export function ModelBody({ model, active, onSelect }: ModelProps) {
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
    <span
      className="inline-flex items-center gap-0.5"
      aria-label={`Accuracy ${value}`}
    >
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
        <span
          className={cn("inline-flex items-center gap-1 whitespace-nowrap", tone)}
        >
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
        tone === "muted" &&
          "bg-sidebar-accent dark:text-white/60 text-slate-900/60"
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
