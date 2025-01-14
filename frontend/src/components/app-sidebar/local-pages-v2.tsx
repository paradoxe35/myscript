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
import React, {
  forwardRef,
  PropsWithChildren,
  useEffect,
  useMemo,
} from "react";
import { AnimatePresence, motion } from "motion/react";
import { repository } from "~wails/models";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "../ui/collapsible";

import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";

export function LocalPages() {
  const { createNewPage, localPages, reorderLocalPages } =
    useSidebarItemsContext();

  const adaptedPages = useMemo(() => {
    type AP = { id: number | string; children: AP[] };
    const adaptedPages: AP[] = [];

    const mutate = (page: repository.Page, adaptedPages: AP[]) => {
      const adaptedPage: AP = {
        id: page.ID,
        children: [],
      };

      adaptedPages.push(adaptedPage);

      for (const child of page.Children) {
        mutate(child, adaptedPage.children);
      }
    };

    for (const page of localPages) {
      mutate(page, adaptedPages);
    }

    return adaptedPages;
  }, [localPages]);

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  useEffect(() => {
    console.log(adaptedPages);
  }, [adaptedPages]);

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
          {/* Animated items */}
          <AnimatePresence>
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              // onDragStart={handleDragStart}
              // onDragEnd={handleDragEnd}
            >
              <SortableContext
                items={adaptedPages}
                strategy={verticalListSortingStrategy}
              >
                {localPages.map((item, index) => {
                  return <PageItem key={item.ID} item={item} />;
                })}
              </SortableContext>
            </DndContext>
          </AnimatePresence>
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}

function PageItem({
  item,
  subMenu,
}: PropsWithChildren<{
  item: repository.Page;
  subMenu?: boolean;
}>) {
  const { togglePageExpanded, onLocalPageClick, activePage } =
    useSidebarItemsContext();

  const { attributes, listeners, setNodeRef, transform, transition } =
    useSortable({ id: item.ID });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const active =
    activePage?.__typename === "local_page" && activePage.page.ID === item.ID;

  const pageId = item.ID;

  const Icon = item.is_folder ? FolderIcon : FileTextIcon;

  const children = item.Children || [];

  const button = (
    <SidebarMenuButton
      isActive={active}
      onClick={() => {
        // Clickable for pages
        !item.is_folder && onLocalPageClick(item);

        console.log("onLocalPageClick", item);

        // Toggle expanded state for folder
        item.is_folder && togglePageExpanded(item);
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
              item.is_folder && item.expanded && "hidden"
            )}
          />

          <FolderOpen
            className={cn(
              "mr-1 h-4 w-4 hidden -mt-[3.5px] transition-transform",
              item.expanded && item.is_folder && ["inline-block"]
            )}
          />

          <span>{item.title}</span>
        </span>

        {item.is_folder ? (
          <MoreOptionButton page={item} />
        ) : (
          <DeletePageButton page={item} />
        )}
      </div>
    </SidebarMenuButton>
  );

  return (
    <SidebarMenuItem
      key={pageId}
      className={cn(subMenu && "pl-3")}
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
    >
      <motion.div
        key={pageId}
        exit={{
          opacity: 0,
          scale: 0.95,
          transition: {
            opacity: { duration: 0.15, ease: "easeOut" },
            scale: { duration: 0.2, ease: "easeOut" },
          },
        }}
      >
        {!item.is_folder && <div className="group/local">{button}</div>}

        {item.is_folder && (
          <Collapsible open={item.expanded}>
            <CollapsibleTrigger asChild>
              <div className="group/local">{button}</div>
            </CollapsibleTrigger>

            <CollapsibleContent
              className={cn("pt-1", children.length === 0 && "py-2")}
            >
              {children.map((subItem, index) => {
                return <PageItem key={item.ID} item={subItem} subMenu={true} />;
              })}
            </CollapsibleContent>
          </Collapsible>
        )}
      </motion.div>
    </SidebarMenuItem>
  );
}

const Item = forwardRef<HTMLDivElement, React.ComponentProps<"div">>(
  ({ id, ...props }, ref) => {
    return (
      <div {...props} ref={ref}>
        {id}
      </div>
    );
  }
);
