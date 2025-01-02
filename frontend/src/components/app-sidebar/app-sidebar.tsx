import * as React from "react";
import { Trash, Plus, RotateCw } from "lucide-react";

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
  useSidebar,
} from "@/components/ui/sidebar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useNotionPagesStore } from "@/store/notion-pages";
import { cn } from "@/lib/utils";
import { Separator } from "../ui/separator";
import { useLocalPagesStore } from "@/store/local-pages";
import { useActivePageStore } from "@/store/active-page";
import { repository } from "~wails/models";
import { NotionSimplePage } from "@/types";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { setOpenMobile } = useSidebar();

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
      viewOnly: false,
    });

    setOpenMobile(false);
  };

  const onLocalPageClick = (page: repository.Page) => {
    activePageStore.setActivePage({
      __typename: "local_page",
      viewOnly: false,
      page,
    });

    activePageStore.fetchPageBlocks();
    setOpenMobile(false);
  };

  const onNotionPageClick = (page: NotionSimplePage) => {
    activePageStore.setActivePage({
      __typename: "notion_page",
      viewOnly: true,
      page,
    });

    activePageStore.fetchPageBlocks();
    setOpenMobile(false);
  };

  const refreshNotionPages = () => {
    notionPagesStore.getPages();
    if (activePage?.__typename === "notion_page") {
      activePageStore.fetchPageBlocks();
    }
  };

  // Refresh notion pages when the user is online
  React.useEffect(() => {
    addEventListener("online", refreshNotionPages);

    return () => {
      removeEventListener("online", refreshNotionPages);
    };
  }, []);

  return (
    <Sidebar {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <Settings />
          </SidebarMenuItem>
        </SidebarMenu>

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
                          "whitespace-nowrap text-ellipsis group-hover/local:pr-10 leading-3"
                        )}
                      >
                        <span className="align-middle">{item.title}</span>

                        <DeletePageButton page={item} />
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

function DeletePageButton({ page }: { page: repository.Page }) {
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
        <a
          className={cn(
            "absolute right-1 top-1 w-5 opacity-0 group-hover/local:opacity-100 cursor-pointer",
            "flex aspect-square items-center justify-center rounded-md text-sidebar-foreground",
            "ring-sidebar-ring hover:text-sidebar-accent-foreground focus-visible:ring-2 transition hover:bg-red-500/20"
          )}
        >
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
