import defaults from "lodash/defaults";
import React, { ChangeEvent, useCallback, useEffect, useState } from "react";
import { LegacyForms, Select, Spinner } from "@grafana/ui";
import { QueryEditorProps, SelectableValue } from "@grafana/data";
import { DataSource } from "../datasource";
import {
  GCDataSourceOptions,
  GCProduct,
  GCGranularity,
  GCGrouping,
  GCQuery,
  GCMetric,
  GC_PRODUCTS,
  GCDNSGranularity,
  GCDNSRecordType,
  GCVariable,
} from "../types";
import { GCSelectMetric } from "./GCSelectMetric";
import { GCSelectGranularity } from "./GCSelectGranularity";
import { GCSelectGrouping } from "./GCSelectGrouping";
import { GCInput } from "./GCInput";
import {
  defaultQuery,
  defaultDNSQuery,
  defaultFastEdgeQuery,
  defaultWAAPQuery,
} from "../defaults";

const { FormField } = LegacyForms;

const DNS_GRANULARITY_OPTS: Array<SelectableValue<string>> = Object.values(
  GCDNSGranularity
).map((g) => ({ value: g, label: g }));

const DNS_RECORD_TYPE_OPTS: Array<SelectableValue<string>> = Object.values(
  GCDNSRecordType
).map((r) => ({ value: r, label: r }));

const WAAP_METRIC_OPTS: Array<SelectableValue<string>> = [
  { label: "total_requests", value: "total_requests" },
  { label: "total_bytes", value: "total_bytes" },
];

const WAAP_GRANULARITY_OPTS: Array<SelectableValue<string>> = [
  { label: "1h", value: "1h" },
  { label: "1d", value: "1d" },
];

const FASTEDGE_METRIC_OPTS: Array<SelectableValue<string>> = [
  { label: "avg", value: "avg" },
  { label: "min", value: "min" },
  { label: "max", value: "max" },
  { label: "median", value: "median" },
  { label: "perc75", value: "perc75" },
  { label: "perc90", value: "perc90" },
];

type Props = QueryEditorProps<DataSource, GCQuery, GCDataSourceOptions>;

export const GCQueryEditor = ({
  query,
  datasource,
  onChange,
  onRunQuery,
}: Props) => {
  const [dnsZones, setDnsZones] = useState<Array<SelectableValue<string>>>([]);
  const [dnsZonesLoading, setDnsZonesLoading] = useState(false);
  const [fastedgeApps, setFastedgeApps] = useState<Array<SelectableValue<number>>>(
    []
  );
  const [fastedgeAppsLoading, setFastedgeAppsLoading] = useState(false);

  const product = (query?.product ?? "cdn") as GCProduct;

  const loadDnsZones = useCallback(async () => {
    setDnsZonesLoading(true);
    try {
      const result = await datasource.metricFindQuery({
        selector: { value: GCVariable.Zone },
      });
      const zones: SelectableValue<string>[] = (
        result as Array<{ text: string }>
      ).map((z) => ({ value: z.text, label: z.text }));
      setDnsZones([{ value: "all", label: "All Zones" }, ...zones]);
    } catch {
      setDnsZones([{ value: "all", label: "All Zones" }]);
    } finally {
      setDnsZonesLoading(false);
    }
  }, [datasource]);

  const loadFastEdgeApps = useCallback(async () => {
    setFastedgeAppsLoading(true);
    try {
      const result = await datasource.metricFindQuery({
        selector: { value: GCVariable.App },
      });

      const arr = result as Array<{ text: string; value?: number }>;
      setFastedgeApps([
        { label: "All", value: 0 },
        ...arr
          .filter((app) => typeof app?.value === "number" && app.value >= 0)
          .map((app) => ({
            label: app.text,
            value: app.value as number,
          })),
      ]);
    } catch {
      setFastedgeApps([{ label: "All", value: 0 }]);
    } finally {
      setFastedgeAppsLoading(false);
    }
  }, [datasource]);

  useEffect(() => {
    if (!query?.product) {
      onChange({ ...defaults(query, defaultQuery), product: "cdn" });
      onRunQuery?.();
      return;
    }

    if (product === "dns" && !query.dnsGranularity) {
      onChange({ ...query, ...defaultDNSQuery } as GCQuery);
      onRunQuery?.();
      return;
    }

    if (product === "fastedge" && query.fastedgeMetric === undefined) {
      onChange({ ...query, ...defaultFastEdgeQuery } as GCQuery);
      onRunQuery?.();
      return;
    }

    if (product === "waap" && !query.waapMetric) {
      onChange({ ...query, ...defaultWAAPQuery } as GCQuery);
      onRunQuery?.();
      return;
    }

    if (product === "dns") {
      void loadDnsZones();
    }
    if (product === "fastedge") {
      void loadFastEdgeApps();
    }
  }, [loadDnsZones, loadFastEdgeApps, onChange, onRunQuery, product, query]);

  const onProductChange = (opt: SelectableValue<GCProduct>) => {
    const nextProduct = (opt?.value ?? "cdn") as GCProduct;
    const next: Partial<GCQuery> = { ...query, product: nextProduct };
    if (nextProduct === "cdn") Object.assign(next, defaultQuery);
    else if (nextProduct === "dns") Object.assign(next, defaultDNSQuery);
    else if (nextProduct === "fastedge") Object.assign(next, defaultFastEdgeQuery);
    else if (nextProduct === "waap") Object.assign(next, defaultWAAPQuery);
    onChange(next as GCQuery);
    onRunQuery?.();
  };

  const onMetricChange = (value: SelectableValue<GCMetric>) => {
    onChange({ ...query, metric: value });
    onRunQuery?.();
  };
  const onGroupingChange = (value: Array<SelectableValue<GCGrouping>>) => {
    onChange({ ...query, grouping: value });
    onRunQuery?.();
  };
  const onGranularityChange = (value: SelectableValue<GCGranularity>) => {
    onChange({ ...query, granularity: value });
    onRunQuery?.();
  };
  const onVhostsChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, vhosts: e.target.value });
    onRunQuery?.();
  };
  const onResourcesChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, resources: e.target.value });
    onRunQuery?.();
  };
  const onClientsChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, clients: e.target.value });
    onRunQuery?.();
  };
  const onRegionChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, regions: e.target.value });
    onRunQuery?.();
  };
  const onCountryChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, countries: e.target.value });
    onRunQuery?.();
  };
  const onLegendFormatChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, legendFormat: e.target.value });
    onRunQuery?.();
  };

  const onZoneChange = (opt: SelectableValue<string>) => {
    onChange({ ...query, zone: opt?.value });
    onRunQuery?.();
  };
  const onDnsGranularityChange = (opt: SelectableValue<string>) => {
    onChange({ ...query, dnsGranularity: opt });
    onRunQuery?.();
  };
  const onRecordTypeChange = (opt: SelectableValue<string>) => {
    onChange({ ...query, record_type: opt?.value });
    onRunQuery?.();
  };
  const onDnsLegendFormatChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, dnsLegendFormat: e.target.value });
    onRunQuery?.();
  };

  const onAppChange = (opt: SelectableValue<number>) => {
    const name =
      opt?.value === 0
        ? "All"
        : fastedgeApps.find((a) => a.value === opt?.value)?.label;
    onChange({ ...query, appId: opt?.value ?? 0, appName: name });
    onRunQuery?.();
  };
  const onStepChange = (e: ChangeEvent<HTMLInputElement>) => {
    const step = parseInt(e.target.value, 10);
    onChange({ ...query, step: !Number.isNaN(step) && step > 0 ? step : 60 });
    onRunQuery?.();
  };
  const onNetworkChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, network: e.target.value });
    onRunQuery?.();
  };
  const onFastedgeMetricChange = (opt: SelectableValue<string>) => {
    onChange({ ...query, fastedgeMetric: opt?.value ?? "avg" });
    onRunQuery?.();
  };

  const onWaapMetricChange = (opt: SelectableValue<string>) => {
    onChange({ ...query, waapMetric: opt?.value ?? "total_requests" });
    onRunQuery?.();
  };
  const onWaapGranularityChange = (opt: SelectableValue<string>) => {
    onChange({ ...query, waapGranularity: (opt?.value as "1h" | "1d") ?? "1h" });
    onRunQuery?.();
  };
  const onWaapLegendFormatChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, waapLegendFormat: e.target.value });
  };

  const renderCdnSection = (q: GCQuery) => {
    const {
      metric,
      legendFormat,
      grouping,
      countries,
      granularity,
      regions,
      vhosts,
      resources,
      clients,
    } = defaults(q, defaultQuery);

    return (
      <>
        <div>
          <div className="section" style={{ marginRight: "27px" }}>
            <label className="gf-form-group-label">Query Props</label>
            <div className="gf-form">
              <FormField
                label="Metric"
                tooltip="Metric to query (traffic, cache, requests, responses, etc.)"
                inputEl={
                  <GCSelectMetric
                    width={20}
                    isSearchable
                    maxVisibleValues={20}
                    minMenuHeight={45}
                    menuPlacement="bottom"
                    onChange={onMetricChange}
                    value={metric}
                  />
                }
              />
            </div>
            <div className="gf-form">
              <FormField
                label="Granularity"
                tooltip="Time series granularity"
                inputEl={
                  <GCSelectGranularity
                    width={20}
                    maxVisibleValues={4}
                    minMenuHeight={25}
                    menuPlacement="bottom"
                    onChange={onGranularityChange}
                    value={granularity}
                  />
                }
              />
            </div>
            <div className="gf-form">
              <FormField
                label="Group By"
                tooltip="Fields used for grouping"
                inputEl={
                  <GCSelectGrouping
                    width={20}
                    isSearchable
                    maxVisibleValues={20}
                    minMenuHeight={35}
                    menuPlacement="bottom"
                    onChange={onGroupingChange}
                    value={grouping}
                  />
                }
              />
            </div>
          </div>
          <div className="section" style={{ marginRight: "27px" }}>
            <label className="gf-form-group-label">Filters (comma separated)</label>
            <div className="gf-form">
              <GCInput width={8} value={vhosts} onChange={onVhostsChange} label="Vhosts" tooltip="Filter by vhost" type="text" />
            </div>
            <div className="gf-form">
              <GCInput width={8} value={resources} onChange={onResourcesChange} label="Resources" tooltip="Filter by resource id" type="text" />
            </div>
            <div className="gf-form">
              <GCInput width={8} value={clients} onChange={onClientsChange} label="Clients" tooltip="Filter by client id" type="text" />
            </div>
            <div className="gf-form">
              <GCInput width={8} value={regions} onChange={onRegionChange} label="Regions" tooltip="Filter by region" type="text" />
            </div>
            <div className="gf-form">
              <GCInput width={8} value={countries} onChange={onCountryChange} label="Countries" tooltip="Filter by country" type="text" />
            </div>
          </div>
        </div>
        <div className="gf-form">
          <GCInput
            inputWidth={30}
            value={legendFormat}
            onChange={onLegendFormatChange}
            label="Legend"
            placeholder="legend format"
            tooltip="Controls the name of the time series."
            type="text"
          />
        </div>
      </>
    );
  };

  const renderDnsSection = (q: GCQuery) => {
    const dnsQuery = defaults(q, defaultDNSQuery) as GCQuery;
    const zone = dnsQuery.zone ?? "all";
    const selectedZone =
      zone === "all"
        ? { value: "all", label: "All Zones" }
        : dnsZones.find((z) => z.value === zone) ?? { value: zone, label: zone };
    const dnsGranularity = dnsQuery.dnsGranularity ?? {
      value: GCDNSGranularity.FiveMinutes,
      label: "5m",
    };
    const recordType = dnsQuery.record_type ?? GCDNSRecordType.All;

    return (
      <div style={{ display: "flex", flexDirection: "column", gap: "12px", maxWidth: "420px" }}>
        <label className="gf-form-group-label">Query Props</label>
        <FormField
          label="Zone"
          labelWidth={10}
          tooltip="Select a DNS Zone"
          inputEl={
            <Select
              width={20}
              options={dnsZones}
              isLoading={dnsZonesLoading}
              isSearchable
              menuPlacement="bottom"
              onChange={onZoneChange}
              value={selectedZone}
              placeholder={dnsZonesLoading ? "Loading zones..." : "Select zone"}
            />
          }
        />
        <FormField
          label="Granularity"
          labelWidth={10}
          tooltip="DNS statistics aggregation interval"
          inputEl={
            <Select
              width={20}
              options={DNS_GRANULARITY_OPTS}
              value={
                DNS_GRANULARITY_OPTS.find(
                  (o) => o.value === (dnsGranularity?.value ?? dnsGranularity)
                ) ?? DNS_GRANULARITY_OPTS[0]
              }
              onChange={onDnsGranularityChange}
            />
          }
        />
        <FormField
          label="Record type"
          labelWidth={10}
          tooltip="DNS record type to include in statistics"
          inputEl={
            <Select
              width={20}
              options={DNS_RECORD_TYPE_OPTS}
              value={
                DNS_RECORD_TYPE_OPTS.find((o) => o.value === recordType) ??
                DNS_RECORD_TYPE_OPTS[0]
              }
              onChange={onRecordTypeChange}
            />
          }
        />
        <GCInput
          inputWidth={30}
          value={dnsQuery.dnsLegendFormat ?? ""}
          onChange={onDnsLegendFormatChange}
          label="Legend"
          placeholder="e.g. zone, record_type"
          tooltip="Controls how DNS time series are named in the legend."
          type="text"
        />
      </div>
    );
  };

  const renderFastEdgeSection = (q: GCQuery) => {
    const fastedgeQuery = defaults(q, defaultFastEdgeQuery) as GCQuery;
    const appId = fastedgeQuery.appId ?? 0;
    const step = fastedgeQuery.step ?? 60;
    const selectedApp =
      fastedgeApps.find((a) => a.value === appId) ?? fastedgeApps[0];
    const metric = fastedgeQuery.fastedgeMetric ?? "avg";

    return (
      <div style={{ display: "flex", flexDirection: "column", gap: "16px", maxWidth: "420px" }}>
        <label className="gf-form-group-label">Query Props</label>
        <FormField
          label="App Name"
          labelWidth={12}
          tooltip="FastEdge application to query"
          inputEl={
            fastedgeAppsLoading ? (
              <Spinner />
            ) : (
              <Select
                options={fastedgeApps}
                value={selectedApp}
                onChange={onAppChange}
                placeholder="Select an App"
                width={30}
              />
            )
          }
        />
        <FormField
          label="Metric"
          labelWidth={12}
          tooltip="Latency statistic to display"
          inputEl={
            <Select
              options={FASTEDGE_METRIC_OPTS}
              value={
                FASTEDGE_METRIC_OPTS.find((o) => o.value === metric) ??
                FASTEDGE_METRIC_OPTS[0]
              }
              onChange={onFastedgeMetricChange}
              width={20}
            />
          }
        />
        <FormField
          label="Step (s)"
          labelWidth={12}
          tooltip="Sampling interval in seconds"
          inputEl={
            <input
              className="gf-form-input"
              type="number"
              value={String(step)}
              onChange={onStepChange}
              placeholder="Step in seconds"
            />
          }
        />
        <FormField
          label="Network"
          labelWidth={12}
          tooltip="Optional FastEdge network name"
          inputEl={
            <input
              className="gf-form-input"
              type="text"
              value={fastedgeQuery.network ?? ""}
              onChange={onNetworkChange}
              placeholder="Network name"
            />
          }
        />
      </div>
    );
  };

  const renderWaapSection = (q: GCQuery) => {
    const waapQuery = defaults(q, defaultWAAPQuery) as GCQuery;
    const metric = waapQuery.waapMetric ?? "total_requests";
    const granularity = waapQuery.waapGranularity ?? "1h";

    return (
      <div style={{ display: "flex", flexDirection: "column", gap: "16px", maxWidth: "520px" }}>
        <label className="gf-form-group-label">Query Props</label>
        <FormField
          label="Metric"
          labelWidth={10}
          tooltip="WAAP metric to query (for example total_requests or total_bytes)"
          inputEl={
            <Select
              options={WAAP_METRIC_OPTS}
              value={
                WAAP_METRIC_OPTS.find((o) => o.value === metric) ??
                WAAP_METRIC_OPTS[0]
              }
              onChange={onWaapMetricChange}
              width={30}
            />
          }
        />
        <FormField
          label="Granularity"
          labelWidth={10}
          tooltip="Aggregation interval for WAAP statistics"
          inputEl={
            <Select
              options={WAAP_GRANULARITY_OPTS}
              value={
                WAAP_GRANULARITY_OPTS.find((o) => o.value === granularity) ??
                WAAP_GRANULARITY_OPTS[0]
              }
              onChange={onWaapGranularityChange}
              width={30}
            />
          }
        />
        <FormField
          label="Legend Format"
          labelWidth={10}
          tooltip="Controls how WAAP time series are named in the legend."
          inputEl={
            <input
              className="gf-form-input"
              type="text"
              value={waapQuery.waapLegendFormat ?? ""}
              onChange={onWaapLegendFormatChange}
              placeholder="Legend format"
            />
          }
        />
      </div>
    );
  };

  const normalizedQuery = defaults(query, defaultQuery) as GCQuery;
  const productOption =
    GC_PRODUCTS.find((productValue) => productValue.value === product) ??
    GC_PRODUCTS[0];

  return (
    <div style={{ display: "flex", gap: "24px", alignItems: "flex-start" }}>
      <div className="gf-form-group" style={{ minWidth: "240px", maxWidth: "240px" }}>
        <label className="gf-form-group-label">Edge Network</label>
        <FormField
          label="Service"
          labelWidth={6}
          tooltip="Select Gcore service (CDN, DNS, FastEdge, WAAP)"
          inputEl={
            <Select<GCProduct>
              width={18}
              options={GC_PRODUCTS}
              value={productOption}
              onChange={onProductChange}
            />
          }
        />
      </div>

      <div style={{ flex: 1 }}>
        {product === "cdn" && renderCdnSection(normalizedQuery)}
        {product === "dns" && renderDnsSection(normalizedQuery)}
        {product === "fastedge" && renderFastEdgeSection(normalizedQuery)}
        {product === "waap" && renderWaapSection(normalizedQuery)}
      </div>
    </div>
  );
};
