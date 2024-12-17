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
  readMode: true;
  blocks?: Array<any>;
};

type LocalActivePage = {
  __typename: "local_page";
  page: repository.Page;
  readMode: boolean;
  blocks?: Array<any>;
};

type ActivePage = NotionActivePage | LocalActivePage;

type ActivePageStore = {
  page: ActivePage | null;
  version: number;
  getPageId(): number | string | undefined;
  setActivePage: (page: ActivePage) => void;
  unsetActivePage: () => void;
  fetchPageBlocks(): void;
};

export const useActivePageStore = create(
  persist<ActivePageStore>(
    (set, get) => ({
      page: null,

      version: Date.now(),

      getPageId() {
        const activePage = get().page;

        return activePage?.__typename === "local_page"
          ? activePage.page.ID
          : activePage?.page.id;
      },

      setActivePage: (page: ActivePage) => {
        const currentPageId = get().getPageId();
        const newPageId =
          page?.__typename === "local_page" ? page.page.ID : page?.page.id;

        let version = get().version;
        if (newPageId !== currentPageId) {
          version = Date.now();
        }

        set({ page, version });
      },

      fetchPageBlocks() {
        const activePage = get().page;

        if (activePage?.__typename === "local_page") {
          const cacheKey = `${activePage?.__typename}:${activePage?.page.ID}`;
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

          GetLocalPage(activePage.page.ID).then((localPage) => {
            set({
              version: Date.now(),
              page: {
                ...activePage,
                blocks: localPage.blocks,
              },
            });

            SaveCache(cacheKey, localPage.blocks);
          });
        } else if (activePage?.__typename === "notion_page") {
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
            set({
              version: Date.now(),
              page: {
                ...activePage,
                blocks,
              },
            });

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
