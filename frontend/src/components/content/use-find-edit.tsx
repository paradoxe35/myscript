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

  const [currentMatchIndex, setCurrentMatchIndex] = useState(-1);

  const $currentMatchIndex = useSyncRef(currentMatchIndex);

  const [highlights, setHighlights] = useState<HTMLElement[]>([]);
  const $highlights = useSyncRef(highlights);

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

    const newHighlights: HTMLElement[] = [];

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

        newHighlights.push(highlight);
      }
    });

    setHighlights(newHighlights);
    setCurrentMatchIndex(matchesArray.length > 0 ? 0 : -1);

    if (matchesArray.length > 0) {
      scrollToMatch(0, 0);
    }
  }, 500);

  useEffect(() => {
    if (!searchTerm) {
      cleanupHighlights();
      setCurrentMatchIndex(-1);
      return;
    }

    handleSearch(searchTerm);

    return () => cleanupHighlights();
  }, [searchTerm, containerRef]);

  const cleanupHighlights = useCallback(() => {
    $highlights.current.forEach((highlight) => {
      const text = document.createTextNode(highlight.textContent || "");
      highlight.parentNode?.replaceChild(text, highlight);
    });
    setHighlights([]);
  }, [$highlights]);

  const onSearchTerm = setSearchTerm;
  const onReplaceTerm = setReplaceTerm;

  const handleNext = useCallback(() => {
    const currentMatchIndex = $currentMatchIndex.current;

    setCurrentMatchIndex((prev) => (prev + 1) % highlights.length);
    scrollToMatch(currentMatchIndex + 1, currentMatchIndex);
  }, [$currentMatchIndex, highlights]);

  const handlePrevious = useCallback(() => {
    const currentMatchIndex = $currentMatchIndex.current;

    setCurrentMatchIndex(
      (prev) => (prev - 1 + highlights.length) % highlights.length
    );
    scrollToMatch(currentMatchIndex - 1, currentMatchIndex);
  }, [$currentMatchIndex, highlights]);

  const scrollToMatch = useCallback(
    (index: number, currentMatchIndex: number) => {
      const c = containerRef.current;

      // Update style of previous match
      if (index !== currentMatchIndex) {
        const prevMatch = c?.querySelector(`.match-${currentMatchIndex}`);
        prevMatch?.classList.remove(...highlightClass);
        prevMatch?.classList.add(...matchClass);
      }

      const match = c?.querySelector(`.match-${index}`);
      match?.classList.remove(...matchClass);
      match?.classList.add(...highlightClass);
      match?.parentElement?.scrollIntoView({
        behavior: "smooth",
        block: "center",
      });
    },
    [containerRef]
  );

  const handleReplace = useCallback(() => {
    const currentMatchIndex = $currentMatchIndex.current;

    if (currentMatchIndex === -1) return;

    const highlight = highlights[currentMatchIndex];
    if (highlight) {
      const textNode = document.createTextNode(replaceTerm);
      highlight.parentNode?.replaceChild(textNode, highlight);
    }

    // Update matches after replacement
    const newHighlights = highlights.filter((_, i) => i !== currentMatchIndex);
    setHighlights(newHighlights);

    setCurrentMatchIndex(
      highlights.length > 0 ? currentMatchIndex % highlights.length : -1
    );
  }, [$currentMatchIndex, highlights, searchTerm, replaceTerm]);

  const handleReplaceAll = useCallback(() => {
    highlights.forEach((highlight) => {
      // @ts-ignore
      highlight.textContent = highlight.textContent?.replaceAll(
        searchTerm,
        replaceTerm
      );
    });
    setHighlights([]);
    setCurrentMatchIndex(-1);
  }, [highlights, searchTerm, replaceTerm]);

  const reset = useCallback(() => {
    cleanupHighlights();

    setSearchTerm("");
    setReplaceTerm("");
    setCurrentMatchIndex(-1);
  }, []);

  return {
    containerRef,

    replaceTerm,
    searchTerm,
    onReplaceTerm,
    onSearchTerm,

    matches: highlights,
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
        duration: Infinity,
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
