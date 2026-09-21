import { create } from "zustand";
import { GetPageReadProgress, SavePageReadProgress } from "~wails/main/App";
import { repository } from "~wails/models";

type ReadProgress = Pick<repository.ReadProgress, "word" | "total">;

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
    await SavePageReadProgress(String(pageId), progress.word, progress.total);
  },

  loadProgress: async (pageId) => {
    const { word, total } = await GetPageReadProgress(String(pageId));
    return { word, total };
  },
}));
