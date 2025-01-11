import { useActivePageStore } from "@/store/active-page";
import { useLocalPagesStore } from "@/store/local-pages";
import { repository } from "~wails/models";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { Trash } from "lucide-react";

export const dropdownButtonClassName = cn(
  "absolute right-1 top-1 w-5 opacity-0 group-hover/local:opacity-100 cursor-pointer",
  "flex aspect-square items-center justify-center rounded-md text-sidebar-foreground",
  "ring-sidebar-ring hover:text-sidebar-accent-foreground focus-visible:ring-2 transition hover:bg-red-500/20"
);

export function DeletePageButton({ page }: { page: repository.Page }) {
  const activePageStore = useActivePageStore();
  const deletePage = useLocalPagesStore((state) => state.deletePage);

  const handleDeletePage = async (
    e: React.MouseEvent<HTMLDivElement, MouseEvent>
  ) => {
    e.stopPropagation();
    const activePage = activePageStore.page;

    deletePage(page.ID).then(() => {
      if (
        activePage?.__typename === "local_page" &&
        activePage.page.ID === page.ID
      ) {
        activePageStore.unsetActivePage();
      }
    });
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <a className={dropdownButtonClassName}>
          <Trash className="text-red-500/80 w-4" />
        </a>
      </DropdownMenuTrigger>

      <DropdownMenuContent>
        <DropdownMenuItem className="text-red-300" onClick={handleDeletePage}>
          Delete
        </DropdownMenuItem>

        <DropdownMenuItem onClick={(e) => e.stopPropagation()}>
          Cancel
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
