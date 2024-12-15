import * as React from "react";
import { Plus } from "lucide-react";

import { SearchForm } from "@/components/search-form";
import { SettingsButton } from "@/components/settings-button";
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
import { useNotionPagesStore } from "@/store/notion-pages";
import { cn } from "@/lib/utils";
import { Separator } from "./ui/separator";
import { useLocalPagesStore } from "@/store/local-pages";
import { useActivePageStore } from "@/store/active-page";
import { repository } from "~wails/models";
import { NotionSimplePage } from "@/types";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const activePageStore = useActivePageStore();
  const localPagesStore = useLocalPagesStore();
  const notionPagesStore = useNotionPagesStore();

  const activePage = activePageStore.page;

  React.useEffect(() => {
    localPagesStore.getPages();
  }, []);

  const notionPages = React.useMemo(() => {
    return notionPagesStore.getSimplifiedPages();
  }, [notionPagesStore.pages]);

  const createNewPage = async () => {
    const newPage = await localPagesStore.newPage();

    activePageStore.setActivePage({
      __typename: "local_page",
      page: newPage,
      readMode: false,
    });
  };

  const onLocalPageClick = (page: repository.Page) => {
    activePageStore.setActivePage({
      __typename: "local_page",
      readMode: false,
      page,
    });
  };

  const onNotionPageClick = (page: NotionSimplePage) => {
    activePageStore.setActivePage({
      __typename: "notion_page",
      readMode: true,
      page,
    });
  };

  return (
    <Sidebar {...props}>
      <SidebarHeader>
        <SettingsButton />
        <SearchForm />
      </SidebarHeader>

      <SidebarContent>
        {/* Local pages */}
        <SidebarGroup>
          <SidebarMenuButton className="px-0 justify-between group transition">
            <SidebarGroupLabel>{"Local pages"}</SidebarGroupLabel>

            <SidebarGroupAction
              asChild
              className="opacity-0 group-hover:opacity-100 transition-opacity hover:bg-sidebar/60"
              onClick={createNewPage}
            >
              <a href="#">
                <Plus />
              </a>
            </SidebarGroupAction>
          </SidebarMenuButton>

          <SidebarGroupContent>
            <SidebarMenu>
              {localPagesStore.pages.map((item) => {
                const active =
                  activePage?.__typename === "local_page" &&
                  activePage.page.ID === item.ID;

                return (
                  <SidebarMenuItem key={item.ID} className="w-full">
                    <SidebarMenuButton
                      isActive={active}
                      onClick={() => onLocalPageClick(item)}
                      className={cn(
                        "block max-w-full overflow-hidden text-sidebar-foreground/70 font-medium",
                        "whitespace-nowrap text-ellipsis"
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

        {/* Notion pages */}
        {notionPages.length > 0 && (
          <>
            <Separator />

            <SidebarGroup>
              <SidebarGroupLabel>{"Notion pages"}</SidebarGroupLabel>

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
                            "whitespace-nowrap text-ellipsis"
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
