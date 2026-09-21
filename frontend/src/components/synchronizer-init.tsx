import { useActivePageStore } from "@/store/active-page";
import { useConfigStore } from "@/store/config";
import {
  isGoogleAPIInvalidGrantError,
  useGoogleAuthTokenStore,
} from "@/store/google-auth-token";
import { useLocalPagesStore } from "@/store/local-pages";
import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { EventsOn, LogDebug } from "~wails-runtime";
import { StopSynchronizer } from "~wails/main/App";

type AffectedTables = Record<string, string[]>;

enum TABLES {
  PAGES = "pages",
  CONFIG = "configs",
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

          await StopSynchronizer().catch(console.error);

          googleAuthTokenStore.refreshToken().catch(() => {
            googleAuthTokenStore.deleteToken();
            toast.warning("Google auth token expired, please re-authorize");
            console.log("Failed to refresh Google auth token, deleting it");
          });
        }
      }

      console.log(
        "Is Invalid Grant Error:",
        isGoogleAPIInvalidGrantError(error)
      );
      console.error("Synchronization failed:", error);
    });
  }, []);

  useEffect(() => {
    return EventsOn(
      "on-sync-success",
      (affectedTables: AffectedTables | null) => {
        if (!affectedTables) {
          return;
        }

        if (TABLES.PAGES in affectedTables) {
          localPagesStore.getPages();
        }

        if (TABLES.CONFIG in affectedTables) {
          configStore.fetchConfig();
        }

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
