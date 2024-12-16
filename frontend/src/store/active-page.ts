import { NotionSimplePage } from "@/types";
import { create } from "zustand";
import { repository } from "~wails/models";
import { persist, createJSONStorage } from "zustand/middleware";
import { GetLocalPage, GetNotionPageBlocks } from "~wails/main/App";

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
  getPageId(): number | string | undefined;
  setActivePage: (page: ActivePage) => void;
  unsetActivePage: () => void;
  fetchPageBlocks(): void;
};

export const useActivePageStore = create(
  persist<ActivePageStore>(
    (set, get) => ({
      page: null,

      getPageId() {
        const activePage = get().page;

        return activePage?.__typename === "local_page"
          ? activePage.page.ID
          : activePage?.page.id;
      },

      setActivePage: (page: ActivePage) => {
        set({ page });
      },

      fetchPageBlocks() {
        const activePage = get().page;

        if (activePage?.__typename === "local_page") {
          GetLocalPage(activePage.page.ID).then((localPage) => {
            set({
              page: {
                ...activePage,
                blocks: localPage.blocks,
              },
            });
          });
        } else if (activePage?.__typename === "notion_page") {
          GetNotionPageBlocks(activePage.page.id).then((blocks) => {
            set({
              page: {
                ...activePage,
                blocks,
              },
            });
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
