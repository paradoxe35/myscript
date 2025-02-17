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
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import { createTreeTextWalker } from "@/lib/dom";
import { useDebouncedCallback } from "use-debounce";
import { useSyncRef } from "@/hooks/use-sync-ref";

const highlightClass = "bg-yellow-200";
const matchClass = "bg-slate-700/45";

type SearchMatch = {
  node: Node;
  start: number;
  end: number;
};

type Hook = ReturnType<typeof useFindEditMatcher>;

function useFindEditMatcher() {
  const [searchTerm, setSearchTerm] = useState("");
  const [replaceTerm, setReplaceTerm] = useState("");

  const [matches, setMatches] = useState<SearchMatch[]>([]);
  const [currentMatchIndex, setCurrentMatchIndex] = useState(-1);

  const highlightsRef = useRef<HTMLElement[]>([]);
  const containerRef = useRef<HTMLDivElement | null>(null);

  const handleSearch = useDebouncedCallback((searchTerm: string) => {
    if (!containerRef.current) {
      return;
    }

    cleanupHighlights();

    const matchesArray: SearchMatch[] = [];
    const walker = createTreeTextWalker(containerRef.current!);

    while (walker.nextNode()) {
      const node = walker.currentNode;

      if (!node.textContent) {
        continue;
      }

      const idx = node.textContent
        .toLowerCase()
        .indexOf(searchTerm.toLowerCase());

      if (idx > -1) {
        matchesArray.push({ node, start: idx, end: idx + searchTerm.length });
      }
    }

    const highlights: HTMLElement[] = [];

    matchesArray.forEach((match, idx) => {
      const highlight = document.createElement("span");
      highlight.classList.add(matchClass, `match-${idx}`);

      const range = document.createRange();
      range.setStart(match.node, match.start);
      range.setEnd(match.node, match.end);
      range.surroundContents(highlight);

      highlights.push(highlight);
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
  }, []);

  const onSearchTerm = setSearchTerm;
  const onReplaceTerm = setReplaceTerm;

  const handleNext = useCallback(() => {
    setCurrentMatchIndex((prev) => (prev + 1) % matches.length);
    scrollToMatch(currentMatchIndex + 1);
  }, []);

  const handlePrevious = useCallback(() => {
    setCurrentMatchIndex(
      (prev) => (prev - 1 + matches.length) % matches.length
    );
    scrollToMatch(currentMatchIndex - 1);
  }, [currentMatchIndex, matches]);

  const scrollToMatch = useCallback(
    (index: number) => {
      const c = containerRef.current;
      // Update style of previous match
      if (index > -1) {
        const prevMatch = c?.querySelector(`.match-${index - 1}`);
        prevMatch?.classList.remove(highlightClass);
        prevMatch?.classList.add(matchClass);
      }

      const match = c?.querySelector(`.match-${index}`);
      match?.classList.remove(matchClass);
      match?.classList.add(highlightClass);
    },
    [containerRef]
  );

  const handleReplace = useCallback(() => {
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
  }, [currentMatchIndex, matches, searchTerm, replaceTerm]);

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
      <div
        className={cn(
          "flex flex-col gap-2",
          "overflow-hidden rounded-md border p-3 shadow-lg transition-all"
        )}
      >
        <div className="flex gap-2 flex-col">
          <Input
            placeholder="Find"
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
              ? `${hook.currentMatchIndex + 1} of ${hook.matches.length}`
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
  const hook = useFindEditMatcher();
  const [toastId, setToastId] = useState<number | undefined>();

  const handleClose = useCallback((toastId: any) => {
    setToastId(undefined);
    toast.dismiss(toastId);
    hook.reset();
  }, []);

  useEffect(() => {
    let toastId: number | string | undefined;
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === "f" && !toastId) {
        e.preventDefault();
        setToastId(Math.random());
      } else if (e.key === "Escape") {
        handleClose(toastId);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [toastId]);

  useEffect(() => {
    if (toastId) {
      toast.custom(
        () => <ToastContent hook={hook} onClose={() => handleClose(toastId)} />,
        {
          id: toastId,
          duration: Infinity,
          onDismiss: () => setToastId(undefined),
          onAutoClose: () => setToastId(undefined),
        }
      );
    }
  }, [toastId, hook]);

  return hook.containerRef;
}
