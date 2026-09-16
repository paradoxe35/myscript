import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

export const ZOOM_STEPS = [100, 125, 150, 200] as const;

export type Zoom = (typeof ZOOM_STEPS)[number];

const DEFAULT_ZOOM: Zoom = 100;

type ZoomStore = {
  zoom: Zoom;
  zoomIn: () => void;
  zoomOut: () => void;
  reset: () => void;
  canZoomIn: () => boolean;
  canZoomOut: () => boolean;
};

const isZoom = (value: unknown): value is Zoom =>
  ZOOM_STEPS.includes(value as Zoom);

export const useContentZoomStore = create(
  persist<ZoomStore>(
    (set, get) => ({
      zoom: DEFAULT_ZOOM,

      reset: () => set({ zoom: DEFAULT_ZOOM }),

      zoomIn: () => {
        const index = ZOOM_STEPS.indexOf(get().zoom);
        if (index < ZOOM_STEPS.length - 1) set({ zoom: ZOOM_STEPS[index + 1] });
      },

      zoomOut: () => {
        const index = ZOOM_STEPS.indexOf(get().zoom);
        if (index > 0) set({ zoom: ZOOM_STEPS[index - 1] });
      },

      canZoomIn: () => ZOOM_STEPS.indexOf(get().zoom) < ZOOM_STEPS.length - 1,

      canZoomOut: () => ZOOM_STEPS.indexOf(get().zoom) > 0,
    }),

    {
      name: "content-zoom",
      version: 1,
      storage: createJSONStorage(() => localStorage),
      migrate: (state) => {
        const zoom = (state as Partial<ZoomStore>)?.zoom;
        return { zoom: isZoom(zoom) ? zoom : DEFAULT_ZOOM } as ZoomStore;
      },
    }
  )
);
