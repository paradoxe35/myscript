import { cn } from "@/lib/utils";
import {
  FolderPlusIcon,
  Plus,
  FolderIcon,
  FileTextIcon,
  ChevronRight,
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
import {
  DragDropContext,
  Draggable,
  DraggableProvided,
  DraggableStateSnapshot,
  Droppable,
} from "@hello-pangea/dnd";
import { PropsWithChildren } from "react";
import { AnimatePresence, motion } from "motion/react";
import { repository } from "~wails/models";

export function LocalPages() {
  const { createNewPage, localPages, activePage, reorderLocalPages } =
    useSidebarItemsContext();

  const handleDragEnd = (result: { source: any; destination: any }) => {
    const { source, destination } = result;

    // Drop outside the list or no movement
    if (!destination || source.index === destination.index) {
      return;
    }

    reorderLocalPages(source.index, destination.index);
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

      <DragDropContext onDragEnd={handleDragEnd}>
        <Droppable droppableId="droppable-id">
          {(provided) => {
            return (
              <SidebarGroupContent
                ref={provided.innerRef}
                {...provided.droppableProps}
              >
                <SidebarMenu>
                  {/* Animated items */}
                  <AnimatePresence>
                    {localPages.map((item, index) => {
                      return (
                        <Draggable
                          key={item.ID}
                          draggableId={item.ID.toString()}
                          index={index}
                        >
                          {(provided, snapshot) => {
                            return (
                              <PageItem
                                item={item}
                                provided={provided}
                                snapshot={snapshot}
                              />
                            );
                          }}
                        </Draggable>
                      );
                    })}
                  </AnimatePresence>

                  {provided.placeholder}
                </SidebarMenu>
              </SidebarGroupContent>
            );
          }}
        </Droppable>
      </DragDropContext>
    </SidebarGroup>
  );
}

function PageItem({
  provided,
  snapshot,
  item,
}: PropsWithChildren<{
  item: repository.Page;
  snapshot: DraggableStateSnapshot;
  provided: DraggableProvided;
}>) {
  const { togglePageExpanded, onLocalPageClick, activePage } =
    useSidebarItemsContext();

  const active =
    activePage?.__typename === "local_page" && activePage.page.ID === item.ID;

  const pageId = item.ID;

  const Icon = item.is_folder ? FolderIcon : FileTextIcon;

  return (
    <SidebarMenuItem
      key={pageId}
      className={cn(snapshot.isDragging && "opacity-40")}
      ref={provided.innerRef}
      {...provided.draggableProps}
    >
      <motion.div
        key={pageId}
        initial={{ opacity: 0, y: 10, scale: 0.98 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        exit={{
          opacity: 0,
          scale: 0.95,
          transition: {
            opacity: { duration: 0.15, ease: "easeOut" },
            scale: { duration: 0.2, ease: "easeOut" },
          },
        }}
      >
        <div className="group/local">
          <SidebarMenuButton
            isActive={active}
            onClick={() => {
              // Clickable for pages
              !item.is_folder && onLocalPageClick(item);

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
            {...provided.dragHandleProps}
          >
            <div className="w-full">
              <span className="align-middle">
                <Icon
                  className={cn(
                    "mr-1 h-4 w-4 inline-block -mt-[3.5px]",
                    item.is_folder && "group-hover/local:hidden"
                  )}
                />

                <ChevronRight
                  className={cn(
                    "mr-1 h-4 w-4 hidden -mt-[3.5px] transition-transform",
                    item.expanded && "rotate-90",
                    item.is_folder && "group-hover/local:inline-block"
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
        </div>
      </motion.div>
    </SidebarMenuItem>
  );
}
