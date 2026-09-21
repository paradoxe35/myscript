import { beforeEach, describe, expect, it, vi } from "vitest";
import { repository } from "~wails/models";
import { useActivePageStore } from "./active-page";
import { useAIProvidersStore } from "./ai-providers";
import { useConfigStore } from "./config";
import { useLocalPagesStore } from "./local-pages";
import { refreshAfterSync } from "./sync-refresh";

const getPages = vi.fn();
const fetchPageBlocks = vi.fn();
const fetchConfig = vi.fn();
const fetchProviders = vi.fn();

function openLocalPage(ID: string) {
  useActivePageStore.setState({
    page: {
      __typename: "local_page",
      page: repository.Page.createFrom({ ID }),
      viewOnly: false,
    },
  });
}

describe("refreshAfterSync", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useLocalPagesStore.setState({ getPages });
    useActivePageStore.setState({ page: null, fetchPageBlocks });
    useConfigStore.setState({ fetchConfig });
    useAIProvidersStore.setState({ fetch: fetchProviders });
  });

  it("does nothing without affected tables", () => {
    refreshAfterSync(null);

    expect(getPages).not.toHaveBeenCalled();
    expect(fetchConfig).not.toHaveBeenCalled();
  });

  it("reloads the page list when pages changed", () => {
    refreshAfterSync({ pages: ["other"] });

    expect(getPages).toHaveBeenCalledOnce();
    expect(fetchPageBlocks).not.toHaveBeenCalled();
    expect(fetchConfig).not.toHaveBeenCalled();
  });

  it("reloads the open page only when it is among the changed ones", () => {
    openLocalPage("open");

    refreshAfterSync({ pages: ["other"] });
    expect(fetchPageBlocks).not.toHaveBeenCalled();

    refreshAfterSync({ pages: ["other", "open"] });
    expect(fetchPageBlocks).toHaveBeenCalledOnce();
  });

  it("reloads the config and the AI providers when configs changed", () => {
    refreshAfterSync({ configs: ["1"] });

    expect(fetchConfig).toHaveBeenCalledOnce();
    expect(fetchProviders).toHaveBeenCalledOnce();
    expect(getPages).not.toHaveBeenCalled();
  });
});
