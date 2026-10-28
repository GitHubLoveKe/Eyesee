package traces

import (
	"context"

	"coordinator/model"
	"coordinator/module"
)

func init() {
	module.Register(&TracesModule{})
}

type TracesModule struct{}

func (m *TracesModule) Name() string                    { return "traces" }
func (m *TracesModule) Init(_ module.ModuleContext) error { return nil }
func (m *TracesModule) Start(_ context.Context) error    { return nil }
func (m *TracesModule) Stop(_ context.Context) error     { return nil }

func (m *TracesModule) Routes() []module.TaskRoute {
	return []module.TaskRoute{
		{
			WorkerType: model.WorkerOTelCollector,
			TaskFilter: func(t *model.CollectionTask) bool { return t.ModuleName() == "traces" },
		},
	}
}
