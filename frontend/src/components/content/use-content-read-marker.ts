import * as levenshtein from "damerau-levenshtein";
import { createTreeTextWalker, surroundContentsTag } from "@/lib/dom";
import { useActivePageStore } from "@/store/active-page";
import { useTranscriberStore } from "@/store/transcriber";
import { useCallback, useEffect, useRef } from "react";
import { Queue } from "@/lib/queue";
import { splitWithDelimiters } from "@/lib/utils";
import { useContentReadStore } from "@/store/content-read";
import { toast } from "sonner";

const queue = new Queue(1);

const minSimilarity = 0.79;
const fullStringMinSimilarity = 0.6;

const chunkSplitter = /[^\p{L}\p{M}]+/gu;
const iChunkSplitter = /[\p{L}\p{M}]+/gu;

const similar = (v: string, w: string, similarity: number = minSimilarity) => {
  const lev = levenshtein(v.trim().toLowerCase(), w.trim().toLowerCase());

  return {
    isSimilar: lev.similarity >= similarity,
    levenshtein: lev,
  };
};

const fullStringSimilar = (v: string, w: string) => {
  v = v.replace(chunkSplitter, "").trim().toLowerCase();
  w = w.replace(chunkSplitter, "").trim().toLowerCase();

  const lev = levenshtein(v, w);

  return {
    isSimilar: lev.similarity >= fullStringMinSimilarity,
    levenshtein: lev,
  };
};

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

type MatchedNode = { node: Node; start: number; end: number };

type MatchedChunk = {
  node: Node;
  similarity: number;
  matchedString: string;
  matchStartPosition: number;
  matchEndPosition: number;
  position: number;
};

export function useContentReadMarker() {
  const transcriberStore = useTranscriberStore();
  const activePageStore = useActivePageStore();
  const contentReadStore = useContentReadStore();

  const totalContentLength = useRef<number>(0);
  const lastMarkerPosition = useRef<number>(0);

  const containerRef = useRef<HTMLDivElement>(null);

  /**
   * This function will try to match the text with the current page content using levenshtein distance
   * If a match is found, it will surround the matched content with a span element (a marker)
   */
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

    let matchedNodes: MatchedNode[] = [];
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

        // Try to find a start match so we can start similarity calculation
        if (!startMatched && matchingIteration <= minChunksIteration) {
          const isSimilar = () => {
            const i1 = similar(chunks[0], chunk.text).isSimilar;

            return (
              i1 || (!!chunks[1] && similar(chunks[1], chunk.text).isSimilar)
            );
          };

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
          const { isSimilar, levenshtein } = fullStringSimilar(
            matchedString,
            text
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
              matchedChunks = [];
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
        matchedChunks = [];
        matchingIteration = 0;
        startMatched = false;
        endMatched = false;
      }
    }

    if (matchedChunks.length > 0) {
      const bestChunk = matchedChunks.reduce((acc, v) => {
        if (acc.similarity < v.similarity) {
          return v;
        }
        return acc;
      });

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

  const setTotalContentLength = useCallback(() => {
    if (!containerRef.current) return;

    const treeWalker = createTreeTextWalker(containerRef.current);

    let length = 0;
    let node: Node | null = null;
    let ignoredLines = 0;

    while (treeWalker.nextNode()) {
      node = treeWalker.currentNode;
      length += node.nodeValue?.length || 0;

      if (node.nodeValue === "\n") {
        ignoredLines += 1;
      } else {
        ignoredLines = 0;
      }
    }

    totalContentLength.current = length - ignoredLines;
  }, [containerRef]);

  const moveMarkerToLastPosition = useCallback(() => {
    if (!containerRef.current || lastMarkerPosition.current === 0) return;

    const treeWalker = createTreeTextWalker(containerRef.current);
    const nodes: MatchedNode[] = [];

    let position = 0;
    while (treeWalker.nextNode()) {
      const node = treeWalker.currentNode;
      const nodeValueLength = node.nodeValue?.length || 0;

      let matchStartPosition = 0;
      let matchEndPosition = nodeValueLength;

      if (position >= lastMarkerPosition.current) {
        break;
      } else if (
        lastMarkerPosition.current < position + nodeValueLength &&
        lastMarkerPosition.current > position
      ) {
        matchEndPosition = lastMarkerPosition.current - position;
      } else if (node.nodeValue?.trim() === "") {
        position += nodeValueLength;
        continue;
      }

      nodes.push({
        node,
        start: matchStartPosition,
        end: matchEndPosition,
      });

      position += nodeValueLength;
    }

    nodes.forEach((node) => {
      surroundContentsTag(node.node, node.start, node.end);
    });
  }, []);

  const onTranscriptionProgress = useCallback(() => {
    const pageId = activePageStore.getPageId();

    // Set content read progress
    if (pageId) {
      const progress = {
        progress: lastMarkerPosition.current,
        total: totalContentLength.current,
      };

      console.log("Progress: %d%", (progress.progress / progress.total) * 100);

      if (progress.progress >= progress.total && progress.total > 0) {
        progress.total = 0;
        progress.progress = 0;

        // Stop recording
        toast.info("Reached end of page");
        transcriberStore.stopRecording();
      }

      contentReadStore.setContentReadProgress(
        pageId,
        progress.progress,
        progress.total
      );
    }
  }, []);

  // Move marker to last position
  useEffect(() => {
    const pageId = activePageStore.getPageId();

    if (!contentReadStore.resume) {
      lastMarkerPosition.current = 0;
      return;
    }

    if (activePageStore.readMode && pageId) {
      contentReadStore.getContentReadProgress(pageId).then((progress) => {
        if (!progress || progress.progress === 0) {
          return;
        }
        lastMarkerPosition.current = progress.progress;
        moveMarkerToLastPosition();
      });
    }
  }, [activePageStore.readMode]);

  useEffect(() => {
    if (activePageStore.readMode) {
      setTotalContentLength();
    }
  }, [activePageStore.readMode]);

  useEffect(() => {
    return transcriberStore.onTranscribedText((text) => {
      text = cleanUp(text);

      queue.task(() => {
        requestAnimationFrame(() => {
          console.groupCollapsed("New transcription...");

          onTranscribedText(text);
          scrollToLastMarker();
          onTranscriptionProgress();

          console.groupEnd();
        });
      });
    });
  }, [onTranscribedText]);

  return containerRef;
}
