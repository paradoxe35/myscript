import { scrollToEyeLine, wordIndexOf, wrapWords } from "@/lib/dom";
import { ScriptReader } from "@/lib/reading/reader";
import { useActivePageStore } from "@/store/active-page";
import { useContentReadStore } from "@/store/content-read";
import { useTranscriberStore } from "@/store/transcriber";
import { useCallback, useEffect, useRef } from "react";
import { toast } from "sonner";

const END_OF_PAGE_TOAST = "end-of-page";
const AUTO_STOP_DELAY = 5000;

export function useContentReader(html: string) {
  const transcriberStore = useTranscriberStore();
  const activePageStore = useActivePageStore();
  const contentReadStore = useContentReadStore();

  const containerRef = useRef<HTMLDivElement>(null);
  const readerRef = useRef<ScriptReader | null>(null);
  const spansRef = useRef<HTMLElement[]>([]);
  const paintedRef = useRef(0);
  const autoStopRef = useRef<ReturnType<typeof setTimeout> | undefined>(
    undefined
  );

  const readMode = activePageStore.readMode;
  const listening = transcriberStore.state === "listening";
  const pageId = activePageStore.getPageId();

  const paint = useCallback(
    (save = true) => {
      const reader = readerRef.current;
      if (!reader) return;

      const spans = spansRef.current;
      const { position } = reader;
      const from = Math.min(paintedRef.current, position);
      const to = Math.min(
        Math.max(paintedRef.current, position) + 1,
        spans.length
      );

      for (let i = from; i < to; i++) {
        spans[i].classList.toggle("is-read", i < position);
        spans[i].classList.toggle("is-current", i === position);
      }
      paintedRef.current = position;

      const current = spans[position] ?? spans[position - 1];
      if (current) scrollToEyeLine(current);

      contentReadStore.setPosition(position, reader.total);
      if (save && pageId) {
        contentReadStore.saveProgress(pageId, {
          word: position,
          total: reader.total,
        });
      }
    },
    [pageId]
  );

  const cancelAutoStop = useCallback(() => {
    clearTimeout(autoStopRef.current);
    autoStopRef.current = undefined;
  }, []);

  const scheduleAutoStop = useCallback(() => {
    if (autoStopRef.current) return;

    autoStopRef.current = setTimeout(() => {
      autoStopRef.current = undefined;
      toast.dismiss(END_OF_PAGE_TOAST);
      transcriberStore.stopRecording();
    }, AUTO_STOP_DELAY);

    toast.info("Reached the end of the page", {
      id: END_OF_PAGE_TOAST,
      duration: AUTO_STOP_DELAY,
      action: { label: "Keep listening", onClick: cancelAutoStop },
    });
  }, [cancelAutoStop]);

  const moveTo = useCallback(
    (index: number) => {
      readerRef.current?.moveTo(index);
      cancelAutoStop();
      paint();
    },
    [paint, cancelAutoStop]
  );

  // Content can still arrive (Notion) while reading; keep the place when the words match.
  useEffect(() => {
    const container = containerRef.current;
    if (!readMode || !container) return;

    const previous = readerRef.current;
    const { words, spans } = wrapWords(container);
    const reader = new ScriptReader(words);
    if (previous?.total === reader.total) reader.moveTo(previous.position);

    readerRef.current = reader;
    spansRef.current = spans;
    paintedRef.current = 0;
    paint(false);

    return () => {
      readerRef.current = null;
      spansRef.current = [];
      cancelAutoStop();
      contentReadStore.setPosition(0, 0);
    };
  }, [readMode, html]);

  useEffect(() => {
    if (!listening || !readMode || !pageId) return;

    if (!contentReadStore.resume) {
      moveTo(0);
      return;
    }

    contentReadStore.loadProgress(pageId).then(({ word, total }) => {
      if (readerRef.current && total === readerRef.current.total) moveTo(word);
    });
  }, [listening, readMode]);

  useEffect(() => {
    if (!listening) return;

    return transcriberStore.onTranscribedText((text) => {
      const reader = readerRef.current;
      if (!reader || !reader.feed(text)) return;

      requestAnimationFrame(() => paint());
      if (reader.done) scheduleAutoStop();
    });
  }, [listening, paint, scheduleAutoStop]);

  const onClick = useCallback(
    (event: React.MouseEvent) => {
      const index = wordIndexOf(event.target);
      if (readMode && index !== null) moveTo(index);
    },
    [readMode, moveTo]
  );

  return { containerRef, onClick };
}
