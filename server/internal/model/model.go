package model

import "time"

type BusinessSystem struct {
	SystemID    string    `json:"system_id" db:"system_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type BusinessSystemMember struct {
	ID        int64     `json:"id" db:"id"`
	SystemID  string    `json:"system_id" db:"system_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Zone struct {
	ZoneID         string    `json:"zone_id" db:"zone_id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	CollectorType  string    `json:"collector_type" db:"collector_type"`
	RemoteWriteURL string    `json:"remote_write_url" db:"remote_write_url"`
	DefaultSinkID  *string   `json:"default_sink_id" db:"default_sink_id"`
	Config         string    `json:"config" db:"config"`
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type NetworkSegment struct {
	SegmentID   int64     `json:"segment_id" db:"segment_id"`
	ZoneID      string    `json:"zone_id" db:"zone_id"`
	CIDR        string    `json:"cidr" db:"cidr"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type ResourceType struct {
	TypeID      string    `json:"type_id" db:"type_id"`
	ParentID    *string   `json:"parent_id" db:"parent_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Icon        string    `json:"icon" db:"icon"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	Attributes  string    `json:"attributes" db:"attributes"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Instance struct {
	InstanceID     string     `json:"instance_id" db:"instance_id"`
	ZoneID         string     `json:"zone_id" db:"zone_id"`
	ResourceTypeID string     `json:"resource_type_id" db:"resource_type_id"`
	BusinessSysID  *string    `json:"business_system_id" db:"business_system_id"`
	Name           string     `json:"name" db:"name"`
	Host           string     `json:"host" db:"host"`
	Port           *int       `json:"port" db:"port"`
	CredentialID   *string    `json:"credential_id" db:"credential_id"`
	Labels         string     `json:"labels" db:"labels"`
	Extra          string     `json:"extra" db:"extra"`
	Status         string     `json:"status" db:"status"`
	LastTestAt     *time.Time `json:"last_test_at" db:"last_test_at"`
	LastTestResult *string    `json:"last_test_result" db:"last_test_result"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type DataSink struct {
	SinkID      string     `json:"sink_id" db:"sink_id"`
	ZoneID      string     `json:"zone_id" db:"zone_id"`
	Name        string     `json:"name" db:"name"`
	SinkType    string     `json:"sink_type" db:"sink_type"`
	Protocol    string     `json:"protocol" db:"protocol"`
	Config      string     `json:"config" db:"config"`
	IsDefault   bool       `json:"is_default" db:"is_default"`
	Status      string     `json:"status" db:"status"`
	LastCheckAt *time.Time `json:"last_check_at" db:"last_check_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type TaskType struct {
	TypeID          string    `json:"type_id" db:"type_id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	SourceType      string    `json:"source_type" db:"source_type"`
	Collector       string    `json:"collector" db:"collector"`
	FormSchema      string    `json:"form_schema" db:"form_schema"`
	DetailPanels    string    `json:"detail_panels" db:"detail_panels"`
	DeclaredMetrics string    `json:"declared_metrics" db:"declared_metrics"`
	ConfigTemplate  string    `json:"config_template" db:"config_template"`
	IsBuiltin       bool      `json:"is_builtin" db:"is_builtin"`
	Version         int       `json:"version" db:"version"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CollectionTask struct {
	TaskID     string     `json:"task_id" db:"task_id"`
	InstanceID string     `json:"instance_id" db:"instance_id"`
	TypeID     string     `json:"type_id" db:"type_id"`
	Module     string     `json:"module" db:"module"`
	NodeID     *string    `json:"node_id" db:"node_id"`
	DataSinkID *string    `json:"data_sink_id" db:"data_sink_id"`
	Name       string     `json:"name" db:"name"`
	Config     string     `json:"config" db:"config"`
	Enabled    bool       `json:"enabled" db:"enabled"`
	Status     string     `json:"status" db:"status"`
	LastScrape *time.Time `json:"last_scrape_at" db:"last_scrape_at"`
	LastError  *string    `json:"last_error" db:"last_error"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

type CollectorNode struct {
	NodeID         string     `json:"node_id" db:"node_id"`
	ZoneID         string     `json:"zone_id" db:"zone_id"`
	Name           string     `json:"name" db:"name"`
	NodeType       string     `json:"node_type" db:"node_type"`
	Address        string     `json:"address" db:"address"`
	HealthEndpoint string     `json:"health_endpoint" db:"health_endpoint"`
	Labels         string     `json:"labels" db:"labels"`
	Status         string     `json:"status" db:"status"`
	TaskCount      int        `json:"task_count" db:"task_count"`
	LastHeartbeat  *time.Time `json:"last_heartbeat" db:"last_heartbeat"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type SysUser struct {
	UserID      string     `json:"user_id" db:"user_id"`
	Username    string     `json:"username" db:"username"`
	DisplayName string     `json:"display_name" db:"display_name"`
	Email       string     `json:"email" db:"email"`
	Phone       string     `json:"phone" db:"phone"`
	Role        string     `json:"role" db:"role"`
	Status      string     `json:"status" db:"status"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type ScheduledTask struct {
	TaskID        string     `json:"task_id" db:"task_id"`
	BusinessSysID *string    `json:"business_system_id" db:"business_system_id"`
	Name          string     `json:"name" db:"name"`
	TaskType      string     `json:"task_type" db:"task_type"`
	CronExpr      string     `json:"cron_expr" db:"cron_expr"`
	Config        string     `json:"config" db:"config"`
	Enabled       bool       `json:"enabled" db:"enabled"`
	LastRunAt     *time.Time `json:"last_run_at" db:"last_run_at"`
	NextRunAt     *time.Time `json:"next_run_at" db:"next_run_at"`
	Status        string     `json:"status" db:"status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}
