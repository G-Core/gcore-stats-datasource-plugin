# Gcore Grafana Data Source Plugin

This plugin connects Grafana to multiple Gcore edge services, providing a unified query model for visualizing both operational and security metrics across **CDN**, **DNS**, **FastEdge**, and **WAAP**, so instead of managing separate datasources or query editors you select the **Service** once and work with a single, consistent panel experience for all four products.

## Documentation

For a **step-by-step guide** (install, API key, panels per product, dashboard variables) in the style of [View CDN statistics in Grafana](https://gcore.com/docs/cdn/grafana/view-cdn-statistics-in-grafana), see **[docs/GcoreDocsMD.md](docs/GcoreDocsMD.md)**. Screenshots use images under `src/img/GcoreDocsImages/`.

## Overview
---

Grafana supports a wide range of data sources, including Prometheus, MySQL, and Datadog.
This plugin adds native support for Gcore APIs so you can build dashboards directly on top of Gcore traffic and security telemetry.

The data source currently supports the following Gcore products:

- **CDN** – delivery traffic, bandwidth, cache efficiency, and request statistics.
- **FastEdge** – application latency and performance metrics from edge applications.
- **WAAP** – web application and API protection statistics, including request volumes and traffic data.
- **DNS** – authoritative DNS traffic statistics per zone and record type.

## Product-specific capabilities
---

### CDN

#### What is the CDN data source?

The CDN service view connects Grafana to Gcore CDN statistics, focusing on delivery traffic, cache efficiency, and request/response trends.
It is suited for traffic analysis, capacity planning, and day‑to‑day operational monitoring of HTTP delivery.

#### Key features

- Visualize metrics such as total bytes, upstream bytes, cache hit ratios, and request/response counts.
- Filter by vhosts, resources, clients, regions, and countries using comma-separated filters.
- Group series by resource, client, region, country, vhost, or data center for multi-dimensional charts.
- Customize legend format to align series names with your dashboard conventions.

#### Find More

If you want to learn more about Gcore CDN and how its APIs work, see the official documentation at `https://gcore.com/docs/cdn` and the API reference at `https://gcore.com/docs/api-reference/overview`.

![CDN query editor](src/img/Service%20Images/CDNGraphImage.png)

---

### DNS

#### What is the DNS data source?

The DNS service view provides visibility into authoritative DNS traffic across zones and record types.
It helps you monitor query volume, understand load distribution, and validate DNS changes.

#### Key features

- Query statistics per zone or across all zones, with an "All Zones" option.
- Filter by DNS record type (A, AAAA, NS, CNAME, MX, TXT, SVCB, HTTPS).
- Control aggregation using DNS-specific time granularities from 5 minutes up to 24 hours.
- Customize legend format (for example using zone and record type) for clear time-series labels.

#### Find More

If you want to learn more about Gcore DNS and how its APIs work, see the official documentation at `https://gcore.com/docs/dns` and the Managed DNS product page at `https://gcore.com/dns`.

![DNS query editor](src/img/Service%20Images/DnsGraphImage.png)

---

### FastEdge

#### What is the FastEdge data source?

The FastEdge service view exposes application latency and performance metrics for Gcore edge applications.
It helps you understand response times, spot regressions, and compare deployments across networks or applications.

#### Key features

- Query application duration metrics (average, min, max, median, p75, p90) over time.
- Select a specific application or use an "All apps" view from the built‑in app selector.
- Control sampling step (granularity in seconds) to match the desired time-series resolution.
- Optionally filter by network name to focus on a particular FastEdge network.

#### Find More

If you want to learn more about Gcore FastEdge and how its APIs work, see the official documentation at `https://gcore.com/docs/fastedge` and the getting started guide at `https://gcore.com/docs/fastedge/getting-started`.



![FastEdge query editor](src/img/Service%20Images/FastEdgeGraphImage.png)

---

### WAAP

#### What is the WAAP data source?

The WAAP service view surfaces web application and API protection statistics from Gcore WAAP.
It is aimed at security and operations teams that need to track attack and traffic patterns over time.

#### Key features

- Visualize high-level WAAP metrics such as total requests and total bytes.
- Choose between hourly and daily granularity to support both operational and reporting use cases.
- Configure legend format so WAAP series integrate cleanly into existing dashboards.

#### Find More

If you want to learn more about Gcore WAAP and how its APIs work, see the official documentation at `https://gcore.com/docs/waap` and the API reference at `https://api.gcore.com/docs/waap`.

![WAAP query editor](src/img/Service%20Images/WaapGraphImage.png)


## How to start
---

### Run locally on Windows (Grafana installed on host)

#### Install prerequisites

- Local Grafana install (for example `C:\Program Files\GrafanaLabs\grafana`)
- Node.js (v14 or newer) and Yarn

#### Install dependencies

From the project root (`Cdnallplugins`):

```bash
yarn install
```

#### Build the plugin (frontend + backend)

```bash
yarn build:all
```

This bundles the frontend into `dist` and builds backend binaries for Linux and Windows.

#### Copy the plugin into Grafana

Copy the contents of the `dist` directory into Grafana’s plugins folder, for example:

```text
C:\Program Files\GrafanaLabs\grafana\data\plugins\gcore-stats-datasource-plugin
```

#### Allow the unsigned plugin

Edit `conf\custom.ini` in your Grafana install and ensure:

```ini
[plugins]
allow_loading_unsigned_plugins = gcore-stats-datasource-plugin
```

#### Restart Grafana and verify

Restart the Grafana service and open:

```text
http://localhost:3000
```

Go to **Connections → Data sources → Add data source** and search for **Gcore Platform-EN** (or the name from `plugin.json`).

---

### Run with Docker (Grafana + plugin)

#### First-time setup

From the project root (`Cdnallplugins`):

```bash
yarn install
yarn build:all
```

Make sure Docker and Docker Compose are installed.

#### Start Grafana with the plugin (Use Docker)

Recommended (uses the helper script):

```bash
yarn server
```

Or call Docker Compose directly:

```bash
docker compose -f .config/docker-compose-base.yaml up --build
```

#### Access Grafana

Open:

```text
http://localhost:3000
```

Log in (default `admin` / `admin` unless changed) and confirm the `gcore-stats-datasource-plugin` is available under **Connections → Data sources**.
