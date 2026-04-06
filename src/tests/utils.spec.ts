/* eslint-disable no-magic-numbers */
import { getValueVariable } from "../utils";

describe("utils.getValueVariable", () => {
  it("dedupes values and preserves first-seen order", () => {
    const out = getValueVariable(["a", "b", "a", "c", "b"]);
    expect(out).toEqual([{ text: "a" }, { text: "b" }, { text: "c" }]);
  });

  it("dedupes across mixed number/string inputs via stringification", () => {
    const out = getValueVariable([1, "1", 2, "2", 1]);
    expect(out).toEqual([{ text: "1" }, { text: "2" }]);
  });
});

