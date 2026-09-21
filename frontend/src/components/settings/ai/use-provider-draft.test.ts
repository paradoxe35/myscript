import { describe, expect, it } from "vitest";
import { main } from "~wails/models";
import { ProviderDraft, reload } from "./use-provider-draft";

function draft(fields: Partial<ProviderDraft>): ProviderDraft {
  return {
    ...main.AIProvider.createFrom({ Name: "local", ...fields }),
    apiKey: "",
  };
}

describe("reload", () => {
  it("takes the loaded values when nothing is being edited", () => {
    const before = draft({ BaseURL: "http://a" });
    const next = draft({ BaseURL: "http://b" });

    expect(reload({ draft: before, saved: before }, next)).toEqual({
      draft: next,
      saved: next,
    });
  });

  it("keeps an edit in progress and only moves the saved baseline", () => {
    const saved = draft({ BaseURL: "http://a" });
    const editing = draft({ BaseURL: "http://a", Model: "typed" });
    const next = draft({ BaseURL: "http://b" });

    expect(reload({ draft: editing, saved }, next)).toEqual({
      draft: editing,
      saved: next,
    });
  });

  it("drops an edit of another provider", () => {
    const saved = draft({ Name: "other" });
    const editing = draft({ Name: "other", Model: "typed" });
    const next = draft({ Name: "local" });

    expect(reload({ draft: editing, saved }, next)).toEqual({
      draft: next,
      saved: next,
    });
  });

  it("starts from the loaded values when no draft exists yet", () => {
    const next = draft({});

    expect(reload({ draft: null, saved: null }, next)).toEqual({
      draft: next,
      saved: next,
    });
  });
});
