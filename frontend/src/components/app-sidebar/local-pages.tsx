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

export function LocalPages() {
  const {
    createNewPage,
    localPages,
    togglePageExpanded,
    onLocalPageClick,
    activePage,
  } = useSidebarItemsContext();

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
          {localPages.map((item) => {
            const active =
              activePage?.__typename === "local_page" &&
              activePage.page.ID === item.ID;

            const Icon = item.is_folder ? FolderIcon : FileTextIcon;

            return (
              <SidebarMenuItem key={item.ID} className="w-full">
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
                      "group-hover/local:pr-10 has-[.dropdown-menu-open]:pr-10"
                    )}
                  >
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
                  </SidebarMenuButton>
                </div>
              </SidebarMenuItem>
            );
          })}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
