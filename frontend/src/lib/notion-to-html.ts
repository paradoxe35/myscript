interface NotionBlock {
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

function convertBlockToHtml(block: NotionBlock): string {
  const blockType = block.type;
  const content = block[blockType];

  switch (blockType) {
    case "heading_1":
      return `<h1>${convertRichTextToHtml(content.rich_text)}</h1>`;
    case "heading_2":
      return `<h2>${convertRichTextToHtml(content.rich_text)}</h2>`;
    case "heading_3":
      return `<h3>${convertRichTextToHtml(content.rich_text)}</h3>`;
    case "paragraph":
      return `<p>${convertRichTextToHtml(content.rich_text)}</p>`;
    case "bulleted_list_item":
      return `<li>${convertRichTextToHtml(content.rich_text)}</li>`;
    case "numbered_list_item":
      return `<li>${convertRichTextToHtml(content.rich_text)}</li>`;
    case "to_do":
      const checked = content.checked ? " checked" : "";
      return `<div class="todo-item">
          <input type="checkbox"${checked} disabled>
          <span>${convertRichTextToHtml(content.rich_text)}</span>
        </div>`;
    case "code":
      return `<pre><code class="language-${
        content.language
      }">${convertRichTextToHtml(content.rich_text)}</code></pre>`;
    case "quote":
      return `<blockquote>${convertRichTextToHtml(
        content.rich_text
      )}</blockquote>`;
    case "equation":
      return `<div class="equation">${content.expression}</div>`;
    case "callout":
      return `<div class="callout">${convertRichTextToHtml(
        content.rich_text
      )}</div>`;
    case "divider":
      return "<hr>";
    case "toggle":
      return `<details>
          <summary>${convertRichTextToHtml(content.rich_text)}</summary>
          ${
            block.has_children
              ? "<!-- Child blocks should be processed here -->"
              : ""
          }
        </details>`;
    default:
      return `<!-- Unsupported block type: ${blockType} -->`;
  }
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
