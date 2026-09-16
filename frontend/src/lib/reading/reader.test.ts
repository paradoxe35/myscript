import { describe, expect, it } from "vitest";
import { ScriptReader, WIDE_WINDOW, WINDOW } from "./reader";
import { splitWords } from "./similarity";

const script = splitWords(
  "The quick brown fox jumps over the lazy dog. " +
    "Pack my box with five dozen liquor jugs. " +
    "How vexingly quick daft zebras jump."
);

const filler = (n: number) =>
  Array.from({ length: n }, (_, i) => `filler${i}`).join(" ");

describe("ScriptReader", () => {
  it("advances past an exact match", () => {
    const reader = new ScriptReader(script);
    expect(reader.feed("The quick brown fox jumps over the lazy dog")).toEqual({
      from: 0,
      to: 9,
    });
    expect(reader.position).toBe(9);
  });

  it("still matches a misheard word", () => {
    const reader = new ScriptReader(script);
    reader.feed("the quick brawn fox jumps over the lazy dog");
    expect(reader.position).toBe(9);
  });

  it("tolerates a dropped word", () => {
    const reader = new ScriptReader(script);
    reader.feed("the quick brown fox over the lazy dog");
    expect(reader.position).toBe(9);
  });

  it("tolerates an inserted word", () => {
    const reader = new ScriptReader(script);
    reader.feed("the quick um brown fox jumps over the lazy dog");
    expect(reader.position).toBe(9);
  });

  it("stays put when nothing matches", () => {
    const reader = new ScriptReader(script);
    reader.feed("The quick brown fox jumps over the lazy dog");
    expect(reader.feed("completely unrelated words here")).toBeNull();
    expect(reader.position).toBe(9);
  });

  it("strips bracketed annotations and ignores empty utterances", () => {
    const reader = new ScriptReader(script);
    expect(reader.feed("[BLANK_AUDIO]")).toBeNull();
    expect(reader.feed("   ")).toBeNull();
    expect(reader.position).toBe(0);
  });

  it("never moves backward on its own", () => {
    const reader = new ScriptReader(script);
    reader.feed("Pack my box with five dozen liquor jugs");
    expect(reader.position).toBe(17);
    expect(reader.feed("the quick brown fox jumps over the lazy dog")).toBeNull();
    expect(reader.position).toBe(17);
  });

  it("does not jump to a common word beyond the window", () => {
    const words = splitWords(`cat sleeps ${filler(WINDOW + 10)} the end`);
    const reader = new ScriptReader(words);
    expect(reader.feed("the")).toBeNull();
    expect(reader.position).toBe(0);
  });

  it("widens the window once after a second miss", () => {
    const far = "zebras graze quietly beside the river";
    const words = splitWords(`${filler(60)} ${far} ${filler(100)}`);
    const reader = new ScriptReader(words);

    expect(reader.feed(far)).toBeNull();
    expect(reader.feed(far)).toEqual({ from: 60, to: 66 });
    expect(reader.position).toBe(66);
  });

  it("stays put when even the wide window has no match", () => {
    const far = "zebras graze quietly beside the river";
    const words = splitWords(`${filler(WIDE_WINDOW + 20)} ${far}`);
    const reader = new ScriptReader(words);

    reader.feed(far);
    reader.feed(far);
    expect(reader.position).toBe(0);
  });

  it("moves to a clicked word and resumes from there", () => {
    const reader = new ScriptReader(script);
    reader.moveTo(17);
    expect(reader.position).toBe(17);

    reader.feed("how vexingly quick daft zebras jump");
    expect(reader.position).toBe(script.length);
    expect(reader.done).toBe(true);

    reader.moveTo(2);
    expect(reader.position).toBe(2);
    reader.moveTo(-5);
    expect(reader.position).toBe(0);
    reader.moveTo(999);
    expect(reader.position).toBe(script.length);
  });
});
