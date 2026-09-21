import { scrollToEyeLine, wordIndexOf, wrapWords } from "@/lib/dom";
import { ScriptReader } from "@/lib/reading/reader";
import { useActivePageStore } from "@/store/active-page";
import { useContentReadStore } from "@/store/content-read";
import { useTranscriberStore } from "@/store/transcriber";
import { useCallback, useEffect, useRef } from "react";
import { toast } from "sonner";
import { useDebouncedCallback } from "use-debounce";

const END_OF_PAGE_TOAST = "end-of-page";

// Every save is a SQLite write that also lands in the Drive change log, so
// matched phrases are batched and only the stop points write right away.
const SAVE_INTERVAL = 3000;

type PendingProgress = { pageId: string | number; word: number; total: number };

export function useContentReader(html: string) {
  const state = useTranscriberStore((store) => store.state);
  const stopRecording = useTranscriberStore((store) => store.stopRecording);
  const onTranscribedText = useTranscriberStore(
    (store) => store.onTranscribedText
  );

  const readMode = useActivePageStore((store) => store.readMode);
  const pageId = useActivePageStore((store) => store.getPageId());

  const resume = useContentReadStore((store) => store.resume);
  const setPosition = useContentReadStore((store) => store.setPosition);
  const saveProgress = useContentReadStore((store) => store.saveProgress);
  const loadProgress = useContentReadStore((store) => store.loadProgress);

  const containerRef = useRef<HTMLDivElement>(null);
  const readerRef = useRef<ScriptReader | null>(null);
  const spansRef = useRef<HTMLElement[]>([]);
  const paintedRef = useRef(0);
  const pendingRef = useRef<PendingProgress | null>(null);

  const listening = state === "listening";

  const save = useDebouncedCallback(
    () => {
      const pending = pendingRef.current;
      if (!pending) return;

      pendingRef.current = null;
      saveProgress(pending.pageId, { word: pending.word, total: pending.total });
    },
    SAVE_INTERVAL,
    { maxWait: SAVE_INTERVAL }
  );

  const paint = useCallback(
    (persist = true) => {
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

      setPosition(position, reader.total);
      if (persist && pageId) {
        pendingRef.current = { pageId, word: position, total: reader.total };
        save();
      }
    },
    [pageId, setPosition, save]
  );

  const moveTo = useCallback(
    (index: number) => {
      readerRef.current?.moveTo(index);
      paint();
    },
    [paint]
  );

  // The last word has been read: keep the final state and end the take. A stop
  // the user already asked for is left alone, so no second stop and no toast.
  const finish = useCallback(() => {
    if (useTranscriberStore.getState().stopping) return;

    save.flush();
    stopRecording();
    toast.info("Reached the end of the page", { id: END_OF_PAGE_TOAST });
  }, [save, stopRecording]);

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
      save.flush();
      readerRef.current = null;
      spansRef.current = [];
      setPosition(0, 0);
    };
  }, [readMode, html]);

  useEffect(() => {
    if (!listening || !readMode || !pageId) return;

    if (!resume) {
      moveTo(0);
      return;
    }

    loadProgress(pageId).then(({ word, total }) => {
      if (readerRef.current && total === readerRef.current.total) moveTo(word);
    });
  }, [listening, readMode]);

  useEffect(() => {
    if (!listening) return;

    const clear = onTranscribedText((text) => {
      const reader = readerRef.current;
      if (!reader || !reader.feed(text)) return;

      requestAnimationFrame(() => {
        paint();
        if (readerRef.current?.done) finish();
      });
    });

    // The take is over, whatever is still pending has to reach the disk.
    return () => {
      clear();
      save.flush();
    };
  }, [listening, paint, finish]);

  useEffect(() => () => save.flush(), []);

  const onClick = useCallback(
    (event: React.MouseEvent) => {
      const index = wordIndexOf(event.target);
      if (!readMode || index === null) return;

      moveTo(index);
      save.flush();
    },
    [readMode, moveTo, save]
  );

  return { containerRef, onClick };
}
