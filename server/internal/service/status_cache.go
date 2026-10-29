package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type StatusCacheService struct {
	rdb *redis.Client
}

func NewStatusCacheService(rdb *redis.Client) *StatusCacheService {
	return &StatusCacheService{rdb: rdb}
}

type ZoneOverview struct {
	ZoneID        string `json:"zone_id"`
	InstanceCount int64  `json:"instance_count"`
	TaskRunning   int64  `json:"task_running"`
	TaskError     int64  `json:"task_error"`
	ActiveAlerts  int64  `json:"active_alerts"`
	HealthyNodes  int64  `json:"healthy_nodes"`
	LastUpdated   string `json:"last_updated"`
}

type GlobalSummary struct {
	TotalZones     int64 `json:"total_zones"`
	TotalInstances int64 `json:"total_instances"`
	TotalTasks     int64 `json:"total_tasks"`
	ActiveAlerts   int64 `json:"active_alerts"`
	HealthyNodes   int64 `json:"healthy_nodes"`
	OfflineNodes   int64 `json:"offline_nodes"`
}

func (s *StatusCacheService) SetGlobalSummary(ctx context.Context, summary *GlobalSummary) error {
	data, _ := json.Marshal(summary)
	return s.rdb.Set(ctx, "global:summary", data, 5*time.Minute).Err()
}

func (s *StatusCacheService) GetGlobalSummary(ctx context.Context) (*GlobalSummary, error) {
	data, err := s.rdb.Get(ctx, "global:summary").Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var summary GlobalSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, err
	}
	return &summary, nil
}

func (s *StatusCacheService) SetZoneOverview(ctx context.Context, overview *ZoneOverview) error {
	data, _ := json.Marshal(overview)
	key := fmt.Sprintf("zone:%s:overview", overview.ZoneID)
	return s.rdb.Set(ctx, key, data, 5*time.Minute).Err()
}

func (s *StatusCacheService) GetZoneOverview(ctx context.Context, zoneID string) (*ZoneOverview, error) {
	key := fmt.Sprintf("zone:%s:overview", zoneID)
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var overview ZoneOverview
	if err := json.Unmarshal(data, &overview); err != nil {
		return nil, err
	}
	return &overview, nil
}

func (s *StatusCacheService) RefreshZoneOverview(ctx context.Context, zoneID string, instanceCount, taskRunning, taskError, activeAlerts, healthyNodes int64) error {
	overview := &ZoneOverview{
		ZoneID:        zoneID,
		InstanceCount: instanceCount,
		TaskRunning:   taskRunning,
		TaskError:     taskError,
		ActiveAlerts:  activeAlerts,
		HealthyNodes:  healthyNodes,
		LastUpdated:   time.Now().UTC().Format(time.RFC3339),
	}
	return s.SetZoneOverview(ctx, overview)
}
