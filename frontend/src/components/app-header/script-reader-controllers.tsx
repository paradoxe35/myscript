import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { Button } from "../ui/button";
import { CircleStop, Play } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../ui/tooltip";

export function ScriptReaderControllers(props: React.ComponentProps<"div">) {
  const activePageStore = useActivePageStore();
  const activePage = activePageStore.page;
  const readMode = activePageStore.readMode;

  const startReadMode = () => {
    activePageStore.toggleReadMode();
  };

  if (!activePage) return null;
  return (
    <div {...props} className={cn("flex gap-2 items-center", props.className)}>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              onClick={startReadMode}
              className={cn(
                "bg-sidebar-accent hover:bg-sidebar-accent/40",
                readMode && "bg-sidebar-accent/40 hover:bg-sidebar-accent/60"
              )}
            >
              {readMode ? <CircleStop /> : <Play />}
            </Button>
          </TooltipTrigger>

          <TooltipContent>
            {readMode ? <p>Stop reading mode</p> : <p>Start reading mode</p>}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
}
