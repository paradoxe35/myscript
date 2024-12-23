import { create } from "zustand";

type ContentReaderStore = {};

export const useContentReaderStore = create<ContentReaderStore>(
  (set, get) => ({})
);
