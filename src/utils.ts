/**
 * Generic helpers for working with Grafana data frames, templating,
 * and string/number parsing that are reused across the datasource.
 */
import {
  FieldType,
  formatLabels,
  getDisplayProcessor,
  Labels,
  MutableField,
  TIME_SERIES_TIME_FIELD_NAME,
  TIME_SERIES_VALUE_FIELD_NAME,
  toDataFrame,
  ScopedVars,
  MetricFindValue,
} from "@grafana/data";
import { getTemplateSrv } from "@grafana/runtime";
import { GCPoint, GCQuery } from "./types";
import { TimeInSeconds } from "./times";

/**
 * Lightweight templating helper for legend/alias patterns.
 * Replaces `{{ key }}` placeholders with the values from `aliasData`.
 */
export const renderTemplate = (
  aliasPattern: string,
  aliasData: { [key: string]: string }
) => {
  const aliasRegex = /\{\{\s*(.+?)\s*\}\}/g;
  return aliasPattern.replace(aliasRegex, (_match, g1) => {
    if (aliasData[g1]) {
      return aliasData[g1];
    }
    return "";
  });
};

/**
 * Reads a comma-separated list of numbers from a template expression
 * (e.g. `$resources`) and returns them as an array of numbers.
 */
export const takeNumbers = (
  target?: string,
  scopedVars?: ScopedVars,
  format?: string | Function
): number[] => {
  return getTemplateSrv()
    .replace(target || "", scopedVars, format)
    .split(",")
    .filter(Boolean)
    .map((x) => +x);
};

/**
 * Reads a comma-separated list of strings from a template expression
 * (e.g. `$regions`) and returns them as an array of strings.
 */
export const takeStrings = (
  target?: string,
  scopedVars?: ScopedVars,
  format?: string | Function
): string[] => {
  return getTemplateSrv()
    .replace(target || "", scopedVars, format)
    .split(",")
    .filter(Boolean);
};

/**
 * Builds a Grafana time field from GCPoint tuples.
 * When `isMs` is false we treat timestamps as seconds and convert to ms.
 */
export const getTimeField = (data: GCPoint[], isMs = false): MutableField => ({
  name: TIME_SERIES_TIME_FIELD_NAME,
  type: FieldType.time,
  config: {},
  values: data.map((val) => (isMs ? val[0] : val[0] * TimeInSeconds.MS_PER_SECOND)),
});

/** Returns an empty data frame placeholder used when there is no data. */
export const getEmptyDataFrame = () => {
  return toDataFrame({
    name: "dataFrameName",
    fields: [],
  });
};

type ValueFieldOptions = {
  data: GCPoint[];
  valueName?: string;
  parseValue?: boolean;
  labels?: Labels;
  unit?: string;
  decimals?: number;
  displayNameFromDS?: string;
  transform?: (value: number) => number;
};

/**
 * Creates the numeric value field for a time series, applying unit,
 * decimals and an optional transform function.
 */
export const getValueField = ({
  data = [],
  valueName = TIME_SERIES_VALUE_FIELD_NAME,
  decimals = 2,
  labels,
  unit,
  displayNameFromDS,
  transform,
}: ValueFieldOptions): MutableField => ({
  labels,
  name: valueName,
  type: FieldType.number,
  display: getDisplayProcessor(),
  config: {
    unit,
    decimals,
    displayNameFromDS,
    displayName: displayNameFromDS,
  },
  values: data.map((val) => (transform ? transform(val[1]) : val[1])),
});

export interface LabelInfo {
  name: string;
  labels: Labels;
}

/**
 * Produces a friendly series name and labels based on the query's
 * legend format or the default Grafana label formatter.
 */
export const createLabelInfo = (
  labels: Labels,
  query: GCQuery,
  scopedVars: ScopedVars
): LabelInfo => {
  if (query?.legendFormat) {
    const title = renderTemplate(
      getTemplateSrv().replace(query.legendFormat, scopedVars),
      labels
    );
    return { name: title, labels };
  }

  const { metric, ...labelsWithoutMetric } = labels;
  const labelPart = formatLabels(labelsWithoutMetric);
  let title = `${metric} ${labelPart}`;
  return { name: title, labels: labelsWithoutMetric };
};

/**
 * Turns a raw list of values into Grafana variable options, making
 * sure duplicates are removed.
 */
export const getValueVariable = (
  target: Array<string | number>
): MetricFindValue[] => {
  const out: MetricFindValue[] = [];
  const seen = new Set<string>();
  for (const item of target) {
    const text = `${item}`;
    if (!seen.has(text)) {
      seen.add(text);
      out.push({ text });
    }
  }
  return out;
};
