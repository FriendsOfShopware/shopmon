/** Median of the given values, or null for an empty list. */
export function median(values: number[]): number | null {
  if (values.length === 0) return null;
  const sorted = [...values].sort((a, b) => a - b);
  const middle = Math.floor(sorted.length / 2);
  return sorted.length % 2 === 0 ? (sorted[middle - 1] + sorted[middle]) / 2 : sorted[middle];
}

/**
 * Median of each value and the up to `window - 1` values before it. Nulls are
 * skipped, so a missing measurement does not pull the median down. The first
 * entries use the shorter window that exists, so the result starts at index 0.
 */
export function rollingMedian(values: (number | null)[], window: number): (number | null)[] {
  return values.map((_, index) => {
    const slice = values.slice(Math.max(0, index - window + 1), index + 1);
    return median(slice.filter((value): value is number => value != null));
  });
}
