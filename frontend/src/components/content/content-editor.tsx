import { EditorInstance } from "novel";
import { useActivePageStore } from "@/store/active-page";
import { useLocalPagesStore } from "@/store/local-pages";
import NovelEditor from "../noveljs/advanced-editor";
import { useDebouncedCallback } from "use-debounce";
import hljs from "highlight.js";

export function ContentEditor() {
  const activePageStore = useActivePageStore();
  const savePageBlocks = useLocalPagesStore((state) => state.savePageBlocks);

  //Apply Codeblock Highlighting on the HTML from editor.getHTML()
  const highlightCodeblocks = (content: string) => {
    const fragment = document.createRange().createContextualFragment(content);
    fragment.querySelectorAll("pre code").forEach((el) => {
      // @ts-ignore
      // https://highlightjs.readthedocs.io/en/latest/api.html?highlight=highlightElement#highlightelement
      hljs.highlightElement(el);
    });

    return new XMLSerializer().serializeToString(fragment);
  };

  const debouncedUpdates = useDebouncedCallback(
    async (editor: EditorInstance) => {
      const activePage = activePageStore.page;
      if (activePage?.__typename !== "local_page" || activePage?.viewOnly)
        return;

      const json = editor.getJSON();
      const html = highlightCodeblocks(editor.getHTML());

      savePageBlocks(activePage.page, json, html).then((newPage) => {
        activePageStore.setActivePage({
          ...activePage,
          page: newPage,
          blocks: json as any,
        });
      });
    },
    500
  );

  return (
    <NovelEditor
      key={activePageStore.version}
      initialContent={activePageStore.page?.blocks}
      onUpdate={debouncedUpdates}
      className="w-full block mx-auto max-w-[846px]"
    />
  );
}
