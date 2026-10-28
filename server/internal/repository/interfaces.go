package repository

import (
	"context"
	"dcs/internal/model"
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

	LinkZone(ctx context.Context, systemID, zoneID string) error
	UnlinkZone(ctx context.Context, systemID, zoneID string) error
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
