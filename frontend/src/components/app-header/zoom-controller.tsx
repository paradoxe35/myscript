import { Button } from "../ui/button";
import { Minus, Plus, ZoomIn } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { useContentZoomStore } from "@/store/content-zoom";

export function ZoomController() {
  const zoomStore = useContentZoomStore();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="icon"
          className="bg-sidebar-accent hover:bg-sidebar-accent/40"
        >
          <ZoomIn />
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent>
        <DropdownMenuLabel
          onClick={(e) => e.preventDefault()}
          className="flex items-center gap-2"
        >
          <Button
            variant="outline"
            size="sm"
            disabled={!zoomStore.canZoomOut()}
            onClick={zoomStore.zoomOut}
          >
            <Minus />
          </Button>

          <span>{zoomStore.zoom}</span>

          <Button
            disabled={!zoomStore.canZoomIn()}
            onClick={zoomStore.zoomIn}
            variant="outline"
            size="sm"
          >
            <Plus />
          </Button>
        </DropdownMenuLabel>

        <DropdownMenuSeparator />

        <DropdownMenuItem onClick={zoomStore.reset} className="justify-center">
          <span>Reset</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
