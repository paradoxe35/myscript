import { create } from "zustand";
import { GetCache, SaveCache } from "~wails/main/App";

type ReadProgress = { word: number; total: number };

type ContentReadState = {
  resume: boolean;
  position: number;
  total: number;

  setResume: (resume: boolean) => void;
  setPosition: (position: number, total: number) => void;

  saveProgress: (pageId: string | number, progress: ReadProgress) => Promise<void>;
  loadProgress: (pageId: string | number) => Promise<ReadProgress>;
};

export const useContentReadStore = create<ContentReadState>((set) => ({
  resume: false,
  position: 0,
  total: 0,

  setResume: (resume) => set({ resume }),

  setPosition: (position, total) => set({ position, total }),

  saveProgress: async (pageId, progress) => {
    await SaveCache(`page-${pageId}-read-progress`, progress);
  },

  // Entries written before word indexing hold character offsets and are ignored.
  loadProgress: async (pageId) => {
    const cache = await GetCache(`page-${pageId}-read-progress`);
    const value = cache?.value;
    return typeof value?.word === "number"
      ? { word: value.word, total: value.total ?? 0 }
      : { word: 0, total: 0 };
  },
}));
