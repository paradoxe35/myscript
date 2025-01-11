import { repository } from "~wails/models";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { Ellipsis } from "lucide-react";
import { dropdownButtonClassName } from "./delete-page-button";
import { RenameFolderModal } from "./rename-folder-modal";
import { useState } from "react";

export function MoreOptionButton({ page }: { page: repository.Page }) {
  const [renameDialogOpen, setRenameDialogOpen] = useState(false);

  return (
    <div onClick={(e) => e.stopPropagation()}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <a
            className={cn(
              dropdownButtonClassName,
              "dark:hover:bg-white/10 hover:bg-slate-900/10"
            )}
          >
            <Ellipsis className="w-4" />
          </a>
        </DropdownMenuTrigger>

        <DropdownMenuContent>
          <DropdownMenuItem onClick={(e) => setRenameDialogOpen(true)}>
            Rename
          </DropdownMenuItem>

          <DropdownMenuItem className="text-red-300">Delete</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <RenameFolderModal
        page={page}
        open={renameDialogOpen}
        onOpenChange={setRenameDialogOpen}
      />
    </div>
  );
}
