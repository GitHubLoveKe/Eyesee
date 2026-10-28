package repository

import (
	"context"
	"dcs/internal/model"
	"time"
)

type BusinessSystemRepo interface {
	Create(ctx context.Context, bs *model.BusinessSystem) error
	GetByID(ctx context.Context, id string) (*model.BusinessSystem, error)
	List(ctx context.Context, keyword string, offset, limit int) ([]*model.BusinessSystem, int64, error)
	Update(ctx context.Context, bs *model.BusinessSystem) error
	Delete(ctx context.Context, id string) error

	AddMember(ctx context.Context, m *model.BusinessSystemMember) error
	RemoveMember(ctx context.Context, systemID, userID string) error
	ListMembers(ctx context.Context, systemID string) ([]*model.BusinessSystemMember, error)
	UpdateMemberRole(ctx context.Context, systemID, userID, role string) error

	ListByUserID(ctx context.Context, userID string) ([]*model.BusinessSystem, error)

	LinkZone(ctx context.Context, systemID, zoneID string) error
	UnlinkZone(ctx context.Context, systemID, zoneID string) error
}

type NetworkSegmentRepo interface {
	Create(ctx context.Context, seg *model.NetworkSegment) error
	GetByID(ctx context.Context, id int64) (*model.NetworkSegment, error)
	ListByZoneID(ctx context.Context, zoneID string) ([]*model.NetworkSegment, error)
	Update(ctx context.Context, seg *model.NetworkSegment) error
	Delete(ctx context.Context, id int64) error
	CheckCIDROverlap(ctx context.Context, zoneID, cidr string, excludeSegmentID *int64) (bool, error)
}

type ZoneRepo interface {
	Create(ctx context.Context, z *model.Zone) error
	GetByID(ctx context.Context, id string) (*model.Zone, error)
	List(ctx context.Context, status string, offset, limit int) ([]*model.Zone, int64, error)
	Update(ctx context.Context, z *model.Zone) error
	Delete(ctx context.Context, id string) error

	CreateSegment(ctx context.Context, s *model.NetworkSegment) error
	ListSegments(ctx context.Context, zoneID string) ([]*model.NetworkSegment, error)
	UpdateSegment(ctx context.Context, s *model.NetworkSegment) error
	DeleteSegment(ctx context.Context, segmentID int64) error
	CheckCIDROverlap(ctx context.Context, zoneID, cidr string, excludeSegmentID *int64) (bool, error)
}

type ResourceTypeRepo interface {
	Create(ctx context.Context, rt *model.ResourceType) error
	GetByID(ctx context.Context, id string) (*model.ResourceType, error)
	List(ctx context.Context, parentID *string) ([]*model.ResourceType, error)
	Update(ctx context.Context, rt *model.ResourceType) error
	Delete(ctx context.Context, id string) error
	HasChildren(ctx context.Context, id string) (bool, error)
	HasInstances(ctx context.Context, id string) (bool, error)
}

type InstanceRepo interface {
	Create(ctx context.Context, inst *model.Instance) error
	GetByID(ctx context.Context, id string) (*model.Instance, error)
	List(ctx context.Context, filter InstanceFilter) ([]*model.Instance, int64, error)
	Update(ctx context.Context, inst *model.Instance) error
	Delete(ctx context.Context, id string) error
	HasTasks(ctx context.Context, id string) (bool, error)
	UpdateTestResult(ctx context.Context, id string, testAt time.Time, result string) error
}

type InstanceFilter struct {
	ZoneID         string
	ResourceTypeID string
	BusinessSysID  string
	Status         string
	Keyword        string
	Offset         int
	Limit          int
}

type DataSinkRepo interface {
	Create(ctx context.Context, ds *model.DataSink) error
	GetByID(ctx context.Context, id string) (*model.DataSink, error)
	List(ctx context.Context, zoneID, sinkType string) ([]*model.DataSink, error)
	Update(ctx context.Context, ds *model.DataSink) error
	Delete(ctx context.Context, id string) error
	UpdateHealthResult(ctx context.Context, id string, checkAt time.Time, result string) error
}

type TaskTypeRepo interface {
	Create(ctx context.Context, tt *model.TaskType) error
	GetByID(ctx context.Context, id string) (*model.TaskType, error)
	List(ctx context.Context) ([]*model.TaskType, error)
	Update(ctx context.Context, tt *model.TaskType) error
	Delete(ctx context.Context, id string) error
}

type CollectionTaskRepo interface {
	Create(ctx context.Context, ct *model.CollectionTask) error
	GetByID(ctx context.Context, id string) (*model.CollectionTask, error)
	List(ctx context.Context, filter CollectionTaskFilter) ([]*model.CollectionTask, int64, error)
	Update(ctx context.Context, ct *model.CollectionTask) error
	Delete(ctx context.Context, id string) error
	CountByInstance(ctx context.Context, instanceID string) (int, error)
}

type CollectionTaskFilter struct {
	ZoneID     string
	InstanceID string
	TypeID     string
	NodeID     string
	Status     string
	Enabled    *bool
	Keyword    string
	Offset     int
	Limit      int
}

type CollectorNodeRepo interface {
	Create(ctx context.Context, n *model.CollectorNode) error
	GetByID(ctx context.Context, id string) (*model.CollectorNode, error)
	List(ctx context.Context, zoneID, nodeType, status string) ([]*model.CollectorNode, error)
	Update(ctx context.Context, n *model.CollectorNode) error
	Delete(ctx context.Context, id string) error
	UpdateHeartbeat(ctx context.Context, id string) error
}

type UserRepo interface {
	Create(ctx context.Context, u *model.SysUser) error
	GetByID(ctx context.Context, id string) (*model.SysUser, error)
	GetByUsername(ctx context.Context, username string) (*model.SysUser, error)
	List(ctx context.Context, filter UserFilter) ([]*model.SysUser, int64, error)
	Update(ctx context.Context, u *model.SysUser) error
	Delete(ctx context.Context, id string) error
	UpdateLastLogin(ctx context.Context, id string) error
}

type UserAuthSourceRepo interface {
	GetByUserIDAndType(ctx context.Context, userID string, sourceType string) (*model.UserAuthSource, error)
	Create(ctx context.Context, src *model.UserAuthSource) error
	UpdatePassword(ctx context.Context, userID string, sourceType string, passwordHash string) error
}

type UserFilter struct {
	Role    string
	Status  string
	Keyword string
	Offset  int
	Limit   int
}

type NotificationChannelRepo interface {
	Create(ctx context.Context, nc *model.NotificationChannel) error
	GetByID(ctx context.Context, id string) (*model.NotificationChannel, error)
	List(ctx context.Context, channelType string) ([]*model.NotificationChannel, error)
	Update(ctx context.Context, nc *model.NotificationChannel) error
	Delete(ctx context.Context, id string) error
	UpdateTestResult(ctx context.Context, id string, result string) error
}

type AlertGroupRepo interface {
	Create(ctx context.Context, g *model.AlertGroup) error
	GetByID(ctx context.Context, id string) (*model.AlertGroup, error)
	List(ctx context.Context) ([]*model.AlertGroup, error)
	Update(ctx context.Context, g *model.AlertGroup) error
	Delete(ctx context.Context, id string) error

	AddMember(ctx context.Context, m *model.AlertGroupMember) error
	RemoveMember(ctx context.Context, groupID, userID string) error
	ListMembers(ctx context.Context, groupID string) ([]*model.AlertGroupMember, error)
}

type AlertStrategyRepo interface {
	Create(ctx context.Context, s *model.AlertStrategy) error
	GetByID(ctx context.Context, id string) (*model.AlertStrategy, error)
	List(ctx context.Context, filter AlertStrategyFilter) ([]*model.AlertStrategy, int64, error)
	Update(ctx context.Context, s *model.AlertStrategy) error
	Delete(ctx context.Context, id string) error

	LinkInstance(ctx context.Context, si *model.AlertStrategyInstance) error
	UnlinkInstance(ctx context.Context, strategyID, instanceID string) error
	ListLinkedInstances(ctx context.Context, strategyID string) ([]*model.AlertStrategyInstance, error)
}

type AlertStrategyFilter struct {
	Severity string
	Status   string
	Enabled  *bool
	Keyword  string
	Offset   int
	Limit    int
}

type ScheduledTaskRepo interface {
	Create(ctx context.Context, t *model.ScheduledTask) error
	GetByID(ctx context.Context, id string) (*model.ScheduledTask, error)
	List(ctx context.Context, filter ScheduledTaskFilter) ([]*model.ScheduledTask, int64, error)
	Update(ctx context.Context, t *model.ScheduledTask) error
	Delete(ctx context.Context, id string) error

	CreateExecution(ctx context.Context, e *model.ScheduledTaskExecution) error
	UpdateExecution(ctx context.Context, e *model.ScheduledTaskExecution) error
	ListExecutions(ctx context.Context, taskID string, limit int) ([]*model.ScheduledTaskExecution, error)
}

type ScheduledTaskFilter struct {
	TaskType string
	Status   string
	Enabled  *bool
	Keyword  string
	Offset   int
	Limit    int
}

type RCTaskRepo interface {
	Create(ctx context.Context, t *model.RCTask) error
	GetByID(ctx context.Context, id string) (*model.RCTask, error)
	List(ctx context.Context, filter RCTaskFilter) ([]*model.RCTask, int64, error)
	Update(ctx context.Context, t *model.RCTask) error
	Delete(ctx context.Context, id string) error

	CreateRule(ctx context.Context, r *model.RCRule) error
	GetRuleByID(ctx context.Context, id string) (*model.RCRule, error)
	ListRules(ctx context.Context, taskID string) ([]*model.RCRule, error)
	UpdateRule(ctx context.Context, r *model.RCRule) error
	DeleteRule(ctx context.Context, id string) error
}

type RCTaskFilter struct {
	ZoneID  string
	Enabled *bool
	Keyword string
	Offset  int
	Limit   int
}

type MetricDefinitionRepo interface {
	Create(ctx context.Context, m *model.MetricDefinition) error
	GetByID(ctx context.Context, id string) (*model.MetricDefinition, error)
	List(ctx context.Context, filter MetricFilter) ([]*model.MetricDefinition, int64, error)
	Update(ctx context.Context, m *model.MetricDefinition) error
	Delete(ctx context.Context, id string) error

	CreateQuery(ctx context.Context, q *model.CommonQuery) error
	ListQueries(ctx context.Context, metricID string) ([]*model.CommonQuery, error)
	DeleteQuery(ctx context.Context, id string) error
}

type MetricFilter struct {
	Category   string
	MetricType string
	Keyword    string
	Offset     int
	Limit      int
}

type CredentialRepo interface {
	Create(ctx context.Context, c *model.Credential) error
	GetByID(ctx context.Context, id string) (*model.Credential, error)
	List(ctx context.Context, zoneID, credType string) ([]*model.Credential, error)
	Update(ctx context.Context, c *model.Credential) error
	Delete(ctx context.Context, id string) error
	HasInstances(ctx context.Context, id string) (bool, error)
}

type RCWorkerRepo interface {
	Create(ctx context.Context, w *model.RCWorker) error
	GetByID(ctx context.Context, id string) (*model.RCWorker, error)
	List(ctx context.Context, zoneID string) ([]*model.RCWorker, error)
	Update(ctx context.Context, w *model.RCWorker) error
	Delete(ctx context.Context, id string) error
	UpdateHeartbeat(ctx context.Context, id string) error
}

type AlertEventRepo interface {
	Create(ctx context.Context, e *model.AlertEvent) error
	GetByID(ctx context.Context, id int64) (*model.AlertEvent, error)
	List(ctx context.Context, filter AlertEventFilter) ([]*model.AlertEvent, int64, error)
	UpdateStatus(ctx context.Context, id int64, status string, endsAt *time.Time) error
	Statistics(ctx context.Context, zoneID string, period string) (*AlertStatistics, error)
}

type AlertEventFilter struct {
	ZoneID      string
	AlertName   string
	Severity    []string
	Status      []string
	InstanceID  string
	Fingerprint string
	StartTime   *time.Time
	EndTime     *time.Time
	Offset      int
	Limit       int
}

type AlertStatistics struct {
	TotalFiring   int64            `json:"total_firing"`
	TotalResolved int64            `json:"total_resolved"`
	BySeverity    map[string]int64 `json:"by_severity"`
	ByZone        []ZoneAlertCount `json:"by_zone"`
	TopAlerts     []AlertCount     `json:"top_alerts"`
}

type ZoneAlertCount struct {
	ZoneID   string `json:"zone_id"`
	Firing   int64  `json:"firing"`
	Resolved int64  `json:"resolved"`
}

type AlertCount struct {
	AlertName string `json:"alert_name"`
	Count     int64  `json:"count"`
}

type AlertAckRepo interface {
	Create(ctx context.Context, ack *model.AlertAcknowledgement) error
	GetByEventID(ctx context.Context, eventID int64) (*model.AlertAcknowledgement, error)
}

type AlertSilenceRepo interface {
	Create(ctx context.Context, s *model.AlertSilence) error
	GetByID(ctx context.Context, id string) (*model.AlertSilence, error)
	List(ctx context.Context, zoneID, status string) ([]*model.AlertSilence, error)
	Revoke(ctx context.Context, id string) error
}

type UserNotificationTargetRepo interface {
	Create(ctx context.Context, t *model.UserNotificationTarget) error
	ListByUserID(ctx context.Context, userID string) ([]*model.UserNotificationTarget, error)
	Delete(ctx context.Context, id int64) error
	SetPrimary(ctx context.Context, userID string, id int64) error
}

type ZoneCoordinatorRepo interface {
	Create(ctx context.Context, c *model.ZoneCoordinator) error
	GetByID(ctx context.Context, id string) (*model.ZoneCoordinator, error)
	List(ctx context.Context, zoneID string) ([]*model.ZoneCoordinator, error)
	Update(ctx context.Context, c *model.ZoneCoordinator) error
	Delete(ctx context.Context, id string) error
	UpdateHeartbeat(ctx context.Context, id string) error
}

type ServiceDiscoveryConfigRepo interface {
	Create(ctx context.Context, sd *model.ServiceDiscovery) error
	GetByID(ctx context.Context, id string) (*model.ServiceDiscovery, error)
	List(ctx context.Context, zoneID string) ([]*model.ServiceDiscovery, int64, error)
	Update(ctx context.Context, sd *model.ServiceDiscovery) error
	Delete(ctx context.Context, id string) error
}

type ResourceTypeMetricRepo interface {
	Link(ctx context.Context, typeID, metricID string) error
	Unlink(ctx context.Context, typeID, metricID string) error
	ListByType(ctx context.Context, typeID string) ([]*model.ResourceTypeMetric, error)
}

type BusinessOverviewRepo interface {
	GetOverview(ctx context.Context, systemID string) (*BusinessOverview, error)
}

type BusinessOverview struct {
	SystemID       string `json:"system_id"`
	InstanceCount  int64  `json:"instance_count"`
	TaskCount      int64  `json:"task_count"`
	AlertCount     int64  `json:"alert_count"`
	ZoneCount      int64  `json:"zone_count"`
	HealthyNodes   int64  `json:"healthy_nodes"`
	OfflineNodes   int64  `json:"offline_nodes"`
}
