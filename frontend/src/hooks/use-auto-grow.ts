import { RefObject, useLayoutEffect } from "react";

/**
 * Keeps a textarea as tall as its content.
 *
 * Pass everything that changes how the text wraps, including the zoom, which
 * changes the font size without changing any box. The observer watches the
 * parent, not the textarea, so its own resizes cannot retrigger it.
 */
export function useAutoGrow(
  ref: RefObject<HTMLTextAreaElement | null>,
  ...deps: unknown[]
) {
  useLayoutEffect(() => {
    const textarea = ref.current;
    if (!textarea) return;

    const fit = () => {
      textarea.style.height = "auto";
      textarea.style.height = `${textarea.scrollHeight}px`;
    };

    fit();

    const container = textarea.parentElement;
    if (!container) return;

    const observer = new ResizeObserver(fit);
    observer.observe(container);

    return () => observer.disconnect();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ref, ...deps]);
}
