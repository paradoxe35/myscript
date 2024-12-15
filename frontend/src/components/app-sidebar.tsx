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

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const notionPagesStore = useNotionPagesStore();

  const notionPages = React.useMemo(() => {
    return notionPagesStore.getSimplifiedPages();
  }, [notionPagesStore.pages]);

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
            >
              <a href="#">
                <Plus />
              </a>
            </SidebarGroupAction>
          </SidebarMenuButton>

          {/* <SidebarGroupContent>
            <SidebarMenu>
              {item.items.map((item) => (
                <SidebarMenuItem>
                  <SidebarMenuButton asChild isActive={item.isActive}>
                    <a href={item.url}>{item.title}</a>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent> */}
        </SidebarGroup>

        {/* Notion pages */}
        {notionPages.length > 0 && (
          <>
            <Separator />

            <SidebarGroup>
              <SidebarGroupLabel>{"Notion pages"}</SidebarGroupLabel>

              <SidebarGroupContent>
                <SidebarMenu>
                  {notionPages.map((item) => (
                    <SidebarMenuItem key={item.id} className="w-full">
                      <SidebarMenuButton
                        className={cn(
                          "block max-w-full overflow-hidden text-sidebar-foreground/70 font-medium",
                          "whitespace-nowrap text-ellipsis"
                        )}
                      >
                        {item.title}
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  ))}
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
