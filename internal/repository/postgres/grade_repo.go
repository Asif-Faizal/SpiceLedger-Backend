package postgres

import (
    "context"

    "github.com/google/uuid"
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

func (r *gradeRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Grade, error) {
    var grade domain.Grade
    if err := r.db.WithContext(ctx).First(&grade, id).Error; err != nil {
        return nil, err
    }
    return &grade, nil
}
