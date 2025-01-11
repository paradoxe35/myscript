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
import { useAsyncPromptModal } from "../async-prompt-modal";
import { useLocalPagesStore } from "@/store/local-pages";

export function MoreOptionButton({ page }: { page: repository.Page }) {
  const { prompt } = useAsyncPromptModal();

  const deletePage = useLocalPagesStore((state) => state.deletePage);

  const [menuOpen, setMenuOpen] = useState(false);
  const [renameDialogOpen, setRenameDialogOpen] = useState(false);

  const handleDeletePage = async () => {
    const canDelete = await prompt({
      title: `Delete Folder: ${page.title}`,
      confirmationPrompt: true,
    });

    if (canDelete) {
      deletePage(page.ID);
    }
  };

  return (
    <div onClick={(e) => e.stopPropagation()}>
      <DropdownMenu open={menuOpen} onOpenChange={setMenuOpen}>
        <DropdownMenuTrigger asChild>
          <a
            className={cn(
              dropdownButtonClassName,
              "dark:hover:bg-white/10 hover:bg-slate-900/10",
              menuOpen && [
                "opacity-100 dropdown-menu-open",
                "dark:bg-white/10 bg-slate-900/10",
              ]
            )}
          >
            <Ellipsis className="w-4" />
          </a>
        </DropdownMenuTrigger>

        <DropdownMenuContent>
          <DropdownMenuItem onClick={(e) => setRenameDialogOpen(true)}>
            Rename
          </DropdownMenuItem>

          <DropdownMenuItem className="text-red-300" onClick={handleDeletePage}>
            Delete
          </DropdownMenuItem>
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
