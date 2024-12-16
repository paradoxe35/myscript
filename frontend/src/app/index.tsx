import { AppSidebar } from "@/components/app-sidebar/app-sidebar";

import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Content } from "@/components/content";
import { AppHeaderBreadcrumb } from "@/components/app-header/app-header-breadcrumb";
import { ScriptReaderControllers } from "@/components/app-header/script-reader-controllers";
import { ZoomController } from "@/components/app-header/zoom-controller";
import { useActivePageStore } from "@/store/active-page";
import { useZoomStore } from "@/store/zoom";
import { useEffect, useLayoutEffect } from "react";

export default function App() {
  return (
    <>
      <SetZoom />
      <SidebarProvider>
        <AppSidebar />

        <SidebarInset>
          <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />
            <AppHeaderBreadcrumb />

            <RightButtonsHeader />
          </header>

          <Content />
        </SidebarInset>
      </SidebarProvider>
    </>
  );
}

function SetZoom() {
  const zoomStore = useZoomStore();

  useLayoutEffect(() => {
    document.body.style.zoom = `${zoomStore.zoom / 100}`;
  }, [zoomStore.zoom]);

  return null;
}

function RightButtonsHeader() {
  const activePage = useActivePageStore((state) => state.page);

  return (
    <div className="ml-auto flex items-center gap-2">
      {activePage && (
        <>
          <ScriptReaderControllers />
          <Separator orientation="vertical" className="mr-2 h-4" />
        </>
      )}

      <ZoomController />
    </div>
  );
}
