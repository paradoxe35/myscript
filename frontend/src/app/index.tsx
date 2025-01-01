import { AppSidebar } from "@/components/app-sidebar/app-sidebar";

import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Content } from "@/components/content/content";
import { AppHeaderBreadcrumb } from "@/components/app-header/app-header-breadcrumb";
import { RightButtonsHeader } from "@/components/app-header/right-buttons-header";

export default function App() {
  return (
    <SidebarProvider>
      <AppSidebar />

      <SidebarInset>
        <header className="flex sticky top-0 bg-background h-16 shrink-0 items-center gap-2 border-b px-4 z-20">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <AppHeaderBreadcrumb />
          <RightButtonsHeader />
        </header>

        <Content />
      </SidebarInset>
    </SidebarProvider>
  );
}
