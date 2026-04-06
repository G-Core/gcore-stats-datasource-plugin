package core

import "encoding/json"

type SelectableValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// QueryModel is the unified query model; Product determines which fields are used.
type QueryModel struct {
	Product string `json:"product"`

	// CDN Edge
	Metric      SelectableValue   `json:"metric"`
	Granularity SelectableValue   `json:"granularity"`
	Grouping    []SelectableValue `json:"grouping,omitempty"`
	Vhosts      string            `json:"vhosts,omitempty"`
	Resources   string            `json:"resources,omitempty"`
	Countries   string            `json:"countries,omitempty"`
	Regions     string            `json:"regions,omitempty"`
	Clients     string            `json:"clients,omitempty"`
	Legend      string            `json:"legendFormat,omitempty"`

	// DNS Edge
	Zone           string          `json:"zone,omitempty"`
	RecordType     string          `json:"record_type,omitempty"`
	DnsGranularity SelectableValue `json:"dnsGranularity,omitempty"`
	DnsLegendFormat string        `json:"dnsLegendFormat,omitempty"`

	// FastEdge Edge
	AppId          int64  `json:"appId,omitempty"`
	AppName        string `json:"appName,omitempty"`
	Step           int64  `json:"step,omitempty"`
	Network        string `json:"network,omitempty"`
	FastedgeMetric string `json:"fastedgeMetric,omitempty"`

	// WAAP Edge
	WaapMetric       string `json:"waapMetric,omitempty"`
	WaapGranularity  string `json:"waapGranularity,omitempty"`
	WaapLegendFormat string `json:"waapLegendFormat,omitempty"`
}

// CDN Edge
// CDN types
type StatsRequest struct {
	Metrics     []string `json:"metrics"`
	Granularity string   `json:"granularity,omitempty"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	Vhosts      []string `json:"vhosts,omitempty"`
	Resources   []int64  `json:"resources,omitempty"`
	Countries   []string `json:"countries,omitempty"`
	Regions     []string `json:"regions,omitempty"`
	Clients     []int64  `json:"clients,omitempty"`
	GroupBy     []string `json:"group_by,omitempty"`
	Flat        bool     `json:"flat"`
}

type StatsResponse struct {
	Metrics  map[string]json.RawMessage `json:"metrics"`
	Client   *int                       `json:"client,omitempty"`
	Region   string                     `json:"region,omitempty"`
	Vhost    string                     `json:"vhost,omitempty"`
	Country  string                     `json:"country,omitempty"`
	DC       string                     `json:"dc,omitempty"`
	Resource *int                       `json:"resource,omitempty"`
}


// DNS types
type Zone struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type ZonesResponse struct {
	Zones       []Zone `json:"zones"`
	TotalAmount int    `json:"total_amount"`
}

type DNSStatsResponse struct {
	Requests map[string]float64 `json:"requests"`
	Total    float64           `json:"total"`
}

// FastEdge Edge
// FastEdge types
type GCAppDuration struct {
	Time    string  `json:"time"`
	Avg     float64 `json:"avg"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Median  float64 `json:"median"`
	Perc75  float64 `json:"perc75"`
	Perc90  float64 `json:"perc90"`
	Network string  `json:"network,omitempty"`
}

type FastEdgeResponseStats struct {
	Stats []GCAppDuration `json:"stats"`
}

// WAAP Edge
// WAAP types
type WaapDataPoint struct {
	DateTime string  `json:"date_time"`
	Value    float64 `json:"value"`
}

type WaapStatsResponse struct {
	Data  map[string][]WaapDataPoint `json:"data"`
	Total float64                   `json:"total"`
}
