//@ts-nocheck
import { useEffect, useRef, useState } from "react";
import CodeTool from "@editorjs/code";
import Delimiter from "@editorjs/delimiter";
import Ejs from "@editorjs/editorjs";
import Embed from "@editorjs/embed";
import Header from "@editorjs/header";
import InlineCode from "@editorjs/inline-code";
import LinkTool from "@editorjs/link";
import List from "@editorjs/list";
import Marker from "@editorjs/marker";
import ImageTool from "@editorjs/simple-image";
import Underline from "@editorjs/underline";
import DragDrop from "editorjs-drag-drop";
import Table from "@editorjs/table";

import "./style.css";

const createEditorJs = (holder: string | HTMLElement = "editorjs") => {
  const editor = new Ejs({
    holder,
    placeholder: "Write something...",
    autofocus: true,
    onReady: () => {
      new DragDrop(editor);
    },
    tools: {
      header: {
        // @ts-expect-error
        class: Header,
        inlineToolbar: true,
      },
      table: Table,
      delimiter: Delimiter,
      list: {
        // @ts-expect-error
        class: List,
        inlineToolbar: true,
      },
      embed: {
        // @ts-expect-error
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
  });

  return editor;
};

type EditorJSProps = {
  onReady?: () => void;
  data?: any;
};

// @ts-check
export function EditorJS(props: EditorJSProps) {
  const editorEl = useRef<HTMLDivElement>(null);
  const [editor, setEditor] = useState<Ejs | null>(null);
  const initialized = useRef(false);

  useEffect(() => {
    if (!editorEl.current || initialized.current) return;

    initialized.current = true;

    const ejs = createEditorJs(editorEl.current);
    ejs.isReady.then(() => {
      setEditor(ejs);
    });
  }, []);

  useEffect(() => {
    return;
  }, []);

  return <div ref={editorEl} />;
}
