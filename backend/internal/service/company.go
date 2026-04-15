package service

import (
	"context"
	"errors"

	"github.com/haojia/commute/internal/model"
	"github.com/haojia/commute/internal/repository"
)

const CompanySoftLimit = 20

type CompanyService struct {
	repo *repository.CompanyRepo
}

func NewCompanyService(repo *repository.CompanyRepo) *CompanyService {
	return &CompanyService{repo: repo}
}

func (s *CompanyService) List(ctx context.Context, userID int64, q model.CompanyListQuery) (model.CompanyListResult, error) {
	return s.repo.List(ctx, userID, q)
}

func (s *CompanyService) Get(ctx context.Context, userID, id int64) (*model.Company, error) {
	return s.repo.Get(ctx, userID, id)
}

func (s *CompanyService) Create(ctx context.Context, userID int64, in model.CompanyCreateInput) (*model.Company, error) {
	return s.repo.Create(ctx, userID, in)
}

func (s *CompanyService) Update(ctx context.Context, userID, id int64, in model.CompanyUpdateInput) (*model.Company, error) {
	return s.repo.Update(ctx, userID, id, in)
}

func (s *CompanyService) UpdateStatus(ctx context.Context, userID, id int64, status string) (*model.Company, error) {
	return s.repo.UpdateStatus(ctx, userID, id, status)
}

func (s *CompanyService) Delete(ctx context.Context, userID, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}

// Batch 批量创建公司。重复（同名同地址）跳过不报错。
func (s *CompanyService) Batch(ctx context.Context, userID int64, in model.CompanyBatchInput) (model.CompanyBatchResult, error) {
	result := model.CompanyBatchResult{
		Created: make([]model.Company, 0, len(in.Companies)),
		Skipped: make([]model.SkippedCompany, 0),
	}

	for _, item := range in.Companies {
		c, err := s.repo.Create(ctx, userID, item)
		if errors.Is(err, repository.ErrDuplicate) {
			result.Skipped = append(result.Skipped, model.SkippedCompany{
				Name: item.Name, Reason: "duplicate",
			})
			continue
		}
		if err != nil {
			result.Skipped = append(result.Skipped, model.SkippedCompany{
				Name: item.Name, Reason: err.Error(),
			})
			continue
		}
		result.Created = append(result.Created, *c)
	}

	// 软上限提示：当前用户公司总数 > 20 时给 warning
	listQ := model.CompanyListQuery{Page: 1, PageSize: 1}
	r, err := s.repo.List(ctx, userID, listQ)
	if err == nil && r.Pagination.Total > CompanySoftLimit {
		msg := "soft_limit_exceeded"
		result.Warning = &msg
	}
	return result, nil
}
