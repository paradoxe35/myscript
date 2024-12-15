import { EditorJS } from "@/components/editorjs";
import { cn } from "@/lib/utils";
import { useNotionStore } from "@/store/notion";
import { useEffect } from "react";

export function Content() {
  const notionStore = useNotionStore();

  useEffect(() => {
    notionStore.getPages();
  }, [notionStore.apiKey]);

  return (
    <div className="flex flex-1 flex-col gap-4 p-4">
      <div className="flex justify-center">
        <input
          type="text"
          placeholder="New Page"
          className={cn(
            "font-bold text-3xl bg-background text-foreground p-2 rounded-md",
            "max-w-[650px] mb-2 justify-self-center outline-none border-none w-full block"
          )}
        />
      </div>

      <EditorJS />
    </div>
  );
}
