package engine

import "time"

// CommandPattern represents a pattern for command generation and analysis
type CommandPattern struct {
	Pattern            string            `json:"pattern"`
	Template           string            `json:"template"`
	Variables          []string          `json:"variables"`
	Frequency          int               `json:"frequency"`
	SuccessRate        float64           `json:"success_rate"`
	AverageTime        time.Duration     `json:"average_time"`
	Context            string            `json:"context"`
	Command            string            `json:"command"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}

// PerformanceMetrics represents performance measurement data
type PerformanceMetrics struct {
	ResponseTime    time.Duration `json:"response_time"`
	MemoryUsage     int64         `json:"memory_usage"`
	CPUUsage        float64       `json:"cpu_usage"`
	Throughput      float64       `json:"throughput"`
	ErrorRate       float64       `json:"error_rate"`
	Timestamp       time.Time     `json:"timestamp"`
	OperationType   string        `json:"operation_type"`
	ResourceCount   int           `json:"resource_count"`
	CacheHitRate    float64       `json:"cache_hit_rate"`
	NetworkLatency  time.Duration `json:"network_latency"`
}

// OptimizationRule represents a rule for optimizing operations
type OptimizationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ContextPattern represents patterns in context analysis
type ContextPattern struct {
	Pattern     string            `json:"pattern"`
	Context     string            `json:"context"`
	Frequency   int               `json:"frequency"`
	Confidence  float64           `json:"confidence"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    int64     `json:"memory_usage"`
	DiskUsage      int64     `json:"disk_usage"`
	NetworkIO      int64     `json:"network_io"`
	Timestamp      time.Time `json:"timestamp"`
	ProcessCount   int       `json:"process_count"`
	LoadAverage    float64   `json:"load_average"`
	Uptime         time.Duration `json:"uptime"`
}

// MetricsSnapshot represents a snapshot of metrics at a point in time
type MetricsSnapshot struct {
	Timestamp      time.Time         `json:"timestamp"`
	SystemMetrics  SystemMetrics     `json:"system_metrics"`
	AppMetrics     PerformanceMetrics `json:"app_metrics"`
	CustomMetrics  map[string]float64 `json:"custom_metrics"`
}

// Utility functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
