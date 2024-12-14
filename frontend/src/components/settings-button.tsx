import * as React from "react";
import { Settings, FileMinus2 } from "lucide-react";

import { SidebarMenu, SidebarMenuItem } from "@/components/ui/sidebar";
import { APP_NAME } from "@/lib/constants";
import { cn } from "@/lib/utils";
import { Button } from "./ui/button";
import { SettingsModal } from "./settings-modal";

export function SettingsButton() {
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <div className="flex items-center gap-2 ml-2 mb-3">
          <div
            className={cn(
              "flex aspect-square size-8 items-center justify-center",
              "rounded-lg bg-sidebar-primary text-sidebar-primary-foreground"
            )}
          >
            <FileMinus2 className="size-4" />
          </div>

          <div className="flex flex-col gap-0.5 leading-none">
            <span className="font-semibold">{APP_NAME}</span>
          </div>

          <SettingsModal>
            <Button
              size="sm"
              variant="ghost"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground ml-auto"
            >
              <Settings />
            </Button>
          </SettingsModal>
        </div>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
