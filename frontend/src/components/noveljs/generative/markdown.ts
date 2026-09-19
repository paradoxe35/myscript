import type { EditorInstance } from "novel";

/**
 * AI answers come back as Markdown. tiptap-markdown overrides insertContentAt
 * to re-parse its input inline, which flattens headings and lists, so parse to
 * HTML here and insert that through the untouched insertContent.
 */
function toHTML(editor: EditorInstance, markdown: string): string {
  return editor.storage.markdown?.parser?.parse(markdown) ?? markdown;
}

export function insertMarkdown(editor: EditorInstance, markdown: string) {
  editor.chain().focus().insertContent(toHTML(editor, markdown)).run();
}

export function insertMarkdownAt(
  editor: EditorInstance,
  at: number,
  markdown: string,
) {
  editor
    .chain()
    .focus()
    .setTextSelection(at)
    .insertContent(toHTML(editor, markdown))
    .run();
}

export function replaceWithMarkdown(
  editor: EditorInstance,
  range: { from: number; to: number },
  markdown: string,
) {
  editor
    .chain()
    .focus()
    .deleteRange(range)
    .insertContent(toHTML(editor, markdown))
    .run();
}
