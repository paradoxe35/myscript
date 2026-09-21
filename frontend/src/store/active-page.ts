import { NotionSimplePage } from "@/types";
import { create } from "zustand";
import { repository } from "~wails/models";
import { persist, createJSONStorage } from "zustand/middleware";
import {
  GetCachedNotionPageBlocks,
  GetLocalPage,
  GetNotionPageBlocks,
} from "~wails/main/App";

type NotionActivePage = {
  __typename: "notion_page";
  page: NotionSimplePage;
  viewOnly: true;
  blocks?: any;
};

type LocalActivePage = {
  __typename: "local_page";
  page: repository.Page;
  viewOnly: false;
  blocks?: any;
};

type ActivePage = NotionActivePage | LocalActivePage;

type ActivePageStore = {
  page: ActivePage | null;
  readMode: boolean;
  version: number;
  getPageId(): number | string | undefined;
  setActivePage: (page: ActivePage) => void;
  unsetActivePage: () => void;
  fetchPageBlocks(): void;
  toggleReadMode: () => void;
  isReadMode(): boolean;
  setReadMode: (readMode: boolean) => void;
  canEdit(): boolean;
};

export const useActivePageStore = create(
  persist<ActivePageStore>(
    (set, get) => ({
      page: null,

      version: Date.now(),

      readMode: false,

      getPageId() {
        const activePage = get().page;

        return activePage?.__typename === "local_page"
          ? activePage.page.ID
          : activePage?.page.id;
      },

      isReadMode() {
        return get().readMode;
      },

      setReadMode(readMode: boolean) {
        set({ readMode });
      },

      setActivePage: (page: ActivePage) => {
        const currentPageId = get().getPageId();
        let readMode = get().readMode;
        const newPageId =
          page?.__typename === "local_page" ? page.page.ID : page?.page.id;

        let version = get().version;
        if (newPageId !== currentPageId) {
          version = Date.now();
          readMode = false;
        }

        set({ page, version, readMode });
      },

      canEdit() {
        const activePage = get().page;

        return (
          activePage?.__typename === "local_page" &&
          !activePage?.viewOnly &&
          !get().readMode
        );
      },

      toggleReadMode() {
        const readMode = get().readMode;
        set({ readMode: !readMode });

        return !readMode;
      },

      fetchPageBlocks() {
        const activePage = get().page;

        if (activePage?.__typename === "local_page") {
          GetLocalPage(activePage.page.ID).then((localPage) => {
            set({
              version: Date.now(),
              page: {
                ...activePage,
                page: localPage,
                blocks: localPage.blocks,
              },
            });
          });
        }

        if (activePage?.__typename === "notion_page") {
          const pageId = activePage.page.id;
          let fresh = false;

          const showBlocks = (blocks: unknown) => {
            if (get().getPageId() === pageId) {
              set({ version: Date.now(), page: { ...activePage, blocks } });
            }
          };

          // The cached copy renders at once; the fetch replaces it and wins
          // if it lands first.
          GetCachedNotionPageBlocks(pageId).then((blocks) => {
            if (blocks && !fresh) showBlocks(blocks);
          });

          GetNotionPageBlocks(pageId).then((blocks) => {
            fresh = true;
            showBlocks(blocks);
          });
        }
      },

      unsetActivePage() {
        set({ page: null });
      },
    }),

    {
      name: "active-page",
      storage: createJSONStorage(() => localStorage),
      // Only the open page survives a restart. Reading is a mode you enter
      // deliberately, so the app never comes back already reading.
      partialize: (state) =>
        ({ page: state.page }) as unknown as ActivePageStore,
    }
  )
);
