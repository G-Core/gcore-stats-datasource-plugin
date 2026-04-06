/**
 * Grafana datasource implementation for querying multiple Gcore products
 * (CDN, DNS, FastEdge, WAAP) through a single unified plugin.
 */
import {
  DataSourceInstanceSettings,
  MetricFindValue,
} from "@grafana/data";
import { DataSourceWithBackend, getBackendSrv } from "@grafana/runtime";
import { range } from "lodash";
import {
  GCCdnResource,
  GCDataSourceOptions,
  GCProduct,
  GCQuery,
  GCVariable,
  GCVariableQuery,
  GCDNSRecordType,
  Paginator,
} from "./types";
import { getValueVariable } from "./utils";
import { regions } from "./regions";
import { countries } from "./countries";

const VALID_PRODUCTS: GCProduct[] = ["cdn", "dns", "fastedge", "waap"];

/**
 * Type-safe guard that checks if the provided product string is supported
 * by this datasource.
 */
function isValidProduct(p?: string): p is GCProduct {
  return !!p && VALID_PRODUCTS.includes(p as GCProduct);
}

/**
 * Main datasource class wired into Grafana. It delegates all HTTP traffic
 * to the backend implementation while keeping UI logic and variable support
 */
export class DataSource extends DataSourceWithBackend<GCQuery, GCDataSourceOptions> {
  id: number;

  constructor(instanceSettings: DataSourceInstanceSettings<GCDataSourceOptions>) {
    super(instanceSettings);
    this.id = instanceSettings.id;
  }

  /**
   * Basic connectivity/authentication check used by Grafana in the
   * datasource configuration page.
   *
   * Newer stacks expose the IAM endpoint, older ones only `users/me`.
   * We try IAM first and gracefully fall back to the legacy endpoint.
   */
  async testDatasource() {
    try {
      const r1 = await this.getResource("iam/users/me");
      const name = (r1 as { name?: string })?.name;
      return {
        status: "success" as const,
        message: `Authentication successful (IAM): ${name ?? "OK"}`,
      };
    } catch {
      try {
        const r2 = await this.getResource("users/me");
        const name = (r2 as { name?: string })?.name;
        return {
          status: "success" as const,
          message: `Authentication successful: ${name ?? "OK"}`,
        };
      } catch (err: unknown) {
        const e = err as { data?: { error?: string; detail?: string; message?: string }; status?: number };
        const errorMsg =
          e?.data?.error ??
          e?.data?.detail ??
          e?.data?.message ??
          (typeof e?.data === "string" ? e.data : JSON.stringify(e?.data ?? (err as Error)?.message ?? "")) ??
          "Unknown error";
        if (e?.status === 401) {
          return {
            status: "error" as const,
            message: `Invalid API key. Please verify your credentials in the datasource settings. Details: ${errorMsg}`,
          };
        }
        if (e?.status === 500 || e?.status === 502) {
          return {
            status: "error" as const,
            message:
              String(errorMsg) ||
              "Gcore API returned a server error. Please try again later or verify the API URL.",
          };
        }
        return {
          status: "error" as const,
          message: String(errorMsg) || "Failed to authenticate. Check URL, API key, or network.",
        };
      }
    }
  }

  /**
   * Implements Grafana's variable support (`Query type` variables).
   * The selector inside `GCVariableQuery` determines which list of values
   * should be returned (CDN, DNS, FastEdge).
   */
  async metricFindQuery(query: GCVariableQuery): Promise<MetricFindValue[]> {
    if (!query.selector) {
      return [];
    }

    const selector = query.selector.value;

    // CDN variables (vhost / resource / client)
    if (
      selector === GCVariable.Client ||
      selector === GCVariable.Resource ||
      selector === GCVariable.Vhost
    ) {
      const cdnResource: GCCdnResource[] = await this.getAllGCCdnResources();
      switch (selector) {
        case GCVariable.Vhost:
          return getValueVariable(cdnResource.map((item) => item.cname));
        case GCVariable.Resource:
          return getValueVariable(cdnResource.map((item) => item.id));
        case GCVariable.Client:
          return getValueVariable(cdnResource.map((item) => item.client));
      }
    }
    if (selector === GCVariable.Region) {
      return getValueVariable(regions);
    }
    if (selector === GCVariable.Country) {
      return getValueVariable(countries);
    }

    // DNS variables
    if (selector === GCVariable.Zone) {
      try {
        const limit = 1000;
        let offset = 0;
        const all: { name: string }[] = [];
        let total: number | undefined;

        // Attempt to paginate if the endpoint supports `limit/offset` and `total_amount`.
        // Falls back gracefully to first page behavior if not supported.
        for (;;) {
          const raw = await this.getResource("dns/zones", { limit, offset } as any);
          const data = raw as { zones?: { name: string }[]; total_amount?: number };
          const zones = data?.zones ?? [];
          if (typeof data?.total_amount === "number") {
            total = data.total_amount;
          }

          all.push(...zones);
          if (zones.length === 0) break;
          if (total != null && all.length >= total) break;

          // If total is unknown, we stop after first page to avoid an infinite loop.
          if (total == null) break;
          offset += limit;
        }

        return getValueVariable(all.map((z) => z.name));
      } catch {
        return [];
      }
    }
    if (selector === GCVariable.RecordType) {
      return getValueVariable(Object.values(GCDNSRecordType));
    }

    // FastEdge variables: application list
    if (selector === GCVariable.App) {
      try {
        const raw = await this.getResource("fastedge/apps");
        const arr =
          (Array.isArray(raw) ? raw : (raw as { apps?: unknown[] })?.apps ?? []) as Array<{
            id?: number;
            name?: string;
          }>;

        // MetricFindValue supports optional `value`. We return `text` for display
        // (and templating) while keeping `value` as the numeric app id so the
        // query editor can use it without re-fetching.
        return arr.map((app) => ({
          text: app?.name ?? `App ${app?.id ?? "?"}`,
          value: app?.id,
        }));
      } catch {
        return [];
      }
    }

    return [];
  }

  /**
   * Loads the complete list of CDN resources using the backend proxy.
   * The API is paginated, so we fetch the first page, compute the total
   * amount and then request all remaining pages with capped concurrency.
   */
  private async getAllGCCdnResources(): Promise<GCCdnResource[]> {
    const getGCCdnResources = (limit: number, offset = 0) =>
      getBackendSrv().datasourceRequest<Paginator<GCCdnResource>>({
        method: "GET",
        url: `api/datasources/${this.id}/cdn/resources`,
        responseType: "json",
        showErrorAlert: true,
        params: {
          fields: "id,cname,client",
          deleted: true,
          limit,
          offset,
        },
      }).catch((error: { status?: number }) => {
        if (error?.status === 401) {
          throw new Error("Unauthorized: Invalid API Key. Please check your datasource configuration.");
        }
        throw error;
      });

    // 1000 is a good compromise between number of HTTP calls and
    // response payload size.
    const limit = 1000;
    const firstChunk = await getGCCdnResources(limit);
    const data = (firstChunk as { data?: Paginator<GCCdnResource> }).data ?? (firstChunk as unknown as Paginator<GCCdnResource>);
    const cdnResourcesCount = data.count ?? 0;
    const results = data.results ?? [];

    if (cdnResourcesCount <= limit) {
      return results;
    }

    const offsets = range(limit, cdnResourcesCount, limit);
    const concurrency = 6;
    const acc: GCCdnResource[] = [...results];

    for (let i = 0; i < offsets.length; i += concurrency) {
      const batch = offsets.slice(i, i + concurrency);
      const batchChunks = await Promise.all(
        batch.map((offset) => getGCCdnResources(limit, offset))
      );
      for (const res of batchChunks) {
        const d =
          (res as { data?: Paginator<GCCdnResource> }).data ??
          (res as unknown as Paginator<GCCdnResource>);
        acc.push(...(d.results ?? []));
      }
    }

    return acc;
  }

  /**
   * Helper used by the query editor to determine whether a query
   * is complete enough to be executed and shown on the panel.
   */
  filterQuery(query: GCQuery): boolean {
    if (query.hide) return false;
    if (!isValidProduct(query.product)) return false;

    switch (query.product) {
      case "cdn":
        return !!query.metric?.value;
      case "dns":
        return !!query.zone;
      case "fastedge":
        return (
          (query.appId === 0 || (query.appId != null && query.appId > 0)) &&
          !!query.fastedgeMetric
        );
      case "waap":
        return !!query.waapMetric;
      default:
        return false;
    }
  }
}