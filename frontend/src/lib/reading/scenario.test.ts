import { describe, expect, it } from "vitest";
import { ScriptReader, WINDOW } from "./reader";
import {
  PHRASE_MIN_SIMILARITY,
  WORD_MIN_SIMILARITY,
  splitWords,
  wordSimilarity,
} from "./similarity";

const sentences = [
  "Good evening, and welcome to the show.",
  "Tonight we're talking about the café at the end of the street.",
  "Its owner, Zoë, opened it in nineteen ninety-eight.",
  "Every morning she bakes fresh bread before the sun comes up.",
  "Regulars say the smell alone is worth the walk.",
  "Она говорит, что секрет в терпении.",
  "Later this week, the council votes on the new market square.",
  "We'll have the full story after the break.",
];

const script = splitWords(sentences.join(" "));

// Word count of the script up to and including sentence n.
const endOf = (n: number) =>
  sentences.slice(0, n + 1).reduce((sum, s) => sum + splitWords(s).length, 0);

describe("reading a script the way a speech engine returns it", () => {
  const reader = new ScriptReader(script);

  it("exact sentence", () => {
    reader.feed("Good evening and welcome to the show");
    expect(reader.position).toBe(endOf(0));
  });

  it("one misheard word", () => {
    reader.feed("Tonight we're taking about the café at the end of the street");
    expect(reader.position).toBe(endOf(1));
  });

  it("a dropped word", () => {
    reader.feed("Its owner Zoë opened it nineteen ninety eight");
    expect(reader.position).toBe(endOf(2));
  });

  it("an inserted filler word", () => {
    reader.feed("Every morning um she bakes fresh bread before the sun comes up");
    expect(reader.position).toBe(endOf(3));
  });

  it("a sentence split across two utterances", () => {
    reader.feed("Regulars say the smell");
    expect(reader.position).toBe(endOf(3) + 4);
    reader.feed("alone is worth the walk");
    expect(reader.position).toBe(endOf(4));
  });

  it("two sentences merged into one utterance", () => {
    reader.feed(
      "Она говорит что секрет в терпении later this week the council votes on the new market square"
    );
    expect(reader.position).toBe(endOf(6));
  });

  it("a repeated sentence after a stumble does not move back", () => {
    reader.feed("later this week the council votes on the new market square");
    expect(reader.position).toBe(endOf(6));
  });

  it("pure noise", () => {
    reader.feed("uh hmm [BLANK_AUDIO]");
    expect(reader.position).toBe(endOf(6));
  });

  it("finishes the page", () => {
    reader.feed("We'll have the full story after the break");
    expect(reader.position).toBe(endOf(7));
    expect(reader.position).toBe(script.length);
    expect(reader.done).toBe(true);
  });
});

describe("no jump on a common word", () => {
  const next = "the cat sleeps on my bed";
  const filler = Array.from({ length: WINDOW + 20 }, (_, i) => `w${i}`).join(" ");
  const words = splitWords(`${next} ${filler} the dog barks`);

  it("advances to the next sentence, not the far occurrence", () => {
    const reader = new ScriptReader(words);
    reader.feed("the cat sleeps on my bed");
    expect(reader.position).toBe(6);
  });

  it("a lone common word cannot jump past the window", () => {
    const reader = new ScriptReader(words);
    reader.feed("the cat sleeps on my bed");
    expect(reader.feed("the")).toBeNull();
    expect(reader.position).toBe(6);
  });

  it("leaves the position unchanged when nothing matches", () => {
    const reader = new ScriptReader(words);
    reader.feed("the cat sleeps on my bed");
    reader.feed("zebras graze quietly");
    expect(reader.position).toBe(6);
  });
});

describe("no regression against the utterance-level matcher", () => {
  const text = "the quick brown fox jumps over the lazy dog";
  const words = splitWords(text);

  it("keeps the old thresholds", () => {
    expect(WORD_MIN_SIMILARITY).toBe(0.79);
    expect(PHRASE_MIN_SIMILARITY).toBe(0.6);
  });

  it("accepts a near-miss word at the word threshold", () => {
    expect(wordSimilarity("brown", "brawn")).toBeGreaterThanOrEqual(
      WORD_MIN_SIMILARITY
    );
    const reader = new ScriptReader(words);
    reader.feed("the quick brawn fox jumps over the lazy dog");
    expect(reader.position).toBeGreaterThanOrEqual(words.length);
  });

  it("accepts a phrase above the phrase threshold with words below it", () => {
    expect(wordSimilarity("quick", "kwick")).toBeLessThan(WORD_MIN_SIMILARITY);
    expect(wordSimilarity("brown", "braun")).toBeLessThan(WORD_MIN_SIMILARITY);
    const reader = new ScriptReader(words);
    reader.feed("the kwick braun fox jumps over the lazy dog");
    expect(reader.position).toBeGreaterThanOrEqual(words.length);
  });

  it("rejects a phrase below the phrase threshold", () => {
    const reader = new ScriptReader(words);
    reader.feed("the kwick braun fux jamps ovar tha lasy dug");
    expect(reader.position).toBe(0);
  });

  it("resumes from a stored position", () => {
    const reader = new ScriptReader(script);
    reader.moveTo(endOf(3));
    reader.feed("Regulars say the smell alone is worth the walk");
    expect(reader.position).toBeGreaterThanOrEqual(endOf(4));
  });
});

describe("any error-free prefix ends exactly at its end", () => {
  for (let count = 1; count <= sentences.length; count++) {
    it(`${count} sentence(s)`, () => {
      const reader = new ScriptReader(script);
      sentences.slice(0, count).forEach((s) => reader.feed(s));
      expect(reader.position).toBe(endOf(count - 1));
    });
  }

  it("one utterance per word", () => {
    const reader = new ScriptReader(script);
    script.forEach((word, i) => {
      reader.feed(word);
      expect(reader.position).toBe(i + 1);
    });
  });
});
