import { create } from "zustand";
import { useConfigStore } from "./config";
import { NotionPage, NotionSimplePage } from "@/types";
import { GetNotionPages } from "~wails/main/App";

type NotionPagesStore = {
  pages: Array<NotionPage>;
  getSimplifiedPages: () => Array<NotionSimplePage>;
  getPages: () => Promise<void>;
};

export const useNotionPagesStore = create<NotionPagesStore>((set, get) => ({
  pages: [],
  getPages: async () => {
    const pages = await GetNotionPages();
    set({ pages });
  },
  getSimplifiedPages: () => {
    return get().pages.map((page) => {
      return {
        id: page.id,
        title: page.properties.title.title.map((t) => t.plain_text).join(" "),
      };
    });
  },
}));

useConfigStore.subscribe(() => useNotionPagesStore.getState().getPages());
