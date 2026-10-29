package service

import (
	"context"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"fmt"
)

type ServiceDiscoveryService struct {
	taskRepo repository.CollectionTaskRepo
	instRepo repository.InstanceRepo
}

func NewServiceDiscoveryService(taskRepo repository.CollectionTaskRepo, instRepo repository.InstanceRepo) *ServiceDiscoveryService {
	return &ServiceDiscoveryService{taskRepo: taskRepo, instRepo: instRepo}
}

type SDTarget struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func (s *ServiceDiscoveryService) GetTargets(ctx context.Context, zoneID, nodeID string) ([]SDTarget, error) {
	filter := repository.CollectionTaskFilter{
		ZoneID: zoneID,
		NodeID: nodeID,
		Limit:  1000,
	}

	tasks, _, err := s.taskRepo.List(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to list collection tasks")
	}

	var targets []SDTarget
	for _, t := range tasks {
		if !t.Enabled {
			continue
		}

		inst, err := s.instRepo.GetByID(ctx, t.InstanceID)
		if err != nil || inst == nil {
			continue
		}

		port := ""
		if inst.Port != nil {
			port = fmt.Sprintf(":%d", *inst.Port)
		}

		target := SDTarget{
			Targets: []string{fmt.Sprintf("%s%s", inst.Host, port)},
			Labels: map[string]string{
				"job":             fmt.Sprintf("%s-%s", inst.ResourceTypeID, inst.InstanceID),
				"instance":        inst.InstanceID,
				"resource_type":   inst.ResourceTypeID,
				"__meta_task_id":  t.TaskID,
				"__meta_zone_id":  zoneID,
				"__meta_node_id":  nodeID,
			},
		}
		targets = append(targets, target)
	}

	return targets, nil
}
