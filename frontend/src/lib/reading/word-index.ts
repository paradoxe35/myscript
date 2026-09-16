import { WORD_PATTERN } from "./similarity";

export type ScriptWord = {
  index: number;
  text: string;
  segment: number;
  start: number;
  end: number;
};

export function buildWordIndex(segments: string[]): ScriptWord[] {
  const words: ScriptWord[] = [];

  segments.forEach((segment, segmentIndex) => {
    for (const match of segment.matchAll(WORD_PATTERN)) {
      words.push({
        index: words.length,
        text: match[0],
        segment: segmentIndex,
        start: match.index,
        end: match.index + match[0].length,
      });
    }
  });

  return words;
}
