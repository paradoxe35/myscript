"use client";

import { useEffect, useState, useRef } from "react";
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

export function FindReplaceWidget() {
  const [searchTerm, setSearchTerm] = useState("");
  const [replaceTerm, setReplaceTerm] = useState("");
  const [matches, setMatches] = useState<{ node: Node; index: number }[]>([]);
  const [currentMatchIndex, setCurrentMatchIndex] = useState(-1);
  const [showReplace, setShowReplace] = useState(false);
  const highlightsRef = useRef<HTMLElement[]>([]);

  const cleanupHighlights = () => {
    highlightsRef.current.forEach((highlight) => {
      const text = document.createTextNode(highlight.textContent || "");
      highlight.parentNode?.replaceChild(text, highlight);
    });
    highlightsRef.current = [];
  };

  useEffect(() => {
    let toastId: number | string | undefined;
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === "f") {
        e.preventDefault();

        setShowReplace(false);

        toastId = toast.message("", {
          description: (
            <TooltipProvider>
              <div className="flex flex-col gap-2">
                <div className="flex gap-2">
                  <Input
                    placeholder="Find"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="h-8 w-48"
                  />
                  {showReplace && (
                    <Input
                      placeholder="Replace"
                      value={replaceTerm}
                      onChange={(e) => setReplaceTerm(e.target.value)}
                      className="h-8 w-48"
                    />
                  )}
                </div>
                <div className="flex items-center gap-2 text-sm">
                  <span className="text-muted-foreground text-xs">
                    {matches.length
                      ? `${currentMatchIndex + 1} of ${matches.length}`
                      : "No results"}
                  </span>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={handlePrevious}
                        disabled={!matches.length}
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
                        onClick={handleNext}
                        disabled={!matches.length}
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
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={handleReplace}
                        disabled={!matches.length}
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
                        onClick={handleReplaceAll}
                        disabled={!matches.length}
                      >
                        <ClipboardCopy className="h-4 w-4" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Replace all</TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          setSearchTerm("");
                          toast.dismiss(toastId);
                        }}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Close</TooltipContent>
                  </Tooltip>
                </div>
              </div>
            </TooltipProvider>
          ),
          duration: Infinity,
        });
      } else if (e.key === "Escape") {
        setSearchTerm("");
        toast.dismiss(toastId);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [searchTerm, matches, currentMatchIndex, showReplace]);

  useEffect(() => {
    cleanupHighlights();

    if (!searchTerm) {
      setMatches([]);
      setCurrentMatchIndex(-1);
      return;
    }

    const textNodes: Node[] = [];
    const walker = document.createTreeWalker(
      document.body,
      NodeFilter.SHOW_TEXT,
      null
    );

    while (walker.nextNode()) {
      const node = walker.currentNode;
      if (node.textContent?.includes(searchTerm)) {
        textNodes.push(node);
      }
    }

    const newHighlights: HTMLElement[] = [];
    const matchesArray = textNodes.flatMap((node) => {
      const indices = [];
      let idx = node.textContent!.indexOf(searchTerm);
      while (idx !== -1) {
        indices.push({ node, index: idx });
        idx = node.textContent!.indexOf(searchTerm, idx + searchTerm.length);
      }
      return indices;
    });

    matchesArray.forEach(({ node, index }) => {
      const text = node.textContent || "";
      const start = index;
      const end = index + searchTerm.length;
      const parent = node.parentNode;

      if (parent) {
        const before = text.slice(0, start);
        const after = text.slice(end);
        const highlight = document.createElement("span");
        highlight.className = "bg-yellow-200";
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

    highlightsRef.current = newHighlights;
    setMatches(matchesArray);
    setCurrentMatchIndex(matchesArray.length > 0 ? 0 : -1);

    return () => cleanupHighlights();
  }, [searchTerm]);

  const handleNext = () => {
    setCurrentMatchIndex((prev) => (prev + 1) % matches.length);
    scrollToMatch(currentMatchIndex + 1);
  };

  const handlePrevious = () => {
    setCurrentMatchIndex(
      (prev) => (prev - 1 + matches.length) % matches.length
    );
    scrollToMatch(currentMatchIndex - 1);
  };

  const scrollToMatch = (index: number) => {
    if (index >= 0 && index < matches.length) {
      const { node, index: textIndex } = matches[index];
      const range = document.createRange();
      range.setStart(node, textIndex);
      range.setEnd(node, textIndex + searchTerm.length);
      range.getBoundingClientRect();
      range.startContainer.parentElement?.scrollIntoView({ block: "center" });
    }
  };

  const handleReplace = () => {
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
  };

  const handleReplaceAll = () => {
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
  };

  return null;
}
