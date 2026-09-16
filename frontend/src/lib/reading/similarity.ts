import levenshtein from "damerau-levenshtein";

export const WORD_MIN_SIMILARITY = 0.79;
export const PHRASE_MIN_SIMILARITY = 0.6;

export const WORD_PATTERN = /[\p{L}\p{M}]+/gu;
export const WORD_SPLITTER = /[^\p{L}\p{M}]+/u;

export function splitWords(text: string): string[] {
  return text.split(WORD_SPLITTER).filter(Boolean);
}

export function wordSimilarity(a: string, b: string): number {
  return levenshtein(a.toLowerCase(), b.toLowerCase()).similarity;
}

export function isSimilarWord(a: string, b: string): boolean {
  return wordSimilarity(a, b) >= WORD_MIN_SIMILARITY;
}

export function isSimilarPhrase(words: string[], other: string[]): boolean {
  return (
    wordSimilarity(words.join(""), other.join("")) >= PHRASE_MIN_SIMILARITY
  );
}
