import { describe, expect, it } from "vitest";
import { buildWordIndex } from "./word-index";
import { splitWords } from "./similarity";

describe("buildWordIndex", () => {
  it("indexes Unicode letters with segment offsets and skips punctuation", () => {
    const words = buildWordIndex(["Héllo, wörld! Привет—мир.", "naïve café"]);

    expect(words.map((w) => w.text)).toEqual([
      "Héllo",
      "wörld",
      "Привет",
      "мир",
      "naïve",
      "café",
    ]);
    expect(words.map((w) => w.index)).toEqual([0, 1, 2, 3, 4, 5]);
    expect(words.map((w) => [w.segment, w.start, w.end])).toEqual([
      [0, 0, 5],
      [0, 7, 12],
      [0, 14, 20],
      [0, 21, 24],
      [1, 0, 5],
      [1, 6, 10],
    ]);
  });

  it("keeps combining marks inside a word", () => {
    const words = buildWordIndex(["café au lait"]);
    expect(words[0]).toMatchObject({ text: "café", start: 0, end: 5 });
  });

  it("ignores segments without letters", () => {
    expect(buildWordIndex(["", "  ", "1234 --- ..."])).toEqual([]);
  });

  it("splits utterances with the same rule", () => {
    expect(splitWords(" we're  done, ok?! ")).toEqual(["we", "re", "done", "ok"]);
  });
});
