import { create } from "zustand";
import { GetCache, SaveCache } from "~wails/main/App";

type ContentReadState = {
  resume: boolean;

  setResume: (resume: boolean) => void;

  setContentReadProgress: (
    pageId: string | number,
    progress: number,
    total: number
  ) => Promise<void>;

  getContentReadProgress: (
    pageId: string | number
  ) => Promise<{ progress: number; total: number }>;
};

export const useContentReadStore = create<ContentReadState>((set) => ({
  resume: false,

  setResume: async (resume) => {
    set({ resume });
  },

  setContentReadProgress: async (pageId, progress, total) => {
    await SaveCache(`page-${pageId}-read-progress`, { progress, total });
  },

  getContentReadProgress: async (pageId) => {
    const cacheKey = `page-${pageId}-read-progress`;
    const cache = await GetCache(cacheKey);

    return cache?.value || { progress: 0, total: 0 };
  },
}));
