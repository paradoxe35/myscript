import { createTreeTextWalker, surroundContentsTag } from "@/lib/dom";
import { useActivePageStore } from "@/store/active-page";
import { useTranscriberStore } from "@/store/transcriber";
import { useCallback, useEffect, useRef, useState } from "react";
import * as levenshtein from "damerau-levenshtein";

export function useContentReadMarker() {
  const transcriberStore = useTranscriberStore();
  const activePageStore = useActivePageStore();

  const lastMarkerPosition = useRef<number>(0);

  const containerRef = useRef<HTMLDivElement>(null);

  const onTranscribedText = useCallback((text: string) => {
    if (!containerRef.current) return;

    let position = 0;
    let chunks = text.trim().split(" ");

    const minSimilarity = 0.6;

    const treeWalker = createTreeTextWalker(containerRef.current);

    while (treeWalker.nextNode() && chunks.length > 0) {
      const node = treeWalker.currentNode;
      const nodeValueLength = node.nodeValue?.length || 0;

      let matchStartPosition = 0;
      let matchEndPosition = 0;

      if (lastMarkerPosition.current >= position + nodeValueLength) {
        position += nodeValueLength;
        continue;
      } else if (
        lastMarkerPosition.current < position + nodeValueLength &&
        lastMarkerPosition.current > position
      ) {
        matchStartPosition = lastMarkerPosition.current - position;
        matchEndPosition = matchStartPosition;
      }

      const nodeValue = node.nodeValue?.slice(matchStartPosition);
      const nodeValueChunks = nodeValue?.split(" ") || [];

      if (!nodeValueChunks.every((chunk) => /\w/gi.test(chunk))) {
        position += nodeValueLength;
        lastMarkerPosition.current = position;
        continue;
      }

      console.log("matchStartPosition: ", lastMarkerPosition.current);
      console.log("nodeValueChunks: ", nodeValueChunks, chunks);

      for (var i = 0; i < nodeValueChunks.length; i++) {
        const chunk = chunks.shift();
        const nodeValueChunk = nodeValueChunks[i];
        const lastChunk = nodeValueChunks.length - 1 === i;

        if (nodeValueChunk === undefined || chunk === undefined) {
          break;
        }

        const lev = levenshtein(chunk.trim(), nodeValueChunk.trim());

        if (lev.similarity >= minSimilarity) {
          matchEndPosition += nodeValueChunk.length + (lastChunk ? 0 : 1); // Add 1 to account for the space between chunks
        } else {
          break;
        }
      }

      if (matchStartPosition < matchEndPosition) {
        surroundContentsTag(node, matchStartPosition, matchEndPosition);
      }

      position += matchEndPosition;
      lastMarkerPosition.current = position;
    }
  }, []);

  useEffect(() => {
    return transcriberStore.onTranscribedText((text) => {
      onTranscribedText(text);
      console.log("Transcribed text:", text);
      // toast.success("Transcription saved successfully!");
    });
  }, [onTranscribedText]);

  return containerRef;
}
