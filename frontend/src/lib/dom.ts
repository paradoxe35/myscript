/**
 * @param {Node} container
 * @returns
 */
export const createTreeTextWalker = (container: Node) => {
  return document.createTreeWalker(container, NodeFilter.SHOW_TEXT, {
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
  const parentElement = elementToWrap.parentNode;

  parentElement?.insertBefore(wrapperElement, elementToWrap);

  wrapperElement.appendChild(elementToWrap);

  return wrapperElement;
}

export function surroundContentsTag(node: Node, start: number, end: number) {
  const range = document.createRange();

  const tag = document.createElement("span");
  tag.setAttribute("data-mark-id", "true");

  const className =
    "bg-slide-primary bg-double animate-bg-slide screen-reader-marker";
  tag.classList.add(...className.split(" "));

  range.setStart(node, start);
  range.setEnd(node, end);
  range.surroundContents(tag);

  return range;
}

export function cleanMarkTags(container = document.body) {
  Array.from(container.querySelectorAll("[data-mark-id]")).forEach((el) => {
    let textContent = el.textContent;

    let sibling = null;
    // always merge next sibling element with the current node
    while (el.nextSibling && el.nextSibling.nodeType === Node.TEXT_NODE) {
      sibling = el.nextSibling;
      textContent += sibling.textContent || "";
      sibling.parentNode?.removeChild(sibling);
    }

    // always merge previous sibling element with the current node
    while (
      el.previousSibling &&
      el.previousSibling.nodeType === Node.TEXT_NODE
    ) {
      sibling = el.previousSibling;
      textContent = (sibling.textContent || "") + textContent;
      sibling.parentNode?.removeChild(sibling);
    }

    const fragment = document
      .createRange()
      .createContextualFragment(textContent || "");

    el.parentNode?.replaceChild(fragment, el);
  });
}
