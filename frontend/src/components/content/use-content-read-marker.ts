import { createTreeTextWalker, surroundContentsTag } from "@/lib/dom";
import { useActivePageStore } from "@/store/active-page";
import { useTranscriberStore } from "@/store/transcriber";
import { useCallback, useEffect, useRef } from "react";
import * as levenshtein from "damerau-levenshtein";
import { Queue } from "@/lib/queue";
import { splitWithDelimiters } from "@/lib/utils";

const queue = new Queue(1);

const minSimilarity = 0.79;
const fullStringMinSimilarity = 0.6;

const similar = (v: string, w: string, similarity: number = minSimilarity) => {
  return (
    levenshtein(v.trim().toLowerCase(), w.trim().toLowerCase()).similarity >=
    similarity
  );
};

const chunkSplitter = /[^\p{L}\p{M}]+/gu;
const iChunkSplitter = /[\p{L}\p{M}]+/gu;

function cleanUp(str: string) {
  return str.replace(/\[.*?\]/g, "").trim();
}

function strNormalize(str: string) {
  return str.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
}

function scrollToLastMarker() {
  const element = Array.from(
    document.getElementsByClassName("screen-reader-marker")
  ).pop();

  if (element) {
    element.scrollIntoView({ behavior: "smooth", block: "center" });
  }
}

export function useContentReadMarker() {
  const transcriberStore = useTranscriberStore();
  const activePageStore = useActivePageStore();

  const lastMarkerPosition = useRef<number>(0);

  const containerRef = useRef<HTMLDivElement>(null);

  const onTranscribedText = useCallback((text: string) => {
    if (!containerRef.current) return;

    let position = 0;

    let chunks = text.trim().split(chunkSplitter).filter(Boolean);
    let splitter = text.trim().replace(iChunkSplitter, "").length;

    let matchedString = "";

    let matchingIteration = 0;
    let minChunksIteration = chunks.length - Math.round(splitter / 2);
    let maxChunksIteration = chunks.length + Math.round(splitter / 2);

    console.table({
      text,
      minChunksIteration,
      maxChunksIteration,
    });

    let lastMatchedNode: {
      node: Node;
      chunkIndex: number;
      position: number;
      matchedChunk: string;
    } | null = null;

    let jumpToChunkIndex: number | null = null;
    let ignoreCurrentNodeIteration: boolean = false;

    let startMatched = false;
    let endMatched = false;
    let matchedNodes: { node: Node; start: number; end: number }[] = [];

    const treeWalker = createTreeTextWalker(containerRef.current);

    /**
     * This function will move the tree walker back to the last matched node
     */
    const moveBackToLastMatchedNode = () => {
      if (!lastMatchedNode) {
        return false;
      }

      const node = lastMatchedNode.node;

      console.log("Jump back to node: ", lastMatchedNode);

      jumpToChunkIndex = lastMatchedNode.chunkIndex;
      position = lastMatchedNode.position;
      ignoreCurrentNodeIteration = true;

      let nNode: Node | null = treeWalker.currentNode;
      while (nNode && nNode !== node) {
        nNode = treeWalker.previousNode();
      }
      // We need to move back to the node
      // since the tree walker will move to the next node
      treeWalker.previousNode();

      lastMatchedNode = null;

      return true;
    };

    while (treeWalker.nextNode() && chunks.length > 0) {
      const node = treeWalker.currentNode;
      const nodeValueLength = node.nodeValue?.length || 0;

      let matchStartPosition = 0;
      let matchEndPosition = nodeValueLength;

      // If the was not ignored from the previous iteration,
      // we need to reset it
      ignoreCurrentNodeIteration = false;

      if (lastMarkerPosition.current >= position + nodeValueLength) {
        position += nodeValueLength;
        continue;
      } else if (
        lastMarkerPosition.current < position + nodeValueLength &&
        lastMarkerPosition.current > position
      ) {
        matchStartPosition = lastMarkerPosition.current - position;
      }

      const nodeValue = node.nodeValue?.slice(matchStartPosition);
      const nodeValueChunks =
        splitWithDelimiters(nodeValue || "", chunkSplitter) || [];

      if (nodeValue?.trim() === "") {
        position += nodeValueLength;
        continue;
      }

      const arr = nodeValueChunks;
      for (let i = jumpToChunkIndex || 0; i < arr.length; i++) {
        jumpToChunkIndex = null;

        const chunk = arr[i];
        const chunkDelimiter = chunk.delimiter || "";

        const isSimilar = () => {
          return (
            similar(chunks[0], chunk.text) ||
            (!!chunks[1] && similar(chunks[1], chunk.text))
          );
        };

        // This should be before the start matched check
        // if (startMatched && !lastMatchedNode) {
        //   if (isSimilar()) {
        //     lastMatchedNode = {
        //       node,
        //       position,
        //       chunkIndex: i,
        //       matchedChunk: chunk.text,
        //     };
        //   }
        // }

        if (!startMatched && matchingIteration <= minChunksIteration) {
          if (isSimilar()) {
            console.log("%cStart matched", "color: green");
            console.table({
              chunk: chunk.text,
              maxChunksIteration,
            });

            startMatched = true;
          }
        }

        if (startMatched) {
          matchingIteration++;
          matchedString += chunk.text + chunkDelimiter;
        }

        if (
          startMatched &&
          matchingIteration >= minChunksIteration &&
          matchingIteration <= maxChunksIteration
        ) {
          const isSimilar = similar(chunks[chunks.length - 1], chunk.text);

          if (isSimilar) {
            const fullStringSimilarity = similar(
              matchedString,
              text,
              fullStringMinSimilarity
            );

            console.log("%cEnd matched", "color: blue");
            console.table({
              matchedString,
              chunk: chunk.text,
              fullStringSimilarity,
            });

            if (fullStringSimilarity) {
              const step = arr.reduce((acc, v, ii) => {
                if (ii < i) {
                  acc += v.text.length + (v.delimiter?.length || 0);
                }
                return acc;
              }, 0);

              endMatched = true;
              matchEndPosition =
                step + chunk.text.length + (chunkDelimiter?.length || 0);
            } else {
              matchingIteration = 0;
              matchedString = "";
              startMatched = false;
              endMatched = false;
              // When there is no match, we need to move back to the last matched node
              // moveBackToLastMatchedNode();
            }
          }
        }

        if (startMatched && endMatched) {
          break;
        }
      }

      // if (ignoreCurrentNodeIteration) {
      //   ignoreCurrentNodeIteration = false;
      //   continue;
      // }

      const existsNode = matchedNodes.find((n) => n.node === node);

      if (existsNode) {
        existsNode.start = matchStartPosition;
        existsNode.end = matchEndPosition;
      } else {
        matchedNodes.push({
          node,
          start: matchStartPosition,
          end: matchEndPosition,
        });
      }

      position += matchEndPosition;

      if (startMatched && endMatched) {
        break;
      }

      // This should come startMatched and endMatched condition (The one above)
      if (matchingIteration > maxChunksIteration) {
        matchingIteration = 0;
        startMatched = false;
        endMatched = false;

        // When there is no match, we need to move back to the last matched node
        // moveBackToLastMatchedNode();
      }

      console.log("Iteration: ", matchingIteration);
    }

    console.log("Position: ", lastMarkerPosition.current, position);

    if (startMatched && endMatched) {
      lastMarkerPosition.current = position;

      matchedNodes.forEach((node) => {
        surroundContentsTag(node.node, node.start, node.end);
      });
    }
  }, []);

  useEffect(() => {
    return transcriberStore.onTranscribedText((text) => {
      text = cleanUp(text);

      queue.task(() => {
        requestAnimationFrame(() => {
          console.groupCollapsed("New transcription...");
          onTranscribedText(text);
          scrollToLastMarker();
          console.groupEnd();
        });
      });
    });
  }, [onTranscribedText]);

  return containerRef;
}
