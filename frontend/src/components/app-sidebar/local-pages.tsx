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
  OnDragEndResponder,
} from "@hello-pangea/dnd";
import { PropsWithChildren } from "react";
import { AnimatePresence, motion } from "motion/react";
import { repository } from "~wails/models";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "../ui/collapsible";

export function LocalPages() {
  const { createNewPage, localPages, reorderLocalPages } =
    useSidebarItemsContext();

  const handleDragEnd: OnDragEndResponder<string> = (result, provided) => {
    reorderLocalPages(result, provided);
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
        <Droppable droppableId="root" type="droppable-item">
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
  subMenu,
}: PropsWithChildren<{
  item: repository.Page;
  snapshot: DraggableStateSnapshot;
  provided: DraggableProvided;
  subMenu?: boolean;
}>) {
  const { togglePageExpanded, onLocalPageClick, activePage } =
    useSidebarItemsContext();

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
              item.is_folder && item.expanded && "hidden",
              item.is_folder && "group-hover/local:hidden"
            )}
          />

          <ChevronRight
            className={cn(
              "mr-1 h-4 w-4 hidden -mt-[3.5px] transition-transform",
              item.expanded && "rotate-90",
              item.is_folder && "group-hover/local:inline-block",
              item.expanded && item.is_folder && ["inline-block", "rotate-90"]
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
      className={cn(snapshot.isDragging && "opacity-40", subMenu && "pl-3")}
      ref={provided.innerRef}
      {...provided.draggableProps}
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
          <Collapsible open={item.expanded && !snapshot.isDragging}>
            <CollapsibleTrigger asChild>
              <div className="group/local">{button}</div>
            </CollapsibleTrigger>

            <Droppable
              key={item.ID}
              droppableId={item.ID.toString()}
              type="droppable-item"
            >
              {(provided) => {
                return (
                  <CollapsibleContent
                    ref={provided.innerRef}
                    className={cn("pt-1", children.length === 0 && "py-2")}
                    {...provided.droppableProps}
                  >
                    {children.map((subItem, index) => {
                      return (
                        <Draggable
                          key={subItem.ID}
                          draggableId={subItem.ID.toString()}
                          index={index}
                        >
                          {(provided, snapshot) => {
                            return (
                              <PageItem
                                item={subItem}
                                provided={provided}
                                snapshot={snapshot}
                                subMenu={true}
                              />
                            );
                          }}
                        </Draggable>
                      );
                    })}

                    {provided.placeholder}
                  </CollapsibleContent>
                );
              }}
            </Droppable>
          </Collapsible>
        )}
      </motion.div>
    </SidebarMenuItem>
  );
}
