package repo

import (
	"context"
	"errors"
	"time"

	"drone-management/internal/domain"

	"gorm.io/gorm"
)

type JobRepo struct {
	db *gorm.DB
}

func NewJobRepo(db *gorm.DB) *JobRepo {
	return &JobRepo{db: db}
}

func (r *JobRepo) CreateTx(ctx context.Context, tx *gorm.DB, orderID uint, jobType domain.JobType, sourceDroneID *uint, now time.Time) (*domain.Job, error) {
	row := JobRow{
		OrderID:       orderID,
		Type:          jobType,
		Status:        domain.JobStatusOpen,
		SourceDroneID: sourceDroneID,
		CreatedAt:     now,
	}
	if err := tx.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return jobToDomain(&row), nil
}

func (r *JobRepo) ByID(ctx context.Context, id uint) (*domain.Job, error) {
	return r.byIDDB(ctx, r.db, id)
}

func (r *JobRepo) ByIDTx(ctx context.Context, tx *gorm.DB, id uint) (*domain.Job, error) {
	return r.byIDDB(ctx, tx, id)
}

func (r *JobRepo) byIDDB(ctx context.Context, db *gorm.DB, id uint) (*domain.Job, error) {
	var row JobRow
	if err := db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return jobToDomain(&row), nil
}

func (r *JobRepo) ListOpen(ctx context.Context, jobType *domain.JobType) ([]*domain.Job, error) {
	q := r.db.WithContext(ctx).Model(&JobRow{}).
		Where("status = ?", domain.JobStatusOpen).
		Order("created_at ASC")
	if jobType != nil {
		q = q.Where("type = ?", *jobType)
	}
	var rows []JobRow
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Job, 0, len(rows))
	for i := range rows {
		out = append(out, jobToDomain(&rows[i]))
	}
	return out, nil
}

func (r *JobRepo) ByOrderTx(ctx context.Context, tx *gorm.DB, orderID uint) ([]*domain.Job, error) {
	var rows []JobRow
	if err := tx.WithContext(ctx).Where("order_id = ?", orderID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Job, 0, len(rows))
	for i := range rows {
		out = append(out, jobToDomain(&rows[i]))
	}
	return out, nil
}

func (r *JobRepo) ReserveTx(ctx context.Context, tx *gorm.DB, jobID uint, expectedVersion uint, droneID uint, now time.Time) error {
	res := tx.WithContext(ctx).Model(&JobRow{}).
		Where("id = ? AND status = ? AND version = ?", jobID, domain.JobStatusOpen, expectedVersion).
		Updates(map[string]any{
			"status":               domain.JobStatusReserved,
			"reserved_by_drone_id": droneID,
			"reserved_at":          now,
			"version":              gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrAlreadyReserved
	}
	return nil
}

func (r *JobRepo) CompleteTx(ctx context.Context, tx *gorm.DB, jobID uint, expectedVersion uint, now time.Time) error {
	res := tx.WithContext(ctx).Model(&JobRow{}).
		Where("id = ? AND status = ? AND version = ?", jobID, domain.JobStatusReserved, expectedVersion).
		Updates(map[string]any{
			"status":       domain.JobStatusCompleted,
			"completed_at": now,
			"version":      gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (r *JobRepo) CancelOpenForOrderTx(ctx context.Context, tx *gorm.DB, orderID uint, now time.Time) error {
	res := tx.WithContext(ctx).Model(&JobRow{}).
		Where("order_id = ? AND status = ?", orderID, domain.JobStatusOpen).
		Updates(map[string]any{
			"status":       domain.JobStatusCancelled,
			"completed_at": now,
			"version":      gorm.Expr("version + 1"),
		})
	return res.Error
}
