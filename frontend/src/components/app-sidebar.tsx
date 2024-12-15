import * as React from "react";
import { Trash, Plus, RotateCw } from "lucide-react";

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

  const refreshNotionPages = () => {
    notionPagesStore.getPages();
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
          <div className="group/local-add px-0 justify-between transition cursor-default">
            <SidebarGroupLabel>{"Local pages"}</SidebarGroupLabel>

            <SidebarGroupAction
              className="opacity-0 group-hover/local-add:opacity-100 transition hover:bg-white/10"
              onClick={createNewPage}
            >
              <Plus />
            </SidebarGroupAction>
          </div>

          <SidebarGroupContent>
            <SidebarMenu>
              {localPagesStore.pages.map((item) => {
                const active =
                  activePage?.__typename === "local_page" &&
                  activePage.page.ID === item.ID;

                return (
                  <SidebarMenuItem key={item.ID} className="w-full">
                    <div className="group/local">
                      <SidebarMenuButton
                        isActive={active}
                        onClick={() => onLocalPageClick(item)}
                        className={cn(
                          "block max-w-full overflow-hidden transition text-sidebar-foreground/70 font-medium",
                          "whitespace-nowrap text-ellipsis group-hover/local:pr-10"
                        )}
                      >
                        {item.title}

                        <SidebarGroupAction
                          asChild
                          className={cn(
                            "opacity-0 group-hover/local:opacity-100 transition hover:bg-red-500/10",
                            "my-auto"
                          )}
                        >
                          <a href="#">
                            <Trash />
                          </a>
                        </SidebarGroupAction>
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
