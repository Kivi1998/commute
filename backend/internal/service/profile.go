package service

import (
	"context"

	"github.com/haojia/commute/internal/model"
	"github.com/haojia/commute/internal/repository"
)

type ProfileService struct {
	repo *repository.ProfileRepo
}

func NewProfileService(repo *repository.ProfileRepo) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) Get(ctx context.Context, userID int64) (*model.Profile, error) {
	return s.repo.Get(ctx, userID)
}

func (s *ProfileService) Upsert(ctx context.Context, userID int64, in model.ProfileUpsertInput) (*model.Profile, error) {
	return s.repo.Upsert(ctx, userID, in)
}
