import { create } from "zustand";
import { NotionPage, NotionSimplePage } from "@/types";
import { useActivePageStore } from "./active-page";
import { GetNotionPages } from "~wails/main/App";
import { persist, createJSONStorage } from "zustand/middleware";

type NotionPagesStore = {
  pages: Array<NotionPage>;
  getSimplifiedPages(): Array<NotionSimplePage>;
  getPages: () => Promise<void>;
  resetPages: () => Promise<void>;
  /** After the token changes: the list and whatever page is open are both stale. */
  refresh: () => Promise<void>;
};

export const useNotionPagesStore = create(
  persist<NotionPagesStore>(
    (set, get) => ({
      pages: [],

      getPages: async () => {
        set({ pages: (await GetNotionPages()) || [] });
      },

      resetPages: async () => {
        set({ pages: [] });
      },

      refresh: async () => {
        await get().getPages();

        if (useActivePageStore.getState().page?.__typename === "notion_page") {
          useActivePageStore.getState().fetchPageBlocks();
        }
      },

      getSimplifiedPages() {
        return get().pages.map((page) => {
          return {
            id: page.id,
            title: page.properties.title.title
              .map((t) => t.plain_text)
              .join(" "),
          };
        });
      },
    }),

    {
      name: "notion-pages",
      storage: createJSONStorage(() => localStorage),
    },
  ),
);
