package repo

import (
	"context"
	"errors"
	"time"

	"drone-management/internal/domain"

	"gorm.io/gorm"
)

type DroneRepo struct {
	db *gorm.DB
}

func NewDroneRepo(db *gorm.DB) *DroneRepo {
	return &DroneRepo{db: db}
}

func (r *DroneRepo) EnsureForPrincipal(ctx context.Context, principalID uint) (*domain.Drone, error) {
	var row DroneRow
	err := r.db.WithContext(ctx).
		Where("principal_id = ?", principalID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = DroneRow{
			PrincipalID: principalID,
			Status:      domain.DroneStatusAvailable,
			UpdatedAt:   time.Now().UTC(),
		}
		if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		return droneToDomain(&row), nil
	}
	if err != nil {
		return nil, err
	}
	return droneToDomain(&row), nil
}

func (r *DroneRepo) ByID(ctx context.Context, id uint) (*domain.Drone, error) {
	return r.byIDDB(ctx, r.db, id)
}

func (r *DroneRepo) ByIDTx(ctx context.Context, tx *gorm.DB, id uint) (*domain.Drone, error) {
	return r.byIDDB(ctx, tx, id)
}

func (r *DroneRepo) byIDDB(ctx context.Context, db *gorm.DB, id uint) (*domain.Drone, error) {
	var row DroneRow
	if err := db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return droneToDomain(&row), nil
}

func (r *DroneRepo) ByPrincipalID(ctx context.Context, principalID uint) (*domain.Drone, error) {
	var row DroneRow
	if err := r.db.WithContext(ctx).Where("principal_id = ?", principalID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return droneToDomain(&row), nil
}

func (r *DroneRepo) List(ctx context.Context) ([]*domain.Drone, error) {
	var rows []DroneRow
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Drone, 0, len(rows))
	for i := range rows {
		out = append(out, droneToDomain(&rows[i]))
	}
	return out, nil
}

func (r *DroneRepo) UpdateHeartbeat(ctx context.Context, droneID uint, lat, lng float64, now time.Time) error {
	res := r.db.WithContext(ctx).Model(&DroneRow{}).
		Where("id = ?", droneID).
		Updates(map[string]any{
			"last_lat":          lat,
			"last_lng":          lng,
			"last_heartbeat_at": now,
			"updated_at":        now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// TransitionStatusTx flips a drone's status under optimistic lock. Caller may
// optionally clear current_order_id by setting clearOrder=true.
func (r *DroneRepo) TransitionStatusTx(ctx context.Context, tx *gorm.DB, droneID uint, expectedVersion uint, from, to domain.DroneStatus, clearOrder bool, setOrderID *uint, now time.Time) error {
	updates := map[string]any{
		"status":     to,
		"version":    gorm.Expr("version + 1"),
		"updated_at": now,
	}
	if clearOrder {
		updates["current_order_id"] = nil
	}
	if setOrderID != nil {
		updates["current_order_id"] = *setOrderID
	}
	res := tx.WithContext(ctx).Model(&DroneRow{}).
		Where("id = ? AND status = ? AND version = ?", droneID, from, expectedVersion).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrConflict
	}
	return nil
}
