package service

import (
	"context"

	"github.com/haojia/commute/internal/model"
	"github.com/haojia/commute/internal/repository"
)

type AddressService struct {
	repo *repository.AddressRepo
}

func NewAddressService(repo *repository.AddressRepo) *AddressService {
	return &AddressService{repo: repo}
}

func (s *AddressService) List(ctx context.Context, userID int64) ([]model.HomeAddress, error) {
	return s.repo.List(ctx, userID)
}

func (s *AddressService) Get(ctx context.Context, userID, id int64) (*model.HomeAddress, error) {
	return s.repo.Get(ctx, userID, id)
}

func (s *AddressService) Create(ctx context.Context, userID int64, in model.HomeAddressCreateInput) (*model.HomeAddress, error) {
	return s.repo.Create(ctx, userID, in)
}

func (s *AddressService) Update(ctx context.Context, userID, id int64, in model.HomeAddressUpdateInput) (*model.HomeAddress, error) {
	return s.repo.Update(ctx, userID, id, in)
}

func (s *AddressService) SetDefault(ctx context.Context, userID, id int64) (*model.HomeAddress, error) {
	return s.repo.SetDefault(ctx, userID, id)
}

func (s *AddressService) Delete(ctx context.Context, userID, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}
