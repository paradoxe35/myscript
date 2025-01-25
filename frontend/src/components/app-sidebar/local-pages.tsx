import { cn } from "@/lib/utils";
import {
  FolderPlusIcon,
  Plus,
  FolderIcon,
  FileTextIcon,
  FolderOpen,
} from "lucide-react";
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from "../ui/sidebar";
import { useSidebarItemsContext } from "./context";
import { DeletePageButton } from "./delete-page-button";
import { MoreOptionButton } from "./more-option-button";
import { NewFolderModal } from "./new-folder-modal";
import { repository } from "~wails/models";

import Tree, {
  type RenderItemParams,
  type TreeSourcePosition,
  type TreeDestinationPosition,
} from "@atlaskit/tree";

export function LocalPages() {
  const { createNewPage, pagesTree, reorderLocalPages } =
    useSidebarItemsContext();

  const handleDragEnd = (
    sourcePosition: TreeSourcePosition,
    destinationPosition?: TreeDestinationPosition
  ) => {
    reorderLocalPages(sourcePosition, destinationPosition);
  };

  return (
    <SidebarGroup>
      <div className="group/local-add px-0 justify-between transition cursor-default flex mb-1">
        <SidebarGroupLabel>{"Local pages"}</SidebarGroupLabel>

        <div className="flex items-center gap-1">
          <NewFolderModal>
            <SidebarGroupLabel
              className={cn(
                "opacity-0 group-hover/local-add:opacity-100 transition hover:bg-white/10"
              )}
            >
              <FolderPlusIcon />
            </SidebarGroupLabel>
          </NewFolderModal>

          <SidebarGroupLabel
            className="opacity-0 group-hover/local-add:opacity-100 transition hover:bg-white/10"
            onClick={createNewPage}
          >
            <Plus />
          </SidebarGroupLabel>
        </div>
      </div>

      <SidebarGroupContent>
        <SidebarMenu>
          <Tree
            tree={pagesTree}
            renderItem={(item: RenderItemParams) => <PageItem {...item} />}
            offsetPerLevel={12}
            onDragEnd={handleDragEnd}
            isDragEnabled
            isNestingEnabled
          />
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}

function PageItem({ provided, snapshot, item }: RenderItemParams) {
  const { togglePageExpanded, onLocalPageClick, activePage } =
    useSidebarItemsContext();

  const page = item.data as repository.Page;

  const active =
    activePage?.__typename === "local_page" && activePage.page.ID === page.ID;

  const pageId = page.ID;

  const Icon = page.is_folder ? FolderIcon : FileTextIcon;

  const button = (
    <SidebarMenuButton
      isActive={active}
      onClick={() => {
        // Clickable for pages
        !page.is_folder && onLocalPageClick(page);

        // Toggle expanded state for folder
        page.is_folder && togglePageExpanded(page);
      }}
      className={cn(
        "block max-w-full overflow-hidden transition text-sidebar-foreground/70 font-medium",
        "whitespace-nowrap text-ellipsis leading-3",
        "group-hover/local:pr-10 has-[.dropdown-menu-open]:pr-10",
        "!cursor-pointer"
      )}
      asChild
    >
      <div className="w-full">
        <span className="align-middle">
          <Icon
            className={cn(
              "mr-1 h-4 w-4 inline-block -mt-[3.5px]",
              page.is_folder && page.expanded && "hidden"
            )}
          />

          <FolderOpen
            className={cn(
              "mr-1 h-4 w-4 hidden -mt-[3.5px] transition-transform",
              page.expanded && page.is_folder && ["inline-block"]
            )}
          />

          <span>{page.title}</span>
        </span>

        {page.is_folder ? (
          <MoreOptionButton page={page} />
        ) : (
          <DeletePageButton page={page} />
        )}
      </div>
    </SidebarMenuButton>
  );

  return (
    <SidebarMenuItem
      key={pageId}
      className={cn(
        snapshot.isDragging && "opacity-40",
        item.hasChildren && page.expanded && "mb-1"
      )}
      ref={provided.innerRef}
      {...provided.draggableProps}
      {...provided.dragHandleProps}
    >
      <div className="group/local">{button}</div>
    </SidebarMenuItem>
  );
}
