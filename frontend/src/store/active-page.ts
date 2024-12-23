import { NotionSimplePage } from "@/types";
import { create } from "zustand";
import { repository } from "~wails/models";
import { persist, createJSONStorage } from "zustand/middleware";
import {
  GetCache,
  GetLocalPage,
  GetNotionPageBlocks,
  SaveCache,
} from "~wails/main/App";

type NotionActivePage = {
  __typename: "notion_page";
  page: NotionSimplePage;
  viewOnly: true;
  blocks?: Array<any>;
};

type LocalActivePage = {
  __typename: "local_page";
  page: repository.Page;
  viewOnly: false;
  blocks?: Array<any>;
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
                blocks: localPage.blocks,
              },
            });
          });
        } else if (activePage?.__typename === "notion_page") {
          // use Cache for notion page blocks
          const cacheKey = `${activePage?.__typename}:${activePage?.page.id}`;

          GetCache(cacheKey).then((cache) => {
            cache &&
              set({
                version: Date.now(),
                page: {
                  ...activePage,
                  blocks: cache.value,
                },
              });
          });

          GetNotionPageBlocks(activePage.page.id).then((blocks) => {
            if (get().getPageId() === activePage?.page.id) {
              set({
                version: Date.now(),
                page: {
                  ...activePage,
                  blocks,
                },
              });
            }

            SaveCache(cacheKey, blocks);
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
    }
  )
);
