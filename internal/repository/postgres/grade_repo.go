package postgres

import (
	"context"

	"github.com/saravanan/spice_backend/internal/domain"
	"gorm.io/gorm"
)

type gradeRepository struct {
	db *gorm.DB
}

func NewGradeRepository(db *gorm.DB) domain.GradeRepository {
	return &gradeRepository{db: db}
}

func (r *gradeRepository) Create(ctx context.Context, grade *domain.Grade) error {
	return r.db.WithContext(ctx).Create(grade).Error
}

func (r *gradeRepository) FindAll(ctx context.Context) ([]domain.Grade, error) {
	var grades []domain.Grade
	if err := r.db.WithContext(ctx).Find(&grades).Error; err != nil {
		return nil, err
	}
	return grades, nil
}
