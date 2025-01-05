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
  const lev = levenshtein(v.trim().toLowerCase(), w.trim().toLowerCase());

  return {
    isSimilar: lev.similarity >= similarity,
    levenshtein: lev,
  };
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

    let startMatched = false;
    let endMatched = false;

    type MatchedNode = { node: Node; start: number; end: number };
    let matchedNodes: MatchedNode[] = [];

    type MatchedChunk = {
      node: Node;
      similarity: number;
      matchedString: string;
      matchStartPosition: number;
      matchEndPosition: number;
      position: number;
    };
    let matchedChunks: MatchedChunk[] = [];

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
        const chunk = arr[i];
        const chunkDelimiter = chunk.delimiter || "";

        const isSimilar = () => {
          const i1 = similar(chunks[0], chunk.text).isSimilar;

          return (
            i1 || (!!chunks[1] && similar(chunks[1], chunk.text).isSimilar)
          );
        };

        if (!startMatched && matchingIteration <= minChunksIteration) {
          if (isSimilar()) {
            // Reset matched chunks on start matched
            matchedChunks = [];

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
          const { isSimilar, levenshtein } = similar(
            matchedString,
            text,
            fullStringMinSimilarity
          );

          if (isSimilar) {
            const $matchedEndPosition = arr.reduce((acc, v, ii) => {
              if (ii <= i) {
                acc += v.text.length + (v.delimiter?.length || 0);
              }
              return acc;
            }, 0);

            matchedChunks.push({
              node,
              similarity: levenshtein.similarity,
              matchedString,
              matchStartPosition,
              position: position + $matchedEndPosition,
              matchEndPosition: $matchedEndPosition,
            });

            if (matchingIteration === maxChunksIteration) {
              matchEndPosition = $matchedEndPosition;
            }
          }

          if (matchingIteration === maxChunksIteration) {
            if (matchedChunks.length > 0) {
              endMatched = true;

              console.log("%cEnd matched", "color: blue");
              console.table({
                matchedString,
                chunk: chunk.text,
              });
            } else {
              matchingIteration = 0;
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

      // This should come startMatched and endMatched condition (The one above)
      if (matchingIteration > maxChunksIteration) {
        matchingIteration = 0;
        startMatched = false;
        endMatched = false;
      }

      console.log("Iteration: ", matchingIteration);
    }

    if (startMatched && endMatched) {
      const bestChunk = matchedChunks.reduce((acc, v) => {
        if (acc.similarity < v.similarity) {
          return v;
        }
        return acc;
      });

      if (!bestChunk) return;

      const nodes: MatchedNode[] = [];

      for (let match of matchedNodes) {
        if (match.node === bestChunk.node) {
          match.start = bestChunk.matchStartPosition;
          match.end = bestChunk.matchEndPosition;
          nodes.push(match);
          break;
        } else {
          nodes.push(match);
        }
      }

      lastMarkerPosition.current = bestChunk.position;

      console.table({
        bestChunk,
        lastMarkerPosition: lastMarkerPosition.current,
      });

      nodes.forEach((node) => {
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
