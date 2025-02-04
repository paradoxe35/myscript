import { useActivePageStore } from "@/store/active-page";
import { useConfigStore } from "@/store/config";
import { useLocalPagesStore } from "@/store/local-pages";
import { useEffect } from "react";
import { EventsOn } from "~wails-runtime";

type AffectedTables = Record<string, string[]>;

export function SynchronizerInit() {
  const activePageStore = useActivePageStore();
  const localPagesStore = useLocalPagesStore();
  const configStore = useConfigStore();

  useEffect(() => {
    // Update UI on sync success
    return EventsOn(
      "on-sync-success",
      (affectedTables: AffectedTables | null) => {
        console.log("Synchronization succeeded:", affectedTables);

        // Refresh local pages
        localPagesStore.getPages();

        // Refresh active page blocks
        if (activePageStore.page?.__typename === "local_page") {
          activePageStore.fetchPageBlocks();
        }

        // Refresh config
        configStore.fetchConfig();
      }
    );
  }, []);

  useEffect(() => {
    return EventsOn("on-sync-failure", (error) => {
      console.error("Synchronization failed:", error);
    });
  }, []);

  return <></>;
}
