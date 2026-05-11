package service

import (
	"passworder/internal/model"
	"passworder/internal/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(c *model.Category) error {
	return s.repo.Create(c)
}

func (s *CategoryService) Update(c *model.Category) error {
	return s.repo.Update(c)
}

func (s *CategoryService) Delete(id int64) error {
	return s.repo.Delete(id)
}

func (s *CategoryService) Get(id int64) (*model.Category, error) {
	return s.repo.Get(id)
}

func (s *CategoryService) List() ([]model.Category, error) {
	return s.repo.List()
}
