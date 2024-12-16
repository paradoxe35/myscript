import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { Button } from "../ui/button";
import { Play } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../ui/tooltip";

export function ScriptReaderControllers(props: React.ComponentProps<"div">) {
  const activePage = useActivePageStore((state) => state.page);

  if (!activePage) return null;

  return (
    <div {...props} className={cn("flex gap-2 items-center", props.className)}>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              className="bg-sidebar-accent hover:bg-sidebar-accent/40"
            >
              <Play />
            </Button>
          </TooltipTrigger>

          <TooltipContent>
            <p>Start reading</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
}
