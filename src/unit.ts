/**
 * Helpers that decide which unit (bytes, bits/sec, etc.) should be used
 * for a given query and response payload and optionally provide a
 * transformation function to normalize values to that unit.
 */
import { GCQuery, GCResponseStats, GCUnit } from "./types";
import { normalizedBandwidth, preferredBandwidth } from "./bandwidth";
import { getSecondsByGranularity } from "./granularity";
import { createGetterYValues, getUnitByMetric } from "./metric";
import { ByteUnitEnum, normalizeBytes, preferredBytes } from "./byte-converter";

type Transform = (value: number) => number;
const noopTransform = (v: number): number => v;
const getMaxY = (
  data: GCResponseStats[],
  getter: (m: GCResponseStats["metrics"]) => number[]
): number => {
  let max = 0;
  for (const series of data) {
    const ys = getter(series.metrics);
    for (const y of ys) {
      if (y > max) max = y;
    }
  }
  return max;
};

/**
 * Computes the most appropriate unit for the series in `data` based on
 * the selected metric and granularity, and returns a pair
 * `[unit, transform]` where `transform` converts raw values to that unit.
 */
export const getUnit = (
  query: GCQuery,
  data: GCResponseStats[]
): [string, Transform] => {
  const metric = query.metric?.value;
  const granulation = query.granularity?.value;

  if (!metric || !granulation) {
    // Fallback: no metric or granularity selected, use neutral unit and identity transform
    return [GCUnit.Number, noopTransform];
  }
  const getter = createGetterYValues(metric);
  const rawUnit = getUnitByMetric(metric);

  if (rawUnit === GCUnit.Bandwidth) {
    const maxValue = getMaxY(data, getter);
    const period = getSecondsByGranularity(granulation);
    const unit = preferredBandwidth(maxValue, period);
    return [unit, (value) => normalizedBandwidth(value, period, unit)];
  } else if (rawUnit === GCUnit.Bytes) {
    const maxValue = getMaxY(data, getter);
    const unit = preferredBytes(maxValue, ByteUnitEnum.B);
    return [unit, (value) => normalizeBytes(value, ByteUnitEnum.B, unit)];
  } else {
    return [getUnitByMetric(metric), noopTransform];
  }
};
