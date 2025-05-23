import { useEffect, useRef, forwardRef } from "react";
import CodeTool from "@editorjs/code";
import Delimiter from "@editorjs/delimiter";
import Ejs, { EditorConfig, API } from "@editorjs/editorjs";
// @ts-ignore
import Embed from "@editorjs/embed";
import Header from "@editorjs/header";
import InlineCode from "@editorjs/inline-code";
// @ts-ignore
import LinkTool from "@editorjs/link";
import List from "@editorjs/list";
// @ts-ignore
import Marker from "@editorjs/marker";
// @ts-ignore
import ImageTool from "@editorjs/simple-image";
import Underline from "@editorjs/underline";
// @ts-ignore
import DragDrop from "editorjs-drag-drop";
import Table from "@editorjs/table";

import "./style.css";
import { useSyncRef } from "@/hooks/use-sync-ref";
import { cn } from "@/lib/utils";

const createEditorJs = (
  holder: string | HTMLElement = "editorjs",
  configuration?: EditorConfig
) => {
  const editor = new Ejs({
    holder,
    placeholder: "Write something...",
    autofocus: true,
    onReady: () => {
      new DragDrop(editor);
    },
    tools: {
      header: {
        class: Header,
        config: {
          placeholder: "Enter a header",
          defaultLevel: 3,
        },
        inlineToolbar: true,
      },
      table: Table,
      delimiter: Delimiter,
      list: {
        class: List,
        inlineToolbar: true,
      },
      embed: {
        class: Embed,
        config: {
          services: {
            youtube: true,
            coub: true,
          },
        },
      },
      inlineCode: {
        class: InlineCode,
        shortcut: "CMD+SHIFT+M",
      },
      underline: Underline,
      code: CodeTool,
      image: ImageTool,
      linkTool: {
        class: LinkTool,
      },
      Marker: {
        class: Marker,
        shortcut: "CMD+SHIFT+M",
      },
    },

    ...(configuration as any),
  });

  return editor;
};

export type EditorJSProps = {
  onReady?: () => void;
  onChange?: (ejs: API, event: any) => void;
  defaultBlocks?: any[];
};

export { Ejs, type API };

export const EditorJS = forwardRef<Ejs, EditorJSProps>(
  (props: EditorJSProps, ref) => {
    const editorEl = useRef<HTMLDivElement>(null);
    const initialized = useRef(false);

    const $onChange = useSyncRef(props.onChange);

    useEffect(() => {
      if (!editorEl.current || initialized.current) return;

      initialized.current = true;

      const ejs = createEditorJs(editorEl.current, {
        onChange(api, event) {
          $onChange.current?.(api, event);
        },
      });

      ejs.isReady.then(() => {
        if (ref && typeof ref === "function") {
          ref(ejs);
        } else if (ref && typeof ref === "object") {
          ref.current = ejs;
        }

        if (ejs.render) {
          ejs.render({ blocks: props.defaultBlocks || [] });
        }
      });
    }, []);

    return (
      <div
        className={cn(
          "px-8 sm:px-12 max-w-[846px] w-full block mx-auto",
          "prose prose-lg dark:prose-invert "
        )}
        ref={editorEl}
      />
    );
  }
);

EditorJS.displayName = "EditorJS";
