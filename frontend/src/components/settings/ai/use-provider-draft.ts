import { AIProvider, useAIProvidersStore } from "@/store/ai-providers";
import isEqual from "lodash/isEqual";
import { useCallback, useEffect, useMemo, useState } from "react";
import { main } from "~wails/models";

export type ProviderDraft = AIProvider & { apiKey: string };

type DraftState = {
  draft: ProviderDraft | null;
  saved: ProviderDraft | null;
};

function toDraft(provider: AIProvider, apiKey: string): ProviderDraft {
  return { ...main.AIProvider.createFrom(provider), apiKey };
}

/** Takes the freshly loaded values, unless an edit of that provider is in progress. */
export function reload(
  { draft, saved }: DraftState,
  next: ProviderDraft,
): DraftState {
  const editing = draft?.Name === next.Name && !isEqual(draft, saved);
  return { saved: next, draft: editing ? draft : next };
}

/** Edits to a provider, kept local until they are saved. */
export function useProviderDraft(provider: AIProvider | undefined) {
  const loadAPIKey = useAIProvidersStore((store) => store.apiKey);
  const saveProvider = useAIProvidersStore((store) => store.save);

  const [state, setState] = useState<DraftState>({ draft: null, saved: null });

  // The list behind the form is reloaded after a save or a sync. An edit in
  // progress on the same provider is kept; anything else is read afresh.
  useEffect(() => {
    if (!provider) {
      setState({ draft: null, saved: null });
      return;
    }

    let current = true;
    loadAPIKey(provider.Name).then((apiKey) => {
      if (!current) return;

      const next = toDraft(provider, apiKey);
      setState((current) => reload(current, next));
    });

    return () => {
      current = false;
    };
  }, [provider, loadAPIKey]);

  const update = useCallback((patch: Partial<ProviderDraft>) => {
    setState((current) =>
      current.draft
        ? { ...current, draft: { ...current.draft, ...patch } }
        : current,
    );
  }, []);

  const { draft, saved } = state;
  const dirty = useMemo(() => !isEqual(draft, saved), [draft, saved]);

  const save = useCallback(async () => {
    if (!draft) return;

    const { apiKey, ...provider } = draft;
    await saveProvider(main.AIProvider.createFrom(provider), apiKey);
    setState((current) => ({ ...current, saved: draft }));
  }, [draft, saveProvider]);

  return { draft, update, dirty, save };
}
