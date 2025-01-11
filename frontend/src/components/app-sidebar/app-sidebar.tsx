import * as React from "react";
import {
  Plus,
  FolderPlusIcon,
  RotateCw,
  FolderIcon,
  FileTextIcon,
  ChevronRight,
} from "lucide-react";

import { SearchForm } from "@/components/app-sidebar/search-form";
import { Settings } from "@/components/settings/settings";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupAction,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";
import { cn } from "@/lib/utils";
import { Separator } from "../ui/separator";
import { useSidebarItems } from "./use-sidebar-items";
import { NewFolderModal } from "./new-folder-modal";
import { DeletePageButton } from "./delete-page-button";
import { MoreOptionButton } from "./more-option-button";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const {
    search,
    setSearch,

    togglePageExpanded,
    createNewPage,
    localPages,
    notionPages,
    activePage,
    onLocalPageClick,
    refreshNotionPages,
    onNotionPageClick,
  } = useSidebarItems();

  return (
    <Sidebar {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <Settings />
          </SidebarMenuItem>
        </SidebarMenu>

        <SearchForm
          inputProps={{
            value: search,
            onChange: (e) => setSearch(e.target.value),
          }}
        />
      </SidebarHeader>

      <SidebarContent>
        {/* Local pages */}
        <SidebarGroup>
          <div className="group/local-add px-0 justify-between transition cursor-default flex">
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
                          "whitespace-nowrap text-ellipsis group-hover/local:pr-10 leading-3"
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

        {/* Notion pages */}
        {notionPages.length > 0 && (
          <>
            <Separator />

            <SidebarGroup>
              <div className="group/notion-refresh px-0 justify-between transition cursor-default">
                <SidebarGroupLabel>{"Notion pages"}</SidebarGroupLabel>

                <SidebarGroupAction
                  className="opacity-0 group-hover/notion-refresh:opacity-100 transition hover:bg-white/10"
                  onClick={refreshNotionPages}
                >
                  <RotateCw />
                </SidebarGroupAction>
              </div>

              <SidebarGroupContent>
                <SidebarMenu>
                  {notionPages.map((item) => {
                    const active =
                      activePage?.__typename === "notion_page" &&
                      activePage.page.id === item.id;

                    return (
                      <SidebarMenuItem key={item.id} className="w-full">
                        <SidebarMenuButton
                          isActive={active}
                          onClick={() => onNotionPageClick(item)}
                          className={cn(
                            "block max-w-full overflow-hidden text-sidebar-foreground/70 font-medium",
                            "whitespace-nowrap text-ellipsis leading-3"
                          )}
                        >
                          {item.title}
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    );
                  })}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </>
        )}
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  );
}
