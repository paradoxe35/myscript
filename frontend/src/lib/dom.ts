/**
 * @param {Node} element
 * @returns
 */
export const createTreeTextWalker = (element: Node) => {
  return document.createTreeWalker(element, NodeFilter.SHOW_TEXT, {
    acceptNode: function (node) {
      if (!node.nodeValue || !node.nodeValue.length) {
        return NodeFilter.FILTER_SKIP;
      }

      return NodeFilter.FILTER_ACCEPT;
    },
  });
};

export function surroundContentsTag(node: Node, start: number, end: number) {
  const range = document.createRange();
  const tag = document.createElement("mark");

  tag.style.backgroundColor = "#ff0000";
  tag.style.color = "#ffffff";

  range.setStart(node, start);
  range.setEnd(node, end);
  range.surroundContents(tag);

  return range;
}
