/* eslint-disable no-magic-numbers */
import { getUnit } from "../unit";
import { GCGranularity, GCServerMetric, GCUnit } from "../types";
import { preferredBandwidth, normalizedBandwidth } from "../bandwidth";
import { preferredBytes, normalizeBytes, ByteUnitEnum } from "../byte-converter";

describe("unit.getUnit", () => {
  it("returns neutral unit when metric/granularity missing", () => {
    const [unit, transform] = getUnit({ refId: "A" } as any, []);
    expect(unit).toBe(GCUnit.Number);
    expect(transform(123)).toBe(123);
  });

  it("chooses bytes unit based on max value and normalizes accordingly", () => {
    const query = {
      refId: "A",
      metric: { value: GCServerMetric.TotalBytes },
      granularity: { value: GCGranularity.OneHour },
    } as any;

    const data = [
      {
        metrics: {
          [GCServerMetric.TotalBytes]: [
            [1, 500],
            [2, 2000],
          ],
        },
      },
    ];

    const expectedUnit = preferredBytes(2000, ByteUnitEnum.B);
    const [unit, transform] = getUnit(query, data as any);
    expect(unit).toBe(expectedUnit);
    expect(transform(2000)).toBe(normalizeBytes(2000, ByteUnitEnum.B, expectedUnit));
  });

  it("chooses bandwidth unit based on max value and normalizes accordingly", () => {
    const query = {
      refId: "A",
      metric: { value: "bandwidth" },
      granularity: { value: GCGranularity.OneHour },
    } as any;

    const data = [
      {
        metrics: {
          [GCServerMetric.TotalBytes]: [
            [1, 500],
            [2, 2000],
          ],
        },
      },
    ];

    const period = 60 * 60;
    const expectedUnit = preferredBandwidth(2000, period);
    const [unit, transform] = getUnit(query, data as any);
    expect(unit).toBe(expectedUnit);
    expect(transform(2000)).toBe(normalizedBandwidth(2000, period, expectedUnit));
  });
});

