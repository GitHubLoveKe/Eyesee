package service

import (
	"context"
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/pkg/errors"
	"time"

	"github.com/google/uuid"
)

type MetricDefinitionService struct {
	repo repository.MetricDefinitionRepo
}

func NewMetricDefinitionService(repo repository.MetricDefinitionRepo) *MetricDefinitionService {
	return &MetricDefinitionService{repo: repo}
}

func (s *MetricDefinitionService) Create(ctx context.Context, req CreateMetricRequest) (*model.MetricDefinition, error) {
	if req.Name == "" {
		return nil, errors.Wrap(nil, 40001, 400, "name is required")
	}

	category := req.Category
	if category == "" {
		category = "metric"
	}

	m := &model.MetricDefinition{
		MetricID:       "metric-" + uuid.New().String()[:8],
		Name:           req.Name,
		Category:       category,
		MetricType:     req.MetricType,
		Help:           req.Help,
		Unit:           req.Unit,
		PromqlTemplate: req.PromqlTemplate,
		Variables:      req.Variables,
		Labels:         req.Labels,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create metric definition")
	}
	return m, nil
}

type CreateMetricRequest struct {
	Name           string
	Category       string
	MetricType     *string
	Help           *string
	Unit           *string
	PromqlTemplate *string
	Variables      *string
	Labels         *string
}

func (s *MetricDefinitionService) GetByID(ctx context.Context, id string) (*model.MetricDefinition, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get metric definition")
	}
	if m == nil {
		return nil, errors.ErrNotFound
	}
	return m, nil
}

func (s *MetricDefinitionService) List(ctx context.Context, category, metricType, keyword string, page, pageSize int) ([]*model.MetricDefinition, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	filter := repository.MetricFilter{
		Category:   category,
		MetricType: metricType,
		Keyword:    keyword,
		Offset:     (page - 1) * pageSize,
		Limit:      pageSize,
	}
	return s.repo.List(ctx, filter)
}

func (s *MetricDefinitionService) Update(ctx context.Context, id string, req UpdateMetricRequest) (*model.MetricDefinition, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get metric definition")
	}
	if m == nil {
		return nil, errors.ErrNotFound
	}

	if req.Name != "" {
		m.Name = req.Name
	}
	if req.Category != "" {
		m.Category = req.Category
	}
	m.MetricType = req.MetricType
	m.Help = req.Help
	m.Unit = req.Unit
	m.PromqlTemplate = req.PromqlTemplate
	m.Variables = req.Variables
	m.Labels = req.Labels
	m.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, m); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to update metric definition")
	}
	return m, nil
}

type UpdateMetricRequest struct {
	Name           string
	Category       string
	MetricType     *string
	Help           *string
	Unit           *string
	PromqlTemplate *string
	Variables      *string
	Labels         *string
}

func (s *MetricDefinitionService) Delete(ctx context.Context, id string) error {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, 50001, 500, "failed to get metric definition")
	}
	if m == nil {
		return errors.ErrNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *MetricDefinitionService) CreateQuery(ctx context.Context, metricID string, req CreateQueryRequest) (*model.CommonQuery, error) {
	m, err := s.repo.GetByID(ctx, metricID)
	if err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to get metric definition")
	}
	if m == nil {
		return nil, errors.ErrNotFound
	}

	q := &model.CommonQuery{
		QueryID:     "cq-" + uuid.New().String()[:8],
		Name:        req.Name,
		Description: req.Description,
		MetricID:    metricID,
		Promql:      req.Promql,
		Variables:   req.Variables,
		Category:    req.Category,
		IsBuiltin:   false,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateQuery(ctx, q); err != nil {
		return nil, errors.Wrap(err, 50001, 500, "failed to create common query")
	}
	return q, nil
}

type CreateQueryRequest struct {
	Name        string
	Description *string
	Promql      string
	Variables   *string
	Category    *string
}

func (s *MetricDefinitionService) ListQueries(ctx context.Context, metricID string) ([]*model.CommonQuery, error) {
	return s.repo.ListQueries(ctx, metricID)
}

func (s *MetricDefinitionService) DeleteQuery(ctx context.Context, queryID string) error {
	return s.repo.DeleteQuery(ctx, queryID)
}
