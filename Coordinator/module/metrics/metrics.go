package metrics

import (
	"context"
	"fmt"
	"log/slog"

	"coordinator/config"
	"coordinator/model"
	"coordinator/module"

	"gopkg.in/yaml.v3"
)

func init() {
	module.Register(&MetricsModule{})
}

type MetricsModule struct {
	store  *model.Store
	config *config.Config
	logger *slog.Logger
}

func (m *MetricsModule) Name() string { return "metrics" }

func (m *MetricsModule) Init(ctx module.ModuleContext) error {
	m.store = ctx.Store
	m.config = ctx.Config
	m.logger = ctx.Logger
	return nil
}

func (m *MetricsModule) Start(_ context.Context) error { return nil }
func (m *MetricsModule) Stop(_ context.Context) error  { return nil }

func (m *MetricsModule) Routes() []module.TaskRoute {
	return []module.TaskRoute{
		{
			WorkerType: model.WorkerVMAgent,
			TaskFilter: func(t *model.CollectionTask) bool { return t.ModuleName() == "metrics" },
		},
		{
			WorkerType: model.WorkerVMAlert,
			RuleFilter: func(r *model.AlertRule) bool { return r.ModuleName() == "metrics" },
		},
	}
}

// ---------------------------------------------------------------------------
// ConfigGenerator — 为 vmagent 生成 scrape_config，为 vmalert 生成 rule_groups
// ---------------------------------------------------------------------------

func (m *MetricsModule) GenerateWorkerConfig(workerID string, assignment *model.Assignment) ([]byte, error) {
	w := m.store.GetWorker(workerID)
	if w == nil {
		return nil, fmt.Errorf("worker %s not found", workerID)
	}
	switch w.Type {
	case model.WorkerVMAgent:
		return m.generateScrapeConfig(assignment)
	case model.WorkerVMAlert:
		return m.generateAlertRules(assignment)
	default:
		return nil, fmt.Errorf("unsupported worker type: %s", w.Type)
	}
}

func (m *MetricsModule) ConfigContentType() string {
	return "application/x-yaml"
}

func (m *MetricsModule) generateScrapeConfig(assignment *model.Assignment) ([]byte, error) {
	tasks := m.store.GetEnabledTasks()

	taskSet := make(map[string]bool, len(assignment.TaskIDs))
	for _, id := range assignment.TaskIDs {
		taskSet[id] = true
	}

	var scrapeConfigs []map[string]interface{}
	for _, task := range tasks {
		if !taskSet[task.TaskID] {
			continue
		}
		sc := m.buildScrapeConfig(task)
		scrapeConfigs = append(scrapeConfigs, sc)
	}

	zoneResources := m.store.GetResourcesByType(model.ResourceVM)
	var remoteWriteURL string
	if len(zoneResources) > 0 {
		remoteWriteURL = zoneResources[0].Address + "/api/v1/write"
	}

	result := map[string]interface{}{
		"scrape_configs": scrapeConfigs,
	}
	if remoteWriteURL != "" {
		result["remote_write"] = []map[string]interface{}{
			{"url": remoteWriteURL},
		}
	}

	return yaml.Marshal(result)
}

func (m *MetricsModule) buildScrapeConfig(task *model.CollectionTask) map[string]interface{} {
	sc := map[string]interface{}{
		"job_name": task.JobName,
	}

	if task.ScrapeInterval != "" {
		sc["scrape_interval"] = task.ScrapeInterval
	}
	if task.MetricsPath != "" && task.MetricsPath != "/metrics" {
		sc["metrics_path"] = task.MetricsPath
	}

	labels := make(map[string]string)
	for k, v := range task.Labels {
		labels[k] = v
	}
	labels["zone"] = task.ZoneID
	labels["__meta_task_id"] = task.TaskID

	staticConfig := map[string]interface{}{
		"targets": []string{task.Target},
		"labels":  labels,
	}
	sc["static_configs"] = []interface{}{staticConfig}

	if task.Auth != nil && task.Auth.Username != "" {
		sc["basic_auth"] = map[string]interface{}{
			"username": task.Auth.Username,
			"password": task.Auth.Password,
		}
	}

	return sc
}

func (m *MetricsModule) generateAlertRules(assignment *model.Assignment) ([]byte, error) {
	allRules := m.store.GetAllRules()

	ruleSet := make(map[string]bool, len(assignment.RuleIDs))
	for _, id := range assignment.RuleIDs {
		ruleSet[id] = true
	}

	groupMap := make(map[string][]map[string]interface{})

	for _, rule := range allRules {
		if !ruleSet[rule.RuleID] {
			continue
		}

		groupName := rule.GroupName
		if groupName == "" {
			groupName = fmt.Sprintf("zone-%s-default", rule.ZoneID)
		}

		ruleDef := map[string]interface{}{
			"alert": rule.AlertName,
			"expr":  rule.Expr,
		}
		if rule.Duration != "" {
			ruleDef["for"] = rule.Duration
		}

		allLabels := make(map[string]string)
		allLabels["zone"] = rule.ZoneID
		if rule.Severity != "" {
			allLabels["severity"] = rule.Severity
		}
		for k, v := range rule.Labels {
			allLabels[k] = v
		}
		ruleDef["labels"] = allLabels

		if len(rule.Annotations) > 0 {
			ruleDef["annotations"] = rule.Annotations
		}

		groupMap[groupName] = append(groupMap[groupName], ruleDef)
	}

	var groups []map[string]interface{}
	for name, rules := range groupMap {
		group := map[string]interface{}{
			"name":  name,
			"rules": rules,
		}
		groups = append(groups, group)
	}

	result := map[string]interface{}{
		"groups": groups,
	}

	return yaml.Marshal(result)
}
