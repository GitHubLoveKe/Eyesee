package network

import (
	"context"

	"coordinator/model"
	"coordinator/module"
)

func init() {
	module.Register(&NetworkModule{})
}

type NetworkModule struct{}

func (m *NetworkModule) Name() string                    { return "network" }
func (m *NetworkModule) Init(_ module.ModuleContext) error { return nil }
func (m *NetworkModule) Start(_ context.Context) error    { return nil }
func (m *NetworkModule) Stop(_ context.Context) error     { return nil }

func (m *NetworkModule) Routes() []module.TaskRoute {
	return []module.TaskRoute{
		{
			WorkerType: model.WorkerExporter,
			TaskFilter: func(t *model.CollectionTask) bool { return t.ModuleName() == "network" },
		},
	}
}
