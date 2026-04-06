/**
 * Collection of default query templates used by the UI when a user switches
 * between different Gcore products (CDN, DNS, FastEdge, WAAP).
 */
import { createOptionForMetric } from "./metric";
import { createOptionForGrouping } from "./grouping";
import { createOptionForGranularity } from "./granularity";
import {
  GCGranularity,
  GCGrouping,
  GCQuery,
  GCServerMetric,
  GCVariable,
  GCVariableQuery,
  GCDNSGranularity,
  GCDNSRecordType,
} from "./types";
/** Defaults applied when the selected product is CDN. */
export const DefaultCdnQuery: Partial<GCQuery> = {
  product: "cdn",
  vhosts: "",
  clients: "",
  resources: "",
  metric: createOptionForMetric(GCServerMetric.TotalBytes),
  grouping: [createOptionForGrouping(GCGrouping.Resource)],
  granularity: createOptionForGranularity(GCGranularity.OneHour),
};

// Backward-compatible alias used by GCQueryEditor and other components.
export const defaultQuery: Partial<GCQuery> = DefaultCdnQuery;

/** Defaults applied when the selected product is DNS. */
export const defaultDNSQuery: Partial<GCQuery> = {
  product: "dns",
  zone: "all",
  dnsGranularity: { value: GCDNSGranularity.FiveMinutes, label: "5m" },
  record_type: GCDNSRecordType.All,
};

/** Defaults applied when the selected product is FastEdge. */
export const defaultFastEdgeQuery: Partial<GCQuery> = {
  product: "fastedge",
  appId: 0,
  appName: "All",
  step: 60,
  network: "",
  fastedgeMetric: "avg",
};

/** Defaults applied when the selected product is WAAP. */
export const defaultWAAPQuery: Partial<GCQuery> = {
  product: "waap",
  waapMetric: "total_requests",
  waapGranularity: "1h",
};

/** Default template for query variables (templating editor). */
export const defaultVariableQuery: Partial<GCVariableQuery> = {
  selector: { value: GCVariable.Resource, label: "resourceID" },
};
