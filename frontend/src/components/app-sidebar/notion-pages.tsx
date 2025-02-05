import { cn } from "@/lib/utils";
import { RotateCw } from "lucide-react";
import { Separator } from "../ui/separator";
import {
  SidebarGroup,
  SidebarGroupAction,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "../ui/sidebar";
import { useSidebarItemsContext } from "./context";

export function NotionPages() {
  const { notionPages, activePage, onNotionPageClick, refreshNotionPages } =
    useSidebarItemsContext();

  if (notionPages.length < 1) {
    return null;
  }

  return (
    <>
      <Separator />

      <SidebarGroup>
        <div className="group/notion-refresh px-0 justify-between transition flex cursor-default mb-1">
          <SidebarGroupLabel>{"Notion pages"}</SidebarGroupLabel>

          <SidebarGroupLabel
            className="opacity-0 group-hover/notion-refresh:opacity-100 transition dark:hover:bg-white/10 hover:bg-slate-900/10"
            onClick={refreshNotionPages}
          >
            <RotateCw />
          </SidebarGroupLabel>
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
  );
}
