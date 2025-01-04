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

// Function to wrap an element with a parent element
export function wrapElement(elementToWrap: Node, wrapperElement: Node) {
  // Get the parent of the element we want to wrap
  const parentElement = elementToWrap.parentNode;

  // Insert the wrapper before the element we want to wrap
  parentElement?.insertBefore(wrapperElement, elementToWrap);

  // Move the element into the wrapper
  wrapperElement.appendChild(elementToWrap);

  return wrapperElement;
}

export function surroundContentsTag(node: Node, start: number, end: number) {
  const range = document.createRange();

  const tag = document.createElement("span");

  const className = "bg-slide-primary bg-double animate-bg-slide";
  tag.classList.add(...className.split(" "));

  range.setStart(node, start);
  range.setEnd(node, end);
  range.surroundContents(tag);

  return range;
}
