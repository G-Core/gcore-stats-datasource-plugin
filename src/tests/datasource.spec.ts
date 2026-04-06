import { DataSourceInstanceSettings } from "@grafana/data";
import { DataSource } from "../datasource";
import { GCQuery, GCProduct, GCDataSourceOptions } from "../types";

function createMockInstanceSettings(
  overrides?: Partial<DataSourceInstanceSettings<GCDataSourceOptions>>
): DataSourceInstanceSettings<GCDataSourceOptions> {
  return {
    id: overrides?.id ?? 1,
    uid: overrides?.uid ?? "test-uid",
    type: overrides?.type ?? "gcore-stats-datasource-plugin",
    name: overrides?.name ?? "Gcore Unified",
    access: overrides?.access ?? "proxy",
    url: overrides?.url ?? "",
    jsonData: overrides?.jsonData ?? {},
    meta: overrides?.meta ?? ({} as DataSourceInstanceSettings["meta"]),
    readOnly: overrides?.readOnly ?? false,
    ...overrides,
  };
}

describe("DataSource filterQuery", () => {
  let ds: DataSource;

  beforeEach(() => {
    ds = new DataSource(createMockInstanceSettings());
  });

  it("filters out query when product is missing", () => {
    expect(ds.filterQuery({ refId: "A", hide: false } as GCQuery)).toBe(false);
  });

  it("filters out query when product is invalid", () => {
    expect(
      ds.filterQuery({
        refId: "A",
        hide: false,
        product: "invalid" as GCProduct,
      } as GCQuery)
    ).toBe(false);
  });

  it("allows CDN query when product is cdn and metric is set", () => {
    expect(
      ds.filterQuery({
        refId: "A",
        hide: false,
        product: "cdn",
        metric: { value: "total_bytes", label: "Total Bytes" },
      } as GCQuery)
    ).toBe(true);
  });

  it("filters out CDN query when metric is missing", () => {
    expect(
      ds.filterQuery({
        refId: "A",
        hide: false,
        product: "cdn",
      } as GCQuery)
    ).toBe(false);
  });

  it("allows DNS query when product is dns and zone is set", () => {
    expect(
      ds.filterQuery({
        refId: "A",
        hide: false,
        product: "dns",
        zone: "example.com",
      } as GCQuery)
    ).toBe(true);
  });

  it("allows FastEdge query when product is fastedge and required fields set", () => {
    expect(
      ds.filterQuery({
        refId: "A",
        hide: false,
        product: "fastedge",
        appId: 0,
        fastedgeMetric: "avg",
      } as GCQuery)
    ).toBe(true);
  });

  it("allows WAAP query when product is waap and waapMetric is set", () => {
    expect(
      ds.filterQuery({
        refId: "A",
        hide: false,
        product: "waap",
        waapMetric: "total_requests",
      } as GCQuery)
    ).toBe(true);
  });

  it("filters out query when hide is true", () => {
    expect(
      ds.filterQuery({
        refId: "A",
        hide: true,
        product: "cdn",
        metric: { value: "total_bytes", label: "Total Bytes" },
      } as GCQuery)
    ).toBe(false);
  });
});
