package model

// SystemInfo holds core operating system and host details.
type SystemInfo struct {
	Hostname       string `json:"hostname"`
	OS             string `json:"os"`
	Distribution   string `json:"distribution,omitempty"`
	DistroVersion  string `json:"distro_version,omitempty"`
	Kernel         string `json:"kernel"`
	Arch           string `json:"arch"`
	UptimeSeconds  uint64 `json:"uptime_seconds"`
	Uptime         string `json:"uptime"`
	CurrentUser    string `json:"current_user"`
	Shell          string `json:"shell"`
	Desktop        string `json:"desktop,omitempty"`
	WindowManager  string `json:"window_manager,omitempty"`
	InitSystem     string `json:"init_system,omitempty"`
	PackageManager string `json:"package_manager,omitempty"`
}

// CPUInfo holds detailed CPU hardware and performance metrics.
type CPUInfo struct {
	Model         string    `json:"model"`
	Vendor        string    `json:"vendor,omitempty"`
	PhysicalCores int       `json:"physical_cores"`
	LogicalCores  int       `json:"logical_cores"`
	FrequencyMHz  float64   `json:"frequency_mhz,omitempty"`
	UsagePercent  float64   `json:"usage_percent"`
	LoadAverages  [3]float64 `json:"load_averages"` // 1m, 5m, 15m
	TemperatureC  float64   `json:"temperature_c,omitempty"`
}

// MemoryInfo holds RAM and Swap metrics in bytes.
type MemoryInfo struct {
	TotalBytes     uint64  `json:"total_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	FreeBytes      uint64  `json:"free_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	CachedBytes    uint64  `json:"cached_bytes,omitempty"`
	BuffersBytes   uint64  `json:"buffers_bytes,omitempty"`
	UsagePercent   float64 `json:"usage_percent"`

	SwapTotalBytes uint64  `json:"swap_total_bytes,omitempty"`
	SwapUsedBytes  uint64  `json:"swap_used_bytes,omitempty"`
	SwapFreeBytes  uint64  `json:"swap_free_bytes,omitempty"`
	SwapPercent    float64 `json:"swap_percent,omitempty"`
}

// DiskPartition holds storage usage metrics for a mounted volume.
type DiskPartition struct {
	MountPoint   string  `json:"mount_point"`
	Device       string  `json:"device,omitempty"`
	FSType       string  `json:"fs_type,omitempty"`
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	FreeBytes    uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

// GPUInfo holds graphics accelerator metrics.
type GPUInfo struct {
	Name               string  `json:"name"`
	Vendor             string  `json:"vendor,omitempty"`
	Driver             string  `json:"driver,omitempty"`
	VRAMTotalBytes     uint64  `json:"vram_total_bytes,omitempty"`
	VRAMUsedBytes      uint64  `json:"vram_used_bytes,omitempty"`
	UtilizationPercent float64 `json:"utilization_percent,omitempty"`
	TemperatureC       float64 `json:"temperature_c,omitempty"`
}

// NetworkInterface holds metrics for a single network interface.
type NetworkInterface struct {
	Name         string   `json:"name"`
	State        string   `json:"state"` // "up", "down", "unknown"
	IPv4         []string `json:"ipv4,omitempty"`
	IPv6         []string `json:"ipv6,omitempty"`
	MAC          string   `json:"mac,omitempty"`
	SpeedMbps    int64    `json:"speed_mbps,omitempty"`
	RXBytes      uint64   `json:"rx_bytes"`
	TXBytes      uint64   `json:"tx_bytes"`
	RXPackets    uint64   `json:"rx_packets,omitempty"`
	TXPackets    uint64   `json:"tx_packets,omitempty"`
	RXErrors     uint64   `json:"rx_errors,omitempty"`
	TXErrors     uint64   `json:"tx_errors,omitempty"`
}

// ServiceItem holds status of an individual system service.
type ServiceItem struct {
	Name        string `json:"name"`
	Status      string `json:"status"` // "running", "failed", "stopped", "inactive", "enabled"
	Active      bool   `json:"active"`
	Enabled     bool   `json:"enabled,omitempty"`
	Description string `json:"description,omitempty"`
}

// ServiceSummary holds overall service manager status.
type ServiceSummary struct {
	Manager string        `json:"manager"` // "systemd", "openrc", "launchd", "unknown"
	Running int           `json:"running"`
	Failed  int           `json:"failed"`
	Total   int           `json:"total"`
	Items   []ServiceItem `json:"items,omitempty"`
}

// ProcessInfo holds information about a running process.
type ProcessInfo struct {
	PID          int     `json:"pid"`
	User         string  `json:"user"`
	CPUPercent   float64 `json:"cpu_percent"`
	MemoryBytes  uint64  `json:"memory_bytes"`
	MemoryPercent float64 `json:"memory_percent"`
	Command      string  `json:"command"`
	Status       string  `json:"status,omitempty"`
}

// Snapshot holds a complete unified system diagnostic snapshot.
type Snapshot struct {
	Timestamp int64              `json:"timestamp"`
	System    SystemInfo         `json:"system"`
	CPU       CPUInfo            `json:"cpu"`
	Memory    MemoryInfo         `json:"memory"`
	Disks     []DiskPartition    `json:"disks"`
	GPUs      []GPUInfo          `json:"gpus,omitempty"`
	Network   []NetworkInterface `json:"network"`
	Services  ServiceSummary     `json:"services"`
	Processes []ProcessInfo      `json:"processes,omitempty"`
}
