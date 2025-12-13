package service

import (
	"context"
	"time"

	"github.com/saravanan/spice_backend/internal/domain"
)

type GradeService struct {
	gradeRepo domain.GradeRepository
}

func NewGradeService(gradeRepo domain.GradeRepository) *GradeService {
	return &GradeService{gradeRepo: gradeRepo}
}

func (s *GradeService) CreateGrade(ctx context.Context, name, description string) error {
	grade := &domain.Grade{
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}
	return s.gradeRepo.Create(ctx, grade)
}

func (s *GradeService) ListGrades(ctx context.Context) ([]domain.Grade, error) {
	return s.gradeRepo.FindAll(ctx)
}
