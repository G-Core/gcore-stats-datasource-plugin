# View Gcore statistics in Grafana

This guide walks through installing and using the **Gcore Platform** Grafana data source plugin (**Gcore Platform-EN**, id `gcore-stats-datasource-plugin`). It is the same kind of end-to-end documentation as [View CDN statistics in Grafana](https://gcore.com/docs/cdn/grafana/view-cdn-statistics-in-grafana) on the Gcore docs site, but updated for the unified plugin that covers **CDN**, **DNS**, **FastEdge**, and **WAAP** in one data source.

Screenshots in this file live under `src/img/GcoreDocsImages` in the repository.

## What you can visualize

### CDN

With the CDN service selected in the query editor, you can chart metrics such as:

- **Total Traffic** — total traffic volume (origin/shield/edge/user paths as exposed by the API).
- **Byte Cache Hit Ratio** — share of traffic served from cache (byte-oriented ratio).
- **Edges Traffic** — traffic from the edge toward clients.
- **Shield Traffic** — traffic involving [shielding](https://gcore.com/cdn/cdn-resource-options/general/enable-and-configure-origin-shielding).
- **Origin Traffic** — traffic from the origin toward CDN/shield.
- **Total Requests** — request counts to the CDN.
- **2xx / 3xx / 4xx / 5xx Responses** — HTTP response class counts.
- **Bandwidth** — derived from total traffic for rate-style views.
- **Cache Hit Ratio** — request-oriented cache hit ratio.
- **Shield Traffic Ratio** — shield efficiency as returned by the API.
- **Image optimization** — image processing volume where available.

You can **group** CDN series by **Client**, **Resource**, **Region**, **Country**, **Datacenter**, and **Vhost**.

For more CDN context, see [Gcore CDN documentation](https://gcore.com/docs/cdn).

### DNS, FastEdge, and WAAP

- **DNS** — zone-level statistics, record types, and DNS-specific granularities.
- **FastEdge** — application duration statistics (avg, min, max, median, percentiles) with app and optional network filters.
- **WAAP** — `total_requests` and `total_bytes` with hourly or daily buckets.

---

## Requirements

- **Grafana** version **12.0 or higher** (see `grafanaDependency` in `src/plugin.json`).
- A Gcore account with a **permanent API token** that can access the products you query.

---

## Step 1 — Download and install the plugin

1. Download the latest release asset for this plugin (for example a `dist.zip` or packaged archive) from  
   **[github.com/G-Core/gcore-stats-datasource-plugin/releases](https://github.com/G-Core/gcore-stats-datasource-plugin/releases)**.
2. Extract the plugin into your Grafana plugins directory (for example `grafana/data/plugins`; the exact path depends on your install).
3. If the plugin is not signed for your Grafana edition, allow it in Grafana configuration, for example:

   ```ini
   [plugins]
   allow_loading_unsigned_plugins = gcore-stats-datasource-plugin
   ```

4. [Restart Grafana](https://grafana.com/docs/grafana/latest/installation/restart-grafana/) and sign in.

For a developer-oriented build-from-source flow, see the root [README.md](../README.md).

---

## Step 2 — Add the data source and authenticate

1. Open **Connections → Data sources** (or **Configuration → Data sources** on older layouts).
2. Click **Add new data source** and search for **Gcore Platform-EN** (or the name shown in Grafana for this plugin).
3. Configure:
   - **URL** — API hostname, for example `api.gcore.com` (hostname only; no path). Leave default behavior or set a supported API host per your environment.
   - **API key** — your permanent API token. The plugin stores this securely; you may prefix with `APIKey ` if that is how you copy it from Gcore (the editor normalizes the value on save).

   ![Data source: URL and API key](../src/img/GcoreDocsImages/Login/Login.png)

4. Click **Save & test**. You should see a successful health check (for example authentication as your user).

Official reference for tokens: [Create, use, or delete a permanent API token](https://gcore.com/docs/account-settings/create-use-or-delete-a-permanent-api-token).

---

## Step 3 — Build a dashboard and run queries

1. Create a **New dashboard** and **Add visualization** (or add a panel).
2. Select your **Gcore** data source.
3. In the query editor, choose the **product** (**CDN**, **DNS**, **FastEdge**, or **WAAP**), then set metrics, granularity, filters, and legend options as described below.

### CDN panel setup

1. Set **Service** / product to **CDN**.

   ![Select CDN service](../src/img/GcoreDocsImages/CDN/CdnSelect.png)

2. Pick the **metric**, **granularity**, and **Group by** dimensions you need. Use **Filters (comma separated)** to narrow resources, vhosts, clients, regions, or countries as supported by the API.

   ![CDN metric and granularity](../src/img/GcoreDocsImages/CDN/CdnMetric.png)

   ![CDN granularity option](../src/img/GcoreDocsImages/CDN/CdnGranulaity.png)

   ![CDN group by](../src/img/GcoreDocsImages/CDN/CdnGroupby.png)

3. **Legend** — Grafana shows group-by fields and metric names by default. You can use a custom legend string with `{{fieldName}}` placeholders to match your dashboard style (for example `Traffic — {{resource}}`).

Example overview screenshot:

![CDN query overview](../src/img/GcoreDocsImages/CDN/CdnImage.png)

CDN reference on gcore.com (metrics concepts): [View CDN statistics in Grafana](https://gcore.com/docs/cdn/grafana/view-cdn-statistics-in-grafana).

### DNS panel setup

1. Select **DNS** as the product.

   ![Select DNS](../src/img/GcoreDocsImages/DNS/DNS.png)

   ![DNS query editor](../src/img/GcoreDocsImages/DNS/DnsSelect.png)

2. Choose **Zone** (including **All Zones** where applicable), **Record type**, and **Granularity**.

   ![DNS zone](../src/img/GcoreDocsImages/DNS/DnsZone.png)

   ![DNS record type](../src/img/GcoreDocsImages/DNS/DnsRecordType.png)

   ![DNS granularity](../src/img/GcoreDocsImages/DNS/DnsGranuality.png)

### FastEdge panel setup

1. Select **FastEdge**.

   ![Select FastEdge](../src/img/GcoreDocsImages/FastEdge/FastEdge.png)

   ![FastEdge service selection](../src/img/GcoreDocsImages/FastEdge/FastedgeSelect.png)

2. Choose duration **metric** (avg, min, max, median, percentiles), **App**, **Step** (sampling), and optional network filter if shown.

   ![FastEdge metric](../src/img/GcoreDocsImages/FastEdge/FastEdgeMetric.png)

   ![FastEdge app](../src/img/GcoreDocsImages/FastEdge/FastEdgeAppName.png)

   ![FastEdge step](../src/img/GcoreDocsImages/FastEdge/FastEdgeStep.png)

### WAAP panel setup

1. Select **WAAP**.

   ![Select WAAP](../src/img/GcoreDocsImages/Waap/WAAP.png)

   ![WAAP query](../src/img/GcoreDocsImages/Waap/WaapSelect.png)

2. Choose **Metric** (`total_requests` or `total_bytes`) and **Granularity** (`1h` or `1d`).

   ![WAAP metric](../src/img/GcoreDocsImages/Waap/WaapMetric.png)

   ![WAAP granularity](../src/img/GcoreDocsImages/Waap/WaapGranuality.png)

---

## Step 4 — Dashboard variables (optional)

You can drive filters from dashboard variables so users pick resources or other dimensions without editing the query.

1. Open dashboard **Settings → Variables → Add variable**.
2. Choose your Gcore data source and set **Values for** to the dimension you need, for example:
   - **resourceID** — CDN resources  
   - **vhost**, **client**, **country**, **region** — other CDN dimensions  
   - **Zone (DNS)**, **Record type (DNS)**  
   - **App (FastEdge)**

3. Save the variable, return to the dashboard, and reference it in the panel’s **Filters (comma separated)** (for CDN **Resources** or other filter fields as applicable). Use Grafana variable syntax, for example `$my_variable`, in the comma-separated list.

This mirrors the pattern described in the official CDN Grafana article (filter charts by resource via variables).

---

## Related links

| Topic | URL |
| -------- | ----- |
| CDN Grafana (official) | https://gcore.com/docs/cdn/grafana/view-cdn-statistics-in-grafana |
| API tokens | https://gcore.com/docs/account-settings/create-use-or-delete-a-permanent-api-token |
| Grafana restart / install | https://grafana.com/docs/grafana/latest/ |
| Plugin releases | https://github.com/G-Core/gcore-stats-datasource-plugin/releases |

---

## Image index (`src/img/GcoreDocsImages`)

| Folder | Files |
| ------ | ----- |
| `Login/` | `Login.png` |
| `CDN/` | `CdnSelect.png`, `CdnImage.png`, `CdnMetric.png`, `CdnGranulaity.png`, `CdnGroupby.png` |
| `DNS/` | `DNS.png`, `DnsSelect.png`, `DnsZone.png`, `DnsRecordType.png`, `DnsGranuality.png` |
| `FastEdge/` | `FastEdge.png`, `FastedgeSelect.png`, `FastEdgeMetric.png`, `FastEdgeAppName.png`, `FastEdgeStep.png` |
| `Waap/` | `WAAP.png`, `WaapSelect.png`, `WaapMetric.png`, `WaapGranuality.png` |

Paths in this markdown file are relative to the `docs/` directory (for example `../src/img/GcoreDocsImages/...`).
