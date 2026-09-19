import { create } from "zustand";

type EditorAIStore = {
  /** The prompt opened from the slash command, which has no selection to work on. */
  prompting: boolean;
  openPrompt: () => void;
  closePrompt: () => void;
};

export const useEditorAIStore = create<EditorAIStore>((set) => ({
  prompting: false,
  openPrompt: () => set({ prompting: true }),
  closePrompt: () => set({ prompting: false }),
}));
