import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { APP_NAME } from "@/lib/constants";
import { cn } from "@/lib/utils";
import { Cloud, FileText, Mic, Settings2, Sparkles } from "lucide-react";
import { PropsWithChildren, useState } from "react";
import { useSettings } from "./context";
import { AISection } from "./sections/ai-section";
import { BackupSection } from "./sections/backup-section";
import { GeneralSection } from "./sections/general-section";
import { NotionSection } from "./sections/notion-section";
import { SpeechSection } from "./sections/speech-section";

type SectionKey = "general" | "ai" | "notion" | "speech" | "backup";

type Section = {
  key: SectionKey;
  label: string;
  description: string;
  icon: React.ComponentType<{ className?: string }>;
  render: () => React.ReactNode;
};

const SECTIONS: Section[] = [
  {
    key: "speech",
    label: "Speech",
    description: "Transcription engine and models",
    icon: Mic,
    render: () => <SpeechSection />,
  },
  {
    key: "ai",
    label: "AI",
    description: "Providers for the writing assistant",
    icon: Sparkles,
    render: () => <AISection />,
  },
  {
    key: "notion",
    label: "Notion",
    description: "Use your Notion pages as scripts",
    icon: FileText,
    render: () => <NotionSection />,
  },
  {
    key: "backup",
    label: "Backup",
    description: "Sync your scripts with Google Drive",
    icon: Cloud,
    render: () => <BackupSection />,
  },
  {
    key: "general",
    label: "General",
    description: "Appearance and version",
    icon: Settings2,
    render: () => <GeneralSection />,
  },
];

export function SettingsModal(props: PropsWithChildren) {
  const { cloud } = useSettings();
  const [current, setCurrent] = useState<SectionKey>("speech");

  const sections = SECTIONS.filter(
    (section) => section.key !== "backup" || cloud.cloudEnabled,
  );
  const section = sections.find((item) => item.key === current) ?? sections[0];

  return (
    <Dialog>
      <DialogTrigger asChild>{props.children}</DialogTrigger>

      <DialogContent className="max-w-[860px] gap-0 overflow-hidden p-0 sm:max-w-[860px]">
        <div className="flex h-[560px]">
          <nav className="flex w-52 shrink-0 flex-col gap-1 border-r bg-muted/30 p-3">
            <DialogTitle className="px-2 pb-2 pt-1 text-sm font-semibold">
              {APP_NAME} settings
            </DialogTitle>

            {sections.map((item) => (
              <button
                key={item.key}
                type="button"
                onClick={() => setCurrent(item.key)}
                className={cn(
                  "flex items-center gap-2.5 rounded-md px-2.5 py-2 text-sm transition",
                  "hover:bg-accent",
                  item.key === section.key &&
                    "bg-accent font-medium text-accent-foreground",
                )}
              >
                <item.icon className="h-4 w-4 shrink-0 text-muted-foreground" />
                {item.label}
              </button>
            ))}
          </nav>

          <div className="flex min-h-0 min-w-0 flex-1 flex-col">
            <header className="flex flex-col gap-0.5 border-b px-6 py-4">
              <h2 className="text-base font-semibold">{section.label}</h2>
              <DialogDescription className="text-xs">
                {section.description}
              </DialogDescription>
            </header>

            <ScrollArea className="min-h-0 flex-1">
              <div className="px-6 py-5">{section.render()}</div>
            </ScrollArea>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
