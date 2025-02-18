"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  ChevronLeft,
  ChevronRight,
  Replace,
  X,
  ArrowRightLeft,
  ClipboardCopy,
} from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { createTreeTextWalker } from "@/lib/dom";
import { useDebouncedCallback } from "use-debounce";
import { useToast } from "@/hooks/use-toast";
import { useSyncRef } from "@/hooks/use-sync-ref";

const highlightClass = ["bg-yellow-600/60"];
const matchClass = ["dark:bg-slate-400/60", "bg-slate-700/45"];

type SearchMatch = {
  node: Node;
  index: number;
};

type Hook = ReturnType<typeof useFindEditMatcher>;

function useFindEditMatcher() {
  const [searchTerm, setSearchTerm] = useState("");
  const [replaceTerm, setReplaceTerm] = useState("");

  const [matches, setMatches] = useState<SearchMatch[]>([]);
  const [currentMatchIndex, setCurrentMatchIndex] = useState(-1);

  const $currentMatchIndex = useSyncRef(currentMatchIndex);

  const highlightsRef = useRef<HTMLElement[]>([]);
  const containerRef = useRef<HTMLDivElement | null>(null);

  const handleSearch = useDebouncedCallback((searchTerm: string) => {
    if (!containerRef.current) {
      return;
    }

    cleanupHighlights();

    searchTerm = searchTerm.toLowerCase().trim();

    if (searchTerm.length < 2) {
      return;
    }

    const textNodes: Node[] = [];

    const walker = createTreeTextWalker(containerRef.current);

    while (walker.nextNode()) {
      const node = walker.currentNode;
      if (!node.textContent) {
        continue;
      }

      if (node.textContent?.includes(searchTerm)) {
        textNodes.push(node);
      }
    }

    const matchesArray = textNodes.flatMap((node) => {
      const indices = [];
      const textContent = node.textContent!.toLowerCase();

      let idx = textContent.indexOf(searchTerm);
      while (idx !== -1) {
        indices.push({ node, index: idx });
        idx = textContent.indexOf(searchTerm, idx + searchTerm.length);
      }
      return indices;
    });

    const highlights: HTMLElement[] = [];

    matchesArray.forEach(({ node, index }, idx) => {
      const text = node.textContent || "";
      const start = index;
      const end = index + searchTerm.length;
      const parent = node.parentNode;

      if (parent) {
        const before = text.slice(0, start);
        const after = text.slice(end);

        const highlight = document.createElement("span");
        highlight.classList.add(...matchClass, `match-${idx}`);

        highlight.textContent = text.slice(start, end);

        const beforeNode = document.createTextNode(before);
        const afterNode = document.createTextNode(after);

        parent.insertBefore(beforeNode, node);
        parent.insertBefore(highlight, node);
        parent.insertBefore(afterNode, node);
        parent.removeChild(node);

        highlights.push(highlight);
      }
    });

    highlightsRef.current = highlights;
    setMatches(matchesArray);
    setCurrentMatchIndex(matchesArray.length > 0 ? 0 : -1);

    if (matchesArray.length > 0) {
      scrollToMatch(0);
    }
  }, 500);

  useEffect(() => {
    if (!searchTerm) {
      cleanupHighlights();
      setMatches([]);
      setCurrentMatchIndex(-1);
      return;
    }

    handleSearch(searchTerm);

    return () => cleanupHighlights();
  }, [searchTerm, containerRef]);

  const cleanupHighlights = useCallback(() => {
    highlightsRef.current.forEach((highlight) => {
      const text = document.createTextNode(highlight.textContent || "");
      highlight.parentNode?.replaceChild(text, highlight);
    });
    highlightsRef.current = [];
  }, [highlightsRef]);

  const onSearchTerm = setSearchTerm;
  const onReplaceTerm = setReplaceTerm;

  const handleNext = useCallback(() => {
    setCurrentMatchIndex((prev) => (prev + 1) % matches.length);
    scrollToMatch($currentMatchIndex.current + 1);
  }, [$currentMatchIndex, matches]);

  const handlePrevious = useCallback(() => {
    setCurrentMatchIndex(
      (prev) => (prev - 1 + matches.length) % matches.length
    );
    scrollToMatch($currentMatchIndex.current - 1);
  }, [$currentMatchIndex, matches]);

  const scrollToMatch = useCallback(
    (index: number) => {
      const c = containerRef.current;
      // Update style of previous match
      if (index > -1) {
        const prevMatch = c?.querySelector(`.match-${index - 1}`);
        prevMatch?.classList.remove(...highlightClass);
        prevMatch?.classList.add(...matchClass);
      }

      const match = c?.querySelector(`.match-${index}`);
      match?.classList.remove(...matchClass);
      match?.classList.add(...highlightClass);
    },
    [containerRef]
  );

  const handleReplace = useCallback(() => {
    const currentMatchIndex = $currentMatchIndex.current;

    if (currentMatchIndex === -1) return;

    const match = matches[currentMatchIndex];
    const textNode = match.node;
    const newText = textNode.textContent!.replace(searchTerm, replaceTerm);
    textNode.textContent = newText;

    // Update matches after replacement
    const newMatches = matches.filter((_, i) => i !== currentMatchIndex);
    setMatches(newMatches);
    setCurrentMatchIndex(
      newMatches.length > 0 ? currentMatchIndex % newMatches.length : -1
    );
  }, [$currentMatchIndex, matches, searchTerm, replaceTerm]);

  const handleReplaceAll = useCallback(() => {
    matches.forEach((match) => {
      const textNode = match.node;
      // @ts-ignore
      textNode.textContent = textNode.textContent?.replaceAll(
        searchTerm,
        replaceTerm
      );
    });
    setMatches([]);
    setCurrentMatchIndex(-1);
  }, [matches, searchTerm, replaceTerm]);

  const reset = useCallback(() => {
    cleanupHighlights();

    setSearchTerm("");
    setReplaceTerm("");
    setMatches([]);
    setCurrentMatchIndex(-1);
  }, []);

  return {
    containerRef,

    replaceTerm,
    searchTerm,
    onReplaceTerm,
    onSearchTerm,

    matches,
    currentMatchIndex,

    reset,

    handleNext,
    handlePrevious,
    handleReplace,
    handleReplaceAll,
  };
}

type ToastContentProps = {
  onClose?: VoidFunction;
  hook: Hook;
};

function ToastContent(props: ToastContentProps) {
  const { hook } = props;

  const [showReplace, setShowReplace] = useState(false);

  const hasMatches = hook.matches.length > 0;

  const hasNext =
    hasMatches && hook.currentMatchIndex < hook.matches.length - 1;

  const hasPrev = hasMatches && hook.currentMatchIndex > 0;

  const canReplace =
    hook.replaceTerm.length > 0 && hasMatches && hook.currentMatchIndex >= 0;

  const canReplaceAll = hook.replaceTerm.length > 0 && hasMatches;

  return (
    <TooltipProvider>
      <div className={"flex flex-col gap-2"}>
        <div className="flex gap-2 flex-col">
          <Input
            placeholder="Find"
            autoFocus
            value={hook.searchTerm}
            onChange={(e) => hook.onSearchTerm(e.target.value)}
            className="h-8 w-48"
          />

          {showReplace && (
            <div className="flex items-center space-x-2">
              <Input
                placeholder="Replace"
                value={hook.replaceTerm}
                onChange={(e) => hook.onReplaceTerm(e.target.value)}
                className="h-8 w-48"
              />

              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={hook.handleReplace}
                    disabled={!canReplace}
                  >
                    <Replace className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Replace</TooltipContent>
              </Tooltip>

              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={hook.handleReplaceAll}
                    disabled={!canReplaceAll}
                  >
                    <ClipboardCopy className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Replace all</TooltipContent>
              </Tooltip>
            </div>
          )}
        </div>

        <div className="flex items-center gap-2 text-sm">
          <span className="text-muted-foreground text-xs">
            {hook.matches.length
              ? `${hook.currentMatchIndex + 1 || hook.matches.length} of ${
                  hook.matches.length
                }`
              : "No results"}
          </span>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                onClick={hook.handlePrevious}
                disabled={!hasPrev}
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Previous</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                onClick={hook.handleNext}
                disabled={!hasNext}
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Next</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowReplace(!showReplace)}
              >
                <ArrowRightLeft className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              {showReplace ? "Hide replace" : "Show replace"}
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="sm" onClick={props.onClose}>
                <X className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Close</TooltipContent>
          </Tooltip>
        </div>
      </div>
    </TooltipProvider>
  );
}

export function useFindEdit() {
  const { toast } = useToast();

  const hook = useFindEditMatcher();
  const [toaster, setToaster] = useState<
    ReturnType<typeof toast> | undefined
  >();

  const handleClose = useCallback((t: typeof toaster) => {
    t?.dismiss();

    requestAnimationFrame(() => {
      setToaster(undefined);
      hook.reset();
    });
  }, []);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === "f" && !toaster) {
        e.preventDefault();

        const selection = window.getSelection()?.toString();
        hook.onSearchTerm(selection || "");

        setToaster(toast({}));
      } else if (e.key === "Escape") {
        handleClose(toaster);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [toaster]);

  useEffect(() => {
    if (toaster) {
      toaster.update({
        id: toaster.id,
        duration: 9999999,
        description: (
          <ToastContent
            key={toaster.id}
            hook={hook}
            onClose={() => handleClose(toaster)}
          />
        ),
      });
    }
  }, [toaster, hook]);

  return hook.containerRef;
}
