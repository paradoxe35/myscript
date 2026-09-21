import { create } from "zustand";
import { GetCache, SaveCache } from "~wails/main/App";

type ReadProgress = { word: number; total: number };

const RESUME_BY_DEFAULT = true;

type ContentReadState = {
  resume: boolean;
  position: number;
  total: number;

  setResume: (resume: boolean) => void;
  resetResume: () => void;
  setPosition: (position: number, total: number) => void;

  saveProgress: (pageId: string | number, progress: ReadProgress) => Promise<void>;
  loadProgress: (pageId: string | number) => Promise<ReadProgress>;
};

export const useContentReadStore = create<ContentReadState>((set) => ({
  resume: RESUME_BY_DEFAULT,
  position: 0,
  total: 0,

  setResume: (resume) => set({ resume }),

  resetResume: () => set({ resume: RESUME_BY_DEFAULT }),

  setPosition: (position, total) => set({ position, total }),

  saveProgress: async (pageId, progress) => {
    await SaveCache(`page-${pageId}-read-progress`, progress);
  },

  // Entries without a word index are stale and ignored.
  loadProgress: async (pageId) => {
    const cache = await GetCache(`page-${pageId}-read-progress`);
    const value = cache?.value;
    return typeof value?.word === "number"
      ? { word: value.word, total: value.total ?? 0 }
      : { word: 0, total: 0 };
  },
}));
