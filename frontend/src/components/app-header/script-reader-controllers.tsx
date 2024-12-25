import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { Button } from "../ui/button";
import { BookOpenText, Play } from "lucide-react";
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
                readMode && "bg-red-300/40 hover:bg-red-300/60"
              )}
            >
              {readMode ? <BookOpenText /> : <Play />}
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
