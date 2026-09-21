import { useActivePageStore } from "@/store/active-page";
import { useAIProvidersStore } from "@/store/ai-providers";
import { useConfigStore } from "@/store/config";
import { useLocalPagesStore } from "@/store/local-pages";

export type AffectedTables = Record<string, string[]>;

const PAGES = "pages";
const CONFIGS = "configs";

// Reloads what the window derives from the synced tables. The page language
// in `caches` is read when the reading modal opens, so it needs nothing here.
export function refreshAfterSync(affectedTables: AffectedTables | null) {
  if (!affectedTables) return;

  if (PAGES in affectedTables) {
    useLocalPagesStore.getState().getPages();

    const activePage = useActivePageStore.getState();
    const page = activePage.page;
    if (
      page?.__typename === "local_page" &&
      affectedTables[PAGES].includes(page.page.ID)
    ) {
      activePage.fetchPageBlocks();
    }
  }

  if (CONFIGS in affectedTables) {
    useConfigStore.getState().fetchConfig();
    useAIProvidersStore.getState().fetch();
  }
}
