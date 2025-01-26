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
import { useState } from "react";

export const dropdownButtonClassName = cn(
  "absolute right-1 top-1 w-5 opacity-0 group-hover/local:opacity-100 cursor-pointer",
  "flex aspect-square items-center justify-center rounded-md text-sidebar-foreground",
  "ring-sidebar-ring hover:text-sidebar-accent-foreground focus-visible:ring-2 transition hover:bg-red-500/20"
);

export function DeletePageButton({ page }: { page: repository.Page }) {
  const [menuOpen, setMenuOpen] = useState(false);

  const activePageStore = useActivePageStore();
  const localPagesStore = useLocalPagesStore();

  const handleDeletePage = async (e: any) => {
    e?.stopPropagation();

    const activePage = activePageStore.page;

    const pageChildren = localPagesStore.pages.filter(
      (p) => p.ParentID === page.ID
    );

    await Promise.all(
      pageChildren.map((p) => {
        p.ParentID = page.ParentID;
        return localPagesStore.saveNewPageOrder(p);
      })
    );

    localPagesStore.deletePage(page.ID).then(() => {
      if (
        activePage?.__typename === "local_page" &&
        activePage.page.ID === page.ID
      ) {
        activePageStore.unsetActivePage();
      }
    });
  };

  return (
    <DropdownMenu open={menuOpen} onOpenChange={setMenuOpen}>
      <DropdownMenuTrigger asChild>
        <a
          onClick={(e) => e.stopPropagation()}
          className={cn(
            dropdownButtonClassName,
            menuOpen && "opacity-100 dropdown-menu-open bg-red-500/20"
          )}
        >
          <Trash className="text-red-500/80 w-4" />
        </a>
      </DropdownMenuTrigger>

      <DropdownMenuContent onClick={(e) => e.stopPropagation()}>
        <DropdownMenuItem className="text-red-300" onClick={handleDeletePage}>
          Delete
        </DropdownMenuItem>

        <DropdownMenuItem>Cancel</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
