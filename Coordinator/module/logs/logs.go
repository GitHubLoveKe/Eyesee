package logs

import (
	"context"

	"coordinator/model"
	"coordinator/module"
)

func init() {
	module.Register(&LogsModule{})
}

type LogsModule struct{}

func (m *LogsModule) Name() string                    { return "logs" }
func (m *LogsModule) Init(_ module.ModuleContext) error { return nil }
func (m *LogsModule) Start(_ context.Context) error    { return nil }
func (m *LogsModule) Stop(_ context.Context) error     { return nil }

func (m *LogsModule) Routes() []module.TaskRoute {
	return []module.TaskRoute{
		{
			WorkerType: model.WorkerOTelCollector,
			TaskFilter: func(t *model.CollectionTask) bool { return t.ModuleName() == "logs" },
		},
	}
}
