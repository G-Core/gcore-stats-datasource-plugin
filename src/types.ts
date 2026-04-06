/**
 * Shared type definitions used by both the frontend query editor and the
 * backend datasource implementation.
 */
import { DataQuery, DataSourceJsonData, SelectableValue } from "@grafana/data";

/** Product / service type for the unified Gcore datasource. */
export type GCProduct = "cdn" | "dns" | "fastedge" | "waap";

export const GC_PRODUCTS: Array<{ value: GCProduct; label: string }> = [
  { value: "cdn", label: "CDN" },
  { value: "dns", label: "DNS" },
  { value: "fastedge", label: "FastEdge" },
  { value: "waap", label: "WAAP" },
];

/** CDN metrics returned by the `/cdn/statistics` APIs. */
export enum GCServerMetric {
  UpstreamBytes = "upstream_bytes",
  TotalBytes = "total_bytes",
  SentBytes = "sent_bytes",
  ShieldBytes = "shield_bytes",
  Requests = "requests",
  Responses2xx = "responses_2xx",
  Responses3xx = "responses_3xx",
  Responses4xx = "responses_4xx",
  Responses5xx = "responses_5xx",
  CacheHitRequestsRatio = "cache_hit_requests_ratio",
  CacheHitTrafficRatio = "cache_hit_traffic_ratio",
  ShieldTrafficRatio = "shield_traffic_ratio",
  ImageProcessed = "image_processed",
}

/** CDN client-side metrics (derived from server metrics). */
export enum GCClientMetric {
  Bandwidth = "bandwidth",
}

export type GCMetric = GCServerMetric | GCClientMetric;

/** CDN grouping dimensions (for `product === "cdn"`). */
export enum GCGrouping {
  Resource = "resource",
  Region = "region",
  VHost = "vhost",
  Client = "client",
  Country = "country",
  DC = "dc",
}

/** CDN time bucket size (granularity for `product === "cdn"`). */
export enum GCGranularity {
  FiveMinutes = "5m",
  FifteenMinutes = "15m",
  OneHour = "1h",
  OneDay = "1d",
}

/**
 * DNS granularity (separate from CDN `GCGranularity`, used when
 * `product === "dns"`).
 */
export enum GCDNSGranularity {
  FiveMinutes = "5m",
  TenMinutes = "10m",
  FifteenMinutes = "15m",
  ThirtyMinutes = "30m",
  OneHour = "1h",
  NinetyMinutes = "1.5h",
  TwoHoursFortyFiveMinutes = "2h45m",
  OneDay = "24h",
}

/** DNS record type for DNS zone statistics (`product === "dns"`). */
export enum GCDNSRecordType {
  All = "ALL",
  A = "A",
  AAAA = "AAAA",
  NS = "NS",
  CNAME = "CNAME",
  MX = "MX",
  TXT = "TXT",
  SVCB = "SVCB",
  HTTPS = "HTTPS",
}

/** WAAP granularity (time bucket size when `product === "waap"`). */
export type GCWaapGranularity = "1h" | "1d";

export interface GCQuery extends DataQuery {
  /**
   * Which Gcore product this query targets:
   * `"cdn" | "dns" | "fastedge" | "waap"`.
   */
  product?: GCProduct;

  // CDN Edge
  // --- CDN fields 
  metric?: SelectableValue<GCMetric>;
  granularity?: SelectableValue<GCGranularity>;
  grouping?: Array<SelectableValue<GCGrouping>>;
  vhosts?: string;
  resources?: string;
  countries?: string;
  regions?: string;
  clients?: string;
  legendFormat?: string;

  // DNS Edge
  // --- DNS fields 
  zone?: string;
  record_type?: GCDNSRecordType | string;
  dnsGranularity?: SelectableValue<GCDNSGranularity | string>;
  dnsLegendFormat?: string;

  // FastEdge Edge
  // --- FastEdge fields 
  appId?: number;
  appName?: string;
  step?: number;
  network?: string;
  fastedgeMetric?: string;

  // WAAP Edge
  // --- WAAP fields 
  waapMetric?: string;
  waapGranularity?: GCWaapGranularity;
  waapLegendFormat?: string;
}

export enum GCUnit {
  Number = "none",
  Bandwidth = "bit/sec",
  Bytes = "bytes",
  Percent = "none",
}

export interface GCStatsRequestData {
  metrics: GCServerMetric[];
  vhosts?: string[];
  regions?: string[];
  resources?: number[];
  clients?: number[];
  countries?: string[];
  from: string;
  to: string;
  granularity: GCGranularity;
  flat: boolean;
  group_by?: GCGrouping[];
}

export interface GCCdnResource {
  id: number;
  cname: string;
  client: number;
}

export interface GCDataSourceOptions extends DataSourceJsonData {
  path?: string;
  apiUrl?: string;
}

export interface GCSecureJsonData {
  apiKey?: string;
}

export interface GCJsonData {
  apiUrl?: string;
}

export enum GCVariable {
  // CDN
  Resource = "resource",
  Client = "client",
  Vhost = "vhost",
  Region = "region",
  Country = "country",
  Datacenter = "datacenter",
  // DNS
  Zone = "zone",
  RecordType = "record_type",
  // FastEdge (e.g. app list )
  App = "app",
}

export interface GCVariableQuery {
  selector: SelectableValue<GCVariable>;
}

export type GCPoint = [number, number];

export interface GCResponseStats {
  metrics: Partial<Record<GCServerMetric, GCPoint[]>>;
  client?: number;
  region?: string;
  vhost?: string;
  country?: string;
  dc?: string;
  resource?: number;
}

export interface Paginator<T> {
  count: number;
  next: string;
  previous: string;
  results: T[];
}
