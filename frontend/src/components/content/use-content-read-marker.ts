import { createTreeTextWalker, surroundContentsTag } from "@/lib/dom";
import { useActivePageStore } from "@/store/active-page";
import { useTranscriberStore } from "@/store/transcriber";
import { useCallback, useEffect, useRef } from "react";
import * as levenshtein from "damerau-levenshtein";
import { Queue } from "@/lib/queue";

const minSimilarity = 0.8;
const fullStringMinSimilarity = 0.5;

const similar = (v: string, w: string, similarity: number = minSimilarity) => {
  return (
    levenshtein(v.trim().toLowerCase(), w.trim().toLowerCase()).similarity >=
    similarity
  );
};

const chunkSplitter = /[\s-']/;
const iChunkSplitter = /[^\s-']/g;

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

    let iteration = 0;
    let matchedString = "";
    let minChunksIteration = chunks.length - Math.floor(splitter / 2);
    let maxChunksIteration = chunks.length + Math.floor(splitter / 2);

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
      const nodeValueChunks = nodeValue?.split(chunkSplitter) || [];

      if (nodeValue?.trim() === "") {
        position += nodeValueLength;
        continue;
      }

      const arr = nodeValueChunks;
      for (let i = 0; i < arr.length; i++) {
        iteration++;
        const chunk = arr[i];

        if (!startMatched) {
          if (
            similar(chunks[0], chunk) ||
            (chunks[1] && similar(chunks[1], chunk))
          ) {
            startMatched = true;
          }
        }

        if (startMatched) {
          matchedString += chunk;
        }

        if (
          startMatched &&
          iteration > minChunksIteration &&
          iteration < maxChunksIteration
        ) {
          const isSimilar = similar(chunks[chunks.length - 1], chunk);

          if (isSimilar) {
            matchedString += chunk;

            const fullStringSimilarity = similar(
              matchedString,
              text,
              fullStringMinSimilarity
            );

            console.log(
              "Match and End: ",
              matchedString,
              text,

              "minChunksIteration: ",
              minChunksIteration,
              "maxChunksIteration: ",
              maxChunksIteration,
              "fullStringSimilarity: ",
              fullStringSimilarity
            );

            if (fullStringSimilarity) {
              const step = arr.reduce((acc, v, ii) => {
                if (ii < i) {
                  acc += v.length + 1;
                }
                return acc;
              }, 0);

              matchEndPosition = step + chunk.length;
              endMatched = true;
            } else {
              iteration = 0;
              matchedString = "";
              startMatched = false;
              endMatched = false;
            }
          }
        }

        if (startMatched && endMatched) {
          break;
        }
      }

      matchedNodes.push({
        node,
        start: matchStartPosition,
        end: matchEndPosition,
      });

      position += matchEndPosition;

      if (startMatched && endMatched) {
        break;
      }

      if (iteration > maxChunksIteration) {
        iteration = 0;
        startMatched = false;
        endMatched = false;
      }
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
