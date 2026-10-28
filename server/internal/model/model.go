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
	SystemID       *string   `json:"system_id" db:"system_id"`
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

type ScheduledTaskExecution struct {
	ExecutionID  int64      `json:"execution_id" db:"execution_id"`
	TaskID       string     `json:"task_id" db:"task_id"`
	StartedAt    time.Time  `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
	DurationMs   *int       `json:"duration_ms" db:"duration_ms"`
	Status       string     `json:"status" db:"status"`
	ErrorMessage *string    `json:"error_message" db:"error_message"`
}

type NotificationChannel struct {
	ChannelID     string     `json:"channel_id" db:"channel_id"`
	Name          string     `json:"name" db:"name"`
	ChannelType   string     `json:"channel_type" db:"channel_type"`
	Config        string     `json:"config" db:"config"`
	Enabled       bool       `json:"enabled" db:"enabled"`
	LastTestAt    *time.Time `json:"last_test_at" db:"last_test_at"`
	LastTestResult *string   `json:"last_test_result" db:"last_test_result"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type AlertGroup struct {
	GroupID           string    `json:"group_id" db:"group_id"`
	Name              string    `json:"name" db:"name"`
	Description       string    `json:"description" db:"description"`
	ParentID          *string   `json:"parent_id" db:"parent_id"`
	EscalationMinutes *int      `json:"escalation_minutes" db:"escalation_minutes"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

type AlertGroupMember struct {
	ID        int64     `json:"id" db:"id"`
	GroupID   string    `json:"group_id" db:"group_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type AlertStrategy struct {
	StrategyID    string    `json:"strategy_id" db:"strategy_id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	MetricID      string    `json:"metric_id" db:"metric_id"`
	ExprTemplate  string    `json:"expr_template" db:"expr_template"`
	Duration      string    `json:"duration" db:"duration"`
	Severity      string    `json:"severity" db:"severity"`
	ScopeType     string    `json:"scope_type" db:"scope_type"`
	ScopeValue    string    `json:"scope_value" db:"scope_value"`
	AlertGroupID  *string   `json:"alert_group_id" db:"alert_group_id"`
	ChannelID     *string   `json:"channel_id" db:"channel_id"`
	Labels        string    `json:"labels" db:"labels"`
	Annotations   string    `json:"annotations" db:"annotations"`
	Enabled       bool      `json:"enabled" db:"enabled"`
	Status        string    `json:"status" db:"status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type AlertStrategyInstance struct {
	ID          int64     `json:"id" db:"id"`
	StrategyID  string    `json:"strategy_id" db:"strategy_id"`
	InstanceID  string    `json:"instance_id" db:"instance_id"`
	ZoneID      string    `json:"zone_id" db:"zone_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type RCTask struct {
	TaskID        string    `json:"task_id" db:"task_id"`
	ZoneID        string    `json:"zone_id" db:"zone_id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	StrategyID    *string   `json:"strategy_id" db:"strategy_id"`
	QueryAddress  *string   `json:"query_address" db:"query_address"`
	Enabled       bool      `json:"enabled" db:"enabled"`
	Status        string    `json:"status" db:"status"`
	RuleGroup     *string   `json:"rule_group" db:"rule_group"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type RCRule struct {
	RuleID       string    `json:"rule_id" db:"rule_id"`
	TaskID       string    `json:"task_id" db:"task_id"`
	Module       string    `json:"module" db:"module"`
	AlertName    string    `json:"alert_name" db:"alert_name"`
	Expr         string    `json:"expr" db:"expr"`
	Duration     string    `json:"duration" db:"duration"`
	Severity     string    `json:"severity" db:"severity"`
	Labels       string    `json:"labels" db:"labels"`
	Annotations  string    `json:"annotations" db:"annotations"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type RCWorker struct {
	WorkerID      string     `json:"worker_id" db:"worker_id"`
	ZoneID        string     `json:"zone_id" db:"zone_id"`
	Name          string     `json:"name" db:"name"`
	Address       string     `json:"address" db:"address"`
	QueryAddress  string     `json:"query_address" db:"query_address"`
	Capacity      int        `json:"capacity" db:"capacity"`
	OwnedRules    int        `json:"owned_rules" db:"owned_rules"`
	Status        string     `json:"status" db:"status"`
	LastHeartbeat *time.Time `json:"last_heartbeat" db:"last_heartbeat"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type MetricDefinition struct {
	MetricID       string    `json:"metric_id" db:"metric_id"`
	Name           string    `json:"name" db:"name"`
	Category       string    `json:"category" db:"category"`
	MetricType     *string   `json:"metric_type" db:"metric_type"`
	Help           *string   `json:"help" db:"help"`
	Unit           *string   `json:"unit" db:"unit"`
	PromqlTemplate *string   `json:"promql_template" db:"promql_template"`
	Variables      *string   `json:"variables" db:"variables"`
	Labels         *string   `json:"labels" db:"labels"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type CommonQuery struct {
	QueryID     string    `json:"query_id" db:"query_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	MetricID    string    `json:"metric_id" db:"metric_id"`
	Promql      string    `json:"promql" db:"promql"`
	Variables   *string   `json:"variables" db:"variables"`
	Category    *string   `json:"category" db:"category"`
	IsBuiltin   bool      `json:"is_builtin" db:"is_builtin"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Credential struct {
	CredentialID  string     `json:"credential_id" db:"credential_id"`
	ZoneID        string     `json:"zone_id" db:"zone_id"`
	Name          string     `json:"name" db:"name"`
	CredType      string     `json:"cred_type" db:"cred_type"`
	EncryptedData string     `json:"-" db:"encrypted_data"`
	Version       int        `json:"version" db:"version"`
	ExpiresAt     *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type AlertEvent struct {
	EventID     int64      `json:"event_id" db:"event_id"`
	AlertName   string     `json:"alert_name" db:"alert_name"`
	ZoneID      *string    `json:"zone_id" db:"zone_id"`
	InstanceID  *string    `json:"instance_id" db:"instance_id"`
	Severity    *string    `json:"severity" db:"severity"`
	Status      *string    `json:"status" db:"status"`
	Labels      string     `json:"labels" db:"labels"`
	Annotations string     `json:"annotations" db:"annotations"`
	StartsAt    *time.Time `json:"starts_at" db:"starts_at"`
	EndsAt      *time.Time `json:"ends_at" db:"ends_at"`
	Fingerprint *string    `json:"fingerprint" db:"fingerprint"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type AlertAcknowledgement struct {
	ID             int64     `json:"id" db:"id"`
	EventID        int64     `json:"event_id" db:"event_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	Note           string    `json:"note" db:"note"`
	AcknowledgedAt time.Time `json:"acknowledged_at" db:"acknowledged_at"`
}

type AlertSilence struct {
	SilenceID       string    `json:"silence_id" db:"silence_id"`
	ZoneID          *string   `json:"zone_id" db:"zone_id"`
	Matchers        string    `json:"matchers" db:"matchers"`
	StartsAt        time.Time `json:"starts_at" db:"starts_at"`
	EndsAt          time.Time `json:"ends_at" db:"ends_at"`
	Duration        *string   `json:"duration" db:"duration"`
	Reason          string    `json:"reason" db:"reason"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	Status          string    `json:"status" db:"status"`
	EngineSilenceID *string   `json:"engine_silence_id" db:"engine_silence_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type UserNotificationTarget struct {
	ID          int64     `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	ChannelType string    `json:"channel_type" db:"channel_type"`
	TargetValue string    `json:"target_value" db:"target_value"`
	IsPrimary   bool      `json:"is_primary" db:"is_primary"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type ZoneCoordinator struct {
	CoordinatorID string     `json:"coordinator_id" db:"coordinator_id"`
	ZoneID        string     `json:"zone_id" db:"zone_id"`
	Name          string     `json:"name" db:"name"`
	Address       string     `json:"address" db:"address"`
	GossipAddr    string     `json:"gossip_addr" db:"gossip_addr"`
	Role          string     `json:"role" db:"role"`
	ManagedNodes  int        `json:"managed_nodes" db:"managed_nodes"`
	Status        string     `json:"status" db:"status"`
	LastHeartbeat *time.Time `json:"last_heartbeat" db:"last_heartbeat"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type ServiceDiscovery struct {
	DiscoveryID     string     `json:"discovery_id" db:"discovery_id"`
	ZoneID          string     `json:"zone_id" db:"zone_id"`
	Name            string     `json:"name" db:"name"`
	Protocol        string     `json:"protocol" db:"protocol"`
	Config          string     `json:"config" db:"config"`
	RefreshInterval string     `json:"refresh_interval" db:"refresh_interval"`
	Enabled         bool       `json:"enabled" db:"enabled"`
	Status          string     `json:"status" db:"status"`
	LastRefreshAt   *time.Time `json:"last_refresh_at" db:"last_refresh_at"`
	DiscoveredCount int        `json:"discovered_count" db:"discovered_count"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type ResourceTypeMetric struct {
	ID        int64     `json:"id" db:"id"`
	TypeID    string    `json:"type_id" db:"type_id"`
	MetricID  string    `json:"metric_id" db:"metric_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UserAuthSource struct {
	ID           int64     `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	SourceType   string    `json:"source_type" db:"source_type"`
	ExternalID   string    `json:"external_id" db:"external_id"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Extra        string    `json:"extra" db:"extra"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
