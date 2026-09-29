import { describe, expect, it } from "vitest";
import { median, rollingMedian } from "./statistics";

describe("median", () => {
  it("returns null for an empty list", () => {
    expect(median([])).toBeNull();
  });

  it("returns the middle value of an odd count", () => {
    expect(median([300, 100, 200])).toBe(200);
  });

  it("averages the two middle values of an even count", () => {
    expect(median([400, 100, 300, 200])).toBe(250);
  });
});

describe("rollingMedian", () => {
  it("uses the shorter window at the start of the series", () => {
    expect(rollingMedian([100, 300, 200, 900, 400], 3)).toEqual([100, 200, 200, 300, 400]);
  });

  it("ignores a single outlier", () => {
    expect(rollingMedian([100, 100, 5000, 100, 100], 3)).toEqual([100, 100, 100, 100, 100]);
  });

  it("skips nulls instead of counting them as zero", () => {
    expect(rollingMedian([100, null, 300], 3)).toEqual([100, 100, 200]);
  });

  it("returns null for a window that holds only nulls", () => {
    expect(rollingMedian([null, null, 100], 2)).toEqual([null, null, 100]);
  });
});
