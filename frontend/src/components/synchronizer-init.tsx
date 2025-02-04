import { useActivePageStore } from "@/store/active-page";
import { useConfigStore } from "@/store/config";
import {
  isGoogleAPIInvalidGrantError,
  useGoogleAuthTokenStore,
} from "@/store/google-auth-token";
import { useLocalPagesStore } from "@/store/local-pages";
import { useEffect, useRef } from "react";
import { EventsOn } from "~wails-runtime";
import { StopSynchronizer } from "~wails/main/App";

type AffectedTables = Record<string, string[]>;

enum TABLES {
  PAGES = "pages",
  CONFIG = "configs",
  CACHE = "caches",
}

export function SynchronizerInit() {
  const googleAuthTokenStore = useGoogleAuthTokenStore();

  const activePageStore = useActivePageStore();
  const localPagesStore = useLocalPagesStore();
  const configStore = useConfigStore();

  const syncFailures = useRef(0);

  useEffect(() => {
    return EventsOn("on-sync-failure", async (error) => {
      if (isGoogleAPIInvalidGrantError(error)) {
        syncFailures.current += 1;

        if (syncFailures.current === 4) {
          console.log("Too many sync attempts, refreshing Google auth token");

          // Stop the Scheduler
          await StopSynchronizer().catch(console.error);

          googleAuthTokenStore.refreshToken().catch(() => {
            console.log("Failed to refresh Google auth token, deleting it");
            googleAuthTokenStore.deleteToken();
          });
        }
      }

      console.error("Synchronization failed:", error);
    });
  }, []);

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

  return <></>;
}
