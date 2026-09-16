import { isSimilarPhrase, isSimilarWord, splitWords } from "./similarity";

export const WINDOW = 40;
export const WIDE_WINDOW = 120;

const MATCH = 1;
const MISMATCH = -0.5;
const GAP = -0.6;
const MIN_SCORE = 1;

const ANNOTATION = /\[.*?\]/g;

export type WordRange = { from: number; to: number };

// Semi-global alignment: the whole utterance against any stretch of the window,
// so misheard, dropped and inserted words on either side only cost points.
export function alignUtterance(
  script: string[],
  utterance: string[]
): WordRange | null {
  const m = utterance.length;
  const n = script.length;
  if (m === 0 || n === 0) return null;

  let score = new Array<number>(n + 1).fill(0);
  let start = Array.from({ length: n + 1 }, (_, j) => j);

  for (let i = 1; i <= m; i++) {
    const nextScore = new Array<number>(n + 1);
    const nextStart = new Array<number>(n + 1);
    nextScore[0] = score[0] + GAP;
    nextStart[0] = 0;

    for (let j = 1; j <= n; j++) {
      const similar = isSimilarWord(utterance[i - 1], script[j - 1]);
      const diagonal = score[j - 1] + (similar ? MATCH : MISMATCH);
      const skipUtterance = score[j] + GAP;
      const skipScript = nextScore[j - 1] + GAP;

      if (diagonal >= skipUtterance && diagonal >= skipScript) {
        nextScore[j] = diagonal;
        nextStart[j] = start[j - 1];
      } else if (skipUtterance >= skipScript) {
        nextScore[j] = skipUtterance;
        nextStart[j] = start[j];
      } else {
        nextScore[j] = skipScript;
        nextStart[j] = nextStart[j - 1];
      }
    }

    score = nextScore;
    start = nextStart;
  }

  let best = 0;
  for (let j = 1; j <= n; j++) {
    if (score[j] > score[best]) best = j;
  }

  const range = { from: start[best], to: best };
  if (score[best] < MIN_SCORE || range.to <= range.from) return null;
  if (!isSimilarPhrase(script.slice(range.from, range.to), utterance)) {
    return null;
  }

  return range;
}

export class ScriptReader {
  position = 0;
  private misses = 0;

  constructor(readonly words: string[]) {}

  get total() {
    return this.words.length;
  }

  get done() {
    return this.position >= this.total;
  }

  moveTo(index: number) {
    this.position = Math.max(0, Math.min(index, this.total));
    this.misses = 0;
  }

  feed(text: string): WordRange | null {
    const utterance = splitWords(text.replace(ANNOTATION, ""));
    if (utterance.length === 0) return null;

    let range = this.alignAhead(utterance, WINDOW);
    if (!range && this.misses > 0) {
      range = this.alignAhead(utterance, WIDE_WINDOW);
    }

    if (!range) {
      this.misses++;
      return null;
    }

    this.misses = 0;
    this.position = range.to;
    return range;
  }

  private alignAhead(utterance: string[], window: number) {
    const end = Math.min(this.total, this.position + window + utterance.length);
    const range = alignUtterance(
      this.words.slice(this.position, end),
      utterance
    );

    return range
      ? { from: this.position + range.from, to: this.position + range.to }
      : null;
  }
}
