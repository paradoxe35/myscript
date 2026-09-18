import { useCallback, useEffect, useRef, useState } from "react";
import { EventsOn } from "~wails-runtime";
import { CancelAICompletion, StartAICompletion } from "~wails/main/App";
import { main } from "~wails/models";
import { type Option } from "./ai-selector-commands";

type CompletionEvent = {
  ID: string;
  Chunk: string;
  Error: string;
};

type Stream = { text: string; done: boolean; error?: string };

/**
 * Chunks can arrive before StartAICompletion has returned the stream's id, so
 * every stream is recorded by id and the view syncs from the active one.
 */
export function useAICompletion() {
  const [completion, setCompletion] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const streams = useRef(new Map<string, Stream>());
  const activeID = useRef<string | null>(null);

  const sync = useCallback((id: string) => {
    if (activeID.current !== id) return;

    const stream = streams.current.get(id);
    if (!stream) return;

    setCompletion(stream.text);
    setIsLoading(!stream.done);
    if (stream.error) setError(new Error(stream.error));
  }, []);

  const record = useCallback(
    (id: string, patch: Partial<Stream>) => {
      const stream = streams.current.get(id) ?? { text: "", done: false };
      streams.current.set(id, { ...stream, ...patch });
      sync(id);
    },
    [sync],
  );

  useEffect(() => {
    const clear = [
      EventsOn("on-ai-completion-chunk", (event: CompletionEvent) => {
        const stream = streams.current.get(event.ID);
        record(event.ID, { text: (stream?.text ?? "") + event.Chunk });
      }),
      EventsOn("on-ai-completion-done", (event: CompletionEvent) => {
        record(event.ID, { done: true });
      }),
      EventsOn("on-ai-completion-error", (event: CompletionEvent) => {
        record(event.ID, { done: true, error: event.Error });
      }),
    ];

    return () => clear.forEach((off) => off());
  }, [record]);

  const generate = useCallback(
    async (prompt: string, option: Option, command?: string) => {
      streams.current.clear();
      activeID.current = null;

      setCompletion("");
      setError(null);
      setIsLoading(true);

      try {
        const id = await StartAICompletion(
          main.AICompletionRequest.createFrom({
            Task: option,
            Text: prompt,
            Command: command ?? "",
          }),
        );

        activeID.current = id;
        if (!streams.current.has(id)) {
          streams.current.set(id, { text: "", done: false });
        }
        sync(id);
      } catch (err) {
        setIsLoading(false);
        setError(err instanceof Error ? err : new Error(String(err)));
      }
    },
    [sync],
  );

  const cancel = useCallback(() => {
    if (activeID.current) CancelAICompletion(activeID.current);
    setIsLoading(false);
  }, []);

  return { generate, cancel, completion, isLoading, error };
}
