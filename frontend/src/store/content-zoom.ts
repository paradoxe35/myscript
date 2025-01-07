import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

const ZOOM_ITERATIONS = [100, 200, 300] as const;

const DEFAULT_ZOOM = 100;

type ZoomStore = {
  zoom: (typeof ZOOM_ITERATIONS)[number];
  zoomIn: () => void;
  zoomOut: () => void;
  reset: () => void;
  canZoomIn: () => boolean;
  canZoomOut: () => boolean;
};

export const useContentZoomStore = create(
  persist<ZoomStore>(
    (set, get) => ({
      zoom: DEFAULT_ZOOM,

      reset: () => {
        set({ zoom: DEFAULT_ZOOM });
      },

      zoomIn: () => {
        const index = ZOOM_ITERATIONS.indexOf(get().zoom);
        if (index < ZOOM_ITERATIONS.length - 1) {
          set({ zoom: ZOOM_ITERATIONS[index + 1] });
        }
      },

      zoomOut: () => {
        const index = ZOOM_ITERATIONS.indexOf(get().zoom);
        if (index > 0) {
          set({ zoom: ZOOM_ITERATIONS[index - 1] });
        }
      },

      canZoomIn: () => {
        return ZOOM_ITERATIONS.indexOf(get().zoom) < ZOOM_ITERATIONS.length - 1;
      },

      canZoomOut: () => {
        return ZOOM_ITERATIONS.indexOf(get().zoom) > 0;
      },
    }),

    {
      name: "content-zoom",
      storage: createJSONStorage(() => localStorage),
    }
  )
);
