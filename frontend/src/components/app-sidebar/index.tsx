import * as React from "react";

import { SearchForm } from "@/components/app-sidebar/search-form";
import { Settings } from "@/components/settings/settings";
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";
import { LocalPages } from "./local-pages";
import { NotionPages } from "./notion-pages";
import { SidebarItemsProvider } from "./context";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <SidebarItemsProvider>
      <Sidebar {...props}>
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <Settings />
            </SidebarMenuItem>
          </SidebarMenu>

          <SearchForm />
        </SidebarHeader>

        <SidebarContent className="select-none overflow-x-hidden">
          {/* Local pages */}
          <LocalPages />

          {/* Notion pages */}
          <NotionPages />
        </SidebarContent>
        <SidebarRail />
      </Sidebar>
    </SidebarItemsProvider>
  );
}
