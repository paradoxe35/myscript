import { buildWordIndex } from "./reading/word-index";

export const createTreeTextWalker = (element: Node) => {
  return document.createTreeWalker(element, NodeFilter.SHOW_TEXT, {
    acceptNode: (node) =>
      node.nodeValue?.length ? NodeFilter.FILTER_ACCEPT : NodeFilter.FILTER_SKIP,
  });
};

export const WORD_ATTRIBUTE = "data-word";

export type WrappedWords = { words: string[]; spans: HTMLElement[] };

// One pass: every word of the rendered script becomes a span carrying its index.
export function wrapWords(container: HTMLElement): WrappedWords {
  const walker = createTreeTextWalker(container);
  const nodes: Text[] = [];
  while (walker.nextNode()) nodes.push(walker.currentNode as Text);

  const index = buildWordIndex(nodes.map((node) => node.data));
  const spans: HTMLElement[] = [];
  let next = 0;

  nodes.forEach((node, segment) => {
    const first = next;
    while (next < index.length && index[next].segment === segment) next++;
    if (next === first) return;

    const fragment = document.createDocumentFragment();
    let cursor = 0;

    for (const word of index.slice(first, next)) {
      fragment.append(node.data.slice(cursor, word.start));

      const span = document.createElement("span");
      span.className = "reader-word";
      span.setAttribute(WORD_ATTRIBUTE, String(word.index));
      span.textContent = word.text;
      fragment.append(span);

      spans[word.index] = span;
      cursor = word.end;
    }

    fragment.append(node.data.slice(cursor));
    node.replaceWith(fragment);
  });

  return { words: index.map((word) => word.text), spans };
}

export function wordIndexOf(target: EventTarget | null): number | null {
  if (!(target instanceof Element)) return null;
  const span = target.closest(`[${WORD_ATTRIBUTE}]`);
  return span ? Number(span.getAttribute(WORD_ATTRIBUTE)) : null;
}

function scrollParent(element: HTMLElement): HTMLElement | null {
  for (let node = element.parentElement; node; node = node.parentElement) {
    const { overflowY } = getComputedStyle(node);
    if (/(auto|scroll|overlay)/.test(overflowY) && node.scrollHeight > node.clientHeight) {
      return node;
    }
  }
  return null;
}

export function scrollToEyeLine(element: HTMLElement) {
  const parent = scrollParent(element);
  const top = element.getBoundingClientRect().top;

  if (parent) {
    const offset = top - parent.getBoundingClientRect().top;
    parent.scrollTo({
      top: parent.scrollTop + offset - parent.clientHeight / 3,
      behavior: "smooth",
    });
  } else {
    window.scrollTo({
      top: window.scrollY + top - window.innerHeight / 3,
      behavior: "smooth",
    });
  }
}
