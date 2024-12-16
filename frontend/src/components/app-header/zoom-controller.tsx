import { cn } from "@/lib/utils";
import { Button } from "../ui/button";
import { ZoomIn } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../ui/tooltip";

export function ZoomController() {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="outline"
            size="icon"
            className="bg-sidebar-accent hover:bg-sidebar-accent/40"
          >
            <ZoomIn />
          </Button>
        </TooltipTrigger>

        <TooltipContent>
          <p>Zoom in or out</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
