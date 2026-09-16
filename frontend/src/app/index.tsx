import { useState } from "react";
import { AppSidebar } from "@/components/app-sidebar";

import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Content } from "@/components/content/content";
import { AppHeaderBreadcrumb } from "@/components/app-header/app-header-breadcrumb";
import { RightButtonsHeader } from "@/components/app-header/right-buttons-header";
import { ReadingHeader } from "@/components/app-header/reading-header";
import { useActivePageStore } from "@/store/active-page";

// Init components
import { TranscriberInit } from "@/components/transcriber-init";
import { SpeechModelsInit } from "@/components/speech-models-init";
import { AppUpdater } from "@/components/app-updater";
import { SynchronizerInit } from "@/components/synchronizer-init";

export default function App() {
  const readMode = useActivePageStore((state) => state.readMode);
  const [sidebarOpen, setSidebarOpen] = useState(true);

  return (
    <SidebarProvider
      open={sidebarOpen && !readMode}
      onOpenChange={setSidebarOpen}
    >
      <AppSidebar />

      <SidebarInset>
        <header className="sticky top-0 z-30 flex h-16 shrink-0 items-center gap-2 border-b bg-background px-4">
          {readMode ? (
            <ReadingHeader />
          ) : (
            <>
              <SidebarTrigger className="-ml-1" />
              <Separator orientation="vertical" className="mr-2 h-4" />
              <AppHeaderBreadcrumb />
              <RightButtonsHeader />
            </>
          )}
        </header>

        <Content />

        <AppUpdater />
        <SynchronizerInit />
        <TranscriberInit />
        <SpeechModelsInit />
      </SidebarInset>
    </SidebarProvider>
  );
}
