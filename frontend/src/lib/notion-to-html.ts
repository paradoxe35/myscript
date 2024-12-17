import { notion } from "~wails/models";

type NotionBlock = Omit<notion.NotionBlock, "Block"> & { Block: Block };

interface Block {
  object: string;
  id: string;
  type: string;
  [key: string]: any;
}

interface RichText {
  type: string;
  text: {
    content: string;
  };
  annotations: {
    bold: boolean;
    italic: boolean;
    strikethrough: boolean;
    underline: boolean;
    code: boolean;
    color: string;
  };
  plain_text: string;
}

export function convertNotionToHtml(blocks: NotionBlock[]): string {
  const html: string[] = [];

  for (const block of blocks) {
    html.push(convertBlockToHtml(block));
  }

  return html.join("\n");
}

export function wrapLists(html: string): string {
  // Wrap consecutive bulleted list items
  html = html.replace(/(<li>(?:(?!<\/li>).)*<\/li>)\s*(?=<li>)/g, "$1");

  // Wrap in ul tags
  html = html.replace(
    /(?:(?<=^)|(?<=<\/ul>)|(?<=<\/ol>)|(?<=<\/div>)|(?<=<\/p>))(\s*<li>(?:(?!<\/li>).)*<\/li>)+/g,
    "<ul>$1</ul>"
  );

  return html;
}

function convertBlockToHtml(notionBlock: NotionBlock): string {
  const blockType = notionBlock.Block.type;
  const content = notionBlock.Block[blockType];

  const rich_text = recursiveBlockToHtml(content.rich_text, notionBlock);

  switch (blockType) {
    case "heading_1":
      return `<h1>${rich_text}</h1>`;
    case "heading_2":
      return `<h2>${rich_text}</h2>`;
    case "heading_3":
      return `<h3>${rich_text}</h3>`;
    case "paragraph":
      return `<p>${rich_text}</p>`;
    case "bulleted_list_item":
      return `<li>${rich_text}</li>`;
    case "numbered_list_item":
      return `<li>${rich_text}</li>`;
    case "to_do":
      const checked = content.checked ? " checked" : "";
      return `<div class="todo-item">
          <input type="checkbox"${checked} disabled>
          <span>${rich_text}</span>
        </div>`;
    case "code":
      return `<pre><code class="language-${content.language}">${rich_text}</code></pre>`;
    case "quote":
      return `<blockquote>${rich_text}</blockquote>`;
    case "equation":
      return `<div class="equation">${content.expression || rich_text}</div>`;
    case "callout":
      return `<div class="callout p-4 bg-sidebar-accent rounded-md">${rich_text}</div>`;
    case "divider":
      return "<hr>";
    case "toggle":
      return `<details>
          <summary>${rich_text}</summary>
        </details>`;
    default:
      return `<!-- Unsupported block type: ${blockType} -->`;
  }
}

function recursiveBlockToHtml(
  richText: RichText[] | string | undefined,
  notionBlock: NotionBlock
): string {
  let text =
    typeof richText === "string"
      ? richText
      : convertRichTextToHtml(richText || []);

  if (notionBlock.Children.length > 0) {
    text += `<div class="child-blocks">`;
    text += convertNotionToHtml(notionBlock.Children);
    text += "</div>";
  }

  return text;
}

function convertRichTextToHtml(richText: RichText[]): string {
  if (!richText || richText.length === 0) return "";

  return richText
    .map((textObj) => {
      let text = escapeHtml(textObj.plain_text);

      // Apply text annotations
      if (textObj.annotations) {
        if (textObj.annotations.code) text = `<code>${text}</code>`;
        if (textObj.annotations.bold) text = `<strong>${text}</strong>`;
        if (textObj.annotations.italic) text = `<em>${text}</em>`;
        if (textObj.annotations.strikethrough) text = `<del>${text}</del>`;
        if (textObj.annotations.underline) text = `<u>${text}</u>`;
        if (textObj.annotations.color !== "default") {
          text = `<span class="color-${textObj.annotations.color}">${text}</span>`;
        }
      }

      return text;
    })
    .join("");
}

function escapeHtml(text: string): string {
  const htmlEntities: { [key: string]: string } = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;",
  };

  return text.replace(/[&<>"']/g, (char) => htmlEntities[char]);
}
