package repo

import (
	"context"
	"errors"
	"time"

	"drone-management/internal/domain"

	"gorm.io/gorm"
)

type PrincipalRepo struct {
	db *gorm.DB
}

func NewPrincipalRepo(db *gorm.DB) *PrincipalRepo {
	return &PrincipalRepo{db: db}
}

func (r *PrincipalRepo) Upsert(ctx context.Context, name string, role domain.Role) (*domain.Principal, error) {
	var row PrincipalRow
	err := r.db.WithContext(ctx).
		Where("name = ? AND role = ?", name, role).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = PrincipalRow{
			Name:      name,
			Role:      role,
			CreatedAt: time.Now().UTC(),
		}
		if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		return principalToDomain(&row), nil
	}
	if err != nil {
		return nil, err
	}
	return principalToDomain(&row), nil
}

func (r *PrincipalRepo) ByID(ctx context.Context, id uint) (*domain.Principal, error) {
	var row PrincipalRow
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return principalToDomain(&row), nil
}
