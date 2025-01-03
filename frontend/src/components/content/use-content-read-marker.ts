import { createTreeTextWalker, surroundContentsTag } from "@/lib/dom";
import { useActivePageStore } from "@/store/active-page";
import { useTranscriberStore } from "@/store/transcriber";
import { useCallback, useEffect, useRef } from "react";
import * as levenshtein from "damerau-levenshtein";
import { Queue } from "@/lib/queue";
import { splitWithDelimiters } from "@/lib/utils";

const minSimilarity = 0.7;
const fullStringMinSimilarity = 0.5;

const similar = (v: string, w: string, similarity: number = minSimilarity) => {
  return (
    levenshtein(v.trim().toLowerCase(), w.trim().toLowerCase()).similarity >=
    similarity
  );
};

const chunkSplitter = /[^\p{L}\p{M}]+/u;
const iChunkSplitter = /[\p{L}\p{M}]+/gu;

function cleanUp(str: string) {
  return str.replace(/\[.*?\]/g, "");
}

export function strNormalize(str: string) {
  return str.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
}

const queue = new Queue(1);

export function useContentReadMarker() {
  const transcriberStore = useTranscriberStore();
  const activePageStore = useActivePageStore();

  const lastMarkerPosition = useRef<number>(0);

  const containerRef = useRef<HTMLDivElement>(null);

  const onTranscribedText2 = useCallback((text: string) => {
    if (!containerRef.current) return;

    let position = 0;

    let chunks = text.trim().split(chunkSplitter);
    let splitter = text.trim().replace(iChunkSplitter, "").length;

    let matchedString = "";
    let matchingTreeNodeIteration = 0;
    let ignoreIteration = 0;

    let iteration = 0;
    let minChunksIteration = chunks.length - Math.round(splitter / 2);
    let maxChunksIteration = chunks.length + Math.round(splitter / 2);

    console.log(
      "splitter",
      splitter,
      "min",
      minChunksIteration,
      "max",
      maxChunksIteration
    );

    let startMatched = false;
    let endMatched = false;
    let matchedNodes: { node: Node; start: number; end: number }[] = [];

    const treeWalker = createTreeTextWalker(containerRef.current);

    while (treeWalker.nextNode() && chunks.length > 0) {
      const node = treeWalker.currentNode;
      const nodeValueLength = node.nodeValue?.length || 0;

      let matchStartPosition = 0;
      let matchEndPosition = nodeValueLength;

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
      for (let i = 0; i < arr.length; i++) {
        iteration++;
        const chunk = arr[i];
        const chunkDelimiter = chunk.delimiter || "";

        if (!startMatched && iteration <= minChunksIteration) {
          if (
            similar(chunks[0], chunk.text) ||
            (chunks[1] && similar(chunks[1], chunk.text))
          ) {
            console.log("Start matched", chunk.text);
            startMatched = true;
          }
        }

        if (startMatched) {
          matchedString += chunk.text + chunkDelimiter;
        }

        if (
          startMatched &&
          iteration >= minChunksIteration &&
          iteration <= maxChunksIteration
        ) {
          const isSimilar = similar(chunks[chunks.length - 1], chunk.text);

          if (isSimilar) {
            const fullStringSimilarity = similar(
              matchedString,
              text,
              fullStringMinSimilarity
            );

            console.log(
              "Match End: ",
              matchedString,
              "lastChunk: ",
              chunk.text,

              "fullStringSimilarity: ",
              fullStringSimilarity
            );

            if (fullStringSimilarity) {
              const step = arr.reduce((acc, v, ii) => {
                if (ii < i) {
                  acc += v.text.length + (v.delimiter?.length || 0);
                }
                return acc;
              }, 0);

              endMatched = true;
              matchingTreeNodeIteration = 0;
              matchEndPosition = step + chunk.text.length;
            } else {
              iteration = 0;
              matchedString = "";
              startMatched = false;
              endMatched = false;

              // Go back to the previous node
              if (matchingTreeNodeIteration > 1) {
                for (let k = 0; k < matchingTreeNodeIteration - 1; k++) {
                  treeWalker.previousNode();
                  ignoreIteration += 1;
                }
              }
              matchingTreeNodeIteration = 0;
            }
          }
        }

        if (startMatched && endMatched) {
          break;
        }
      }

      if (startMatched) {
        matchingTreeNodeIteration += 1;
      }

      const existingNode = matchedNodes.find((n) => n.node === node);
      if (existingNode) {
        existingNode.start = matchStartPosition;
        existingNode.end = matchEndPosition;
      } else {
        matchedNodes.push({
          node,
          start: matchStartPosition,
          end: matchEndPosition,
        });
      }

      if (ignoreIteration > 0) {
        ignoreIteration -= 1;
      } else {
        position += matchEndPosition;
      }

      if (startMatched && endMatched) {
        break;
      }

      // This should come startMatched and endMatched condition (The one above)
      if (iteration > maxChunksIteration) {
        iteration = 0;
        startMatched = false;
        endMatched = false;

        // Go back to the previous node
        if (matchingTreeNodeIteration > 1) {
          for (let k = 0; k < matchingTreeNodeIteration - 1; k++) {
            treeWalker.previousNode();
            ignoreIteration += 1;
          }
        }
        matchingTreeNodeIteration = 0;
      }

      console.log("iteration: ", iteration);
    }

    console.log("Position: ", lastMarkerPosition.current, position);
    console.log("matchedNodes: ", matchedNodes);

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

      queue.task(onTranscribedText2, text);
      console.log("Transcribed text:", text);
    });
  }, [onTranscribedText2]);

  return containerRef;
}
