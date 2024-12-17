import { useEffect, useRef, useState, forwardRef } from "react";
import CodeTool from "@editorjs/code";
import Delimiter from "@editorjs/delimiter";
import Ejs, { EditorConfig, API } from "@editorjs/editorjs";
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
        onChange: $onChange.current,
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
        className="prose max-w-[650px] dark:prose-invert w-full block mx-auto"
        ref={editorEl}
      />
    );
  }
);

EditorJS.displayName = "EditorJS";
