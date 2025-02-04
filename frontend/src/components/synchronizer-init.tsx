import { useActivePageStore } from "@/store/active-page";
import { useConfigStore } from "@/store/config";
import { useLocalPagesStore } from "@/store/local-pages";
import { useEffect } from "react";
import { EventsOn } from "~wails-runtime";

type AffectedTables = Record<string, string[]>;

enum TABLES {
  PAGES = "pages",
  CONFIG = "configs",
  CACHE = "caches",
}

export function SynchronizerInit() {
  const activePageStore = useActivePageStore();
  const localPagesStore = useLocalPagesStore();
  const configStore = useConfigStore();

  useEffect(() => {
    // Update UI on sync success
    return EventsOn(
      "on-sync-success",
      (affectedTables: AffectedTables | null) => {
        if (!affectedTables) {
          return;
        }

        // Refresh local pages
        if (TABLES.PAGES in affectedTables) {
          localPagesStore.getPages();
        }

        // Refresh config
        if (TABLES.CONFIG in affectedTables) {
          configStore.fetchConfig();
        }

        // Refresh active page blocks
        if (
          activePageStore.page?.__typename === "local_page" &&
          TABLES.PAGES in affectedTables
        ) {
          const activePage = activePageStore.page;
          const canRefreshBlocks = affectedTables[TABLES.PAGES].includes(
            activePage.page.ID
          );

          if (canRefreshBlocks) {
            activePageStore.fetchPageBlocks();
          }
        }
      }
    );
  }, [activePageStore.getPageId()]);

  useEffect(() => {
    return EventsOn("on-sync-failure", (error) => {
      console.error("Synchronization failed:", error);
    });
  }, []);

  return <></>;
}
