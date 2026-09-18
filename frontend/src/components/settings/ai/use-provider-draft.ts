import { AIProvider, useAIProvidersStore } from "@/store/ai-providers";
import isEqual from "lodash/isEqual";
import { useCallback, useEffect, useMemo, useState } from "react";
import { main } from "~wails/models";

export type ProviderDraft = AIProvider & { apiKey: string };

function toDraft(provider: AIProvider, apiKey: string): ProviderDraft {
  return { ...main.AIProvider.createFrom(provider), apiKey };
}

/** Edits to a provider, kept local until they are saved. */
export function useProviderDraft(provider: AIProvider | undefined) {
  const loadAPIKey = useAIProvidersStore((store) => store.apiKey);
  const saveProvider = useAIProvidersStore((store) => store.save);

  const [draft, setDraft] = useState<ProviderDraft | null>(null);
  const [saved, setSaved] = useState<ProviderDraft | null>(null);

  useEffect(() => {
    if (!provider) {
      setDraft(null);
      setSaved(null);
      return;
    }

    let current = true;
    loadAPIKey(provider.Name).then((apiKey) => {
      if (!current) return;
      setDraft(toDraft(provider, apiKey));
      setSaved(toDraft(provider, apiKey));
    });

    return () => {
      current = false;
    };
  }, [provider?.Name, provider?.Model, provider?.BaseURL, loadAPIKey]);

  const update = useCallback((patch: Partial<ProviderDraft>) => {
    setDraft((current) => (current ? { ...current, ...patch } : current));
  }, []);

  const dirty = useMemo(() => !isEqual(draft, saved), [draft, saved]);

  const save = useCallback(async () => {
    if (!draft) return;

    const { apiKey, ...provider } = draft;
    await saveProvider(main.AIProvider.createFrom(provider), apiKey);
    setSaved(draft);
  }, [draft, saveProvider]);

  return { draft, update, dirty, save };
}
