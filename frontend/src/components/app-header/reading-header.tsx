import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { useActivePageStore } from "@/store/active-page";
import { useContentReadStore } from "@/store/content-read";
import { useTranscriberStore } from "@/store/transcriber";
import { X } from "lucide-react";
import { ScriptReaderControllers } from "./script-reader-controllers";
import { ZoomController } from "./zoom-controller";

export function ReadingHeader() {
  const title = useActivePageStore((store) => store.page?.page.title);
  const setReadMode = useActivePageStore((store) => store.setReadMode);
  const stopRecording = useTranscriberStore((store) => store.stopRecording);
  const position = useContentReadStore((state) => state.position);
  const total = useContentReadStore((state) => state.total);

  const percent = total ? Math.round((position / total) * 100) : 0;

  const exit = () => {
    stopRecording();
    setReadMode(false);
  };

  return (
    <>
      <span className="truncate text-sm text-muted-foreground">
        {title}
      </span>

      <div className="ml-auto flex items-center gap-3">
        <ScriptReaderControllers />

        <div className="flex w-44 flex-col gap-1">
          <div className="flex justify-between text-xs tabular-nums text-muted-foreground">
            <span>
              {position} / {total} words
            </span>
            <span>{percent}%</span>
          </div>
          <Progress value={percent} className="h-1" />
        </div>

        <Separator orientation="vertical" className="h-4" />
        <ZoomController />

        <Button variant="ghost" size="icon" title="Exit reading" onClick={exit}>
          <X />
        </Button>
      </div>
    </>
  );
}
