package repo

import (
	"context"
	"errors"
	"time"

	"drone-management/internal/domain"

	"gorm.io/gorm"
)

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

type CreateOrderInput struct {
	EnduserPrincipalID   uint
	OriginLat, OriginLng float64
	DestLat, DestLng     float64
}

func (r *OrderRepo) Create(ctx context.Context, tx *gorm.DB, in CreateOrderInput, now time.Time) (*domain.Order, error) {
	row := OrderRow{
		EnduserPrincipalID: in.EnduserPrincipalID,
		Status:             domain.OrderStatusReadyForPickup,
		OriginLat:          in.OriginLat,
		OriginLng:          in.OriginLng,
		DestLat:            in.DestLat,
		DestLng:            in.DestLng,
		CreatedAt:          now,
		ReadyAt:            ptrTime(now),
		UpdatedAt:          now,
	}
	if err := tx.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return orderToDomain(&row), nil
}

func (r *OrderRepo) ByID(ctx context.Context, id uint) (*domain.Order, error) {
	return r.byIDDB(ctx, r.db, id)
}

func (r *OrderRepo) ByIDTx(ctx context.Context, tx *gorm.DB, id uint) (*domain.Order, error) {
	return r.byIDDB(ctx, tx, id)
}

func (r *OrderRepo) byIDDB(ctx context.Context, db *gorm.DB, id uint) (*domain.Order, error) {
	var row OrderRow
	if err := db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return orderToDomain(&row), nil
}

type ListFilter struct {
	IDs      []uint
	Statuses []domain.OrderStatus
	Limit    int
	Offset   int
}

func (r *OrderRepo) List(ctx context.Context, f ListFilter) ([]*domain.Order, error) {
	q := r.db.WithContext(ctx).Model(&OrderRow{}).Order("id ASC")
	if len(f.IDs) > 0 {
		q = q.Where("id IN ?", f.IDs)
	}
	if len(f.Statuses) > 0 {
		q = q.Where("status IN ?", f.Statuses)
	}
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}
	var rows []OrderRow
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Order, 0, len(rows))
	for i := range rows {
		out = append(out, orderToDomain(&rows[i]))
	}
	return out, nil
}

// UpdateOriginDestination sets origin and/or destination only when order is not yet PICKED_UP.
func (r *OrderRepo) UpdateOriginDestination(ctx context.Context, orderID uint, newOriginLat, newOriginLng, newDestLat, newDestLng *float64, expectedVersion uint, now time.Time) error {
	updates := map[string]any{
		"version":    gorm.Expr("version + 1"),
		"updated_at": now,
	}
	if newOriginLat != nil {
		updates["origin_lat"] = *newOriginLat
	}
	if newOriginLng != nil {
		updates["origin_lng"] = *newOriginLng
	}
	if newDestLat != nil {
		updates["dest_lat"] = *newDestLat
	}
	if newDestLng != nil {
		updates["dest_lng"] = *newDestLng
	}
	res := r.db.WithContext(ctx).Model(&OrderRow{}).
		Where("id = ? AND version = ? AND status IN ?", orderID, expectedVersion, []domain.OrderStatus{
			domain.OrderStatusCreated,
			domain.OrderStatusReadyForPickup,
			domain.OrderStatusReserved,
			domain.OrderStatusHandoffRequired,
		}).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrConflict
	}
	return nil
}

// TransitionStatusTx flips an order's status under optimistic lock and updates relevant timestamps.
// Optional fields let callers set assigned drone, current location, ETA.
type OrderTransitionPatch struct {
	AssignDroneID *uint
	ClearDrone    bool
	CurrentLat    *float64
	CurrentLng    *float64
	ETASeconds    *int64
	ClearETA      bool
}

func (r *OrderRepo) TransitionStatusTx(ctx context.Context, tx *gorm.DB, orderID uint, expectedVersion uint, from, to domain.OrderStatus, patch OrderTransitionPatch, now time.Time) error {
	updates := map[string]any{
		"status":     to,
		"version":    gorm.Expr("version + 1"),
		"updated_at": now,
	}
	switch to {
	case domain.OrderStatusReserved:
		updates["reserved_at"] = now
	case domain.OrderStatusPickedUp:
		updates["picked_up_at"] = now
	case domain.OrderStatusDelivered:
		updates["delivered_at"] = now
	case domain.OrderStatusFailed:
		updates["failed_at"] = now
	case domain.OrderStatusWithdrawn:
		updates["withdrawn_at"] = now
	case domain.OrderStatusHandoffRequired:
		updates["handoff_required_at"] = now
	case domain.OrderStatusReadyForPickup:
		updates["ready_at"] = now
	}
	if patch.AssignDroneID != nil {
		updates["assigned_drone_id"] = *patch.AssignDroneID
	}
	if patch.ClearDrone {
		updates["assigned_drone_id"] = nil
	}
	if patch.CurrentLat != nil {
		updates["current_lat"] = *patch.CurrentLat
	}
	if patch.CurrentLng != nil {
		updates["current_lng"] = *patch.CurrentLng
	}
	if patch.ETASeconds != nil {
		updates["eta_seconds"] = *patch.ETASeconds
	}
	if patch.ClearETA {
		updates["eta_seconds"] = nil
	}
	res := tx.WithContext(ctx).Model(&OrderRow{}).
		Where("id = ? AND status = ? AND version = ?", orderID, from, expectedVersion).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (r *OrderRepo) UpdateLocationTx(ctx context.Context, tx *gorm.DB, orderID uint, lat, lng float64, etaSec int64, now time.Time) error {
	res := tx.WithContext(ctx).Model(&OrderRow{}).
		Where("id = ?", orderID).
		Updates(map[string]any{
			"current_lat": lat,
			"current_lng": lng,
			"eta_seconds": etaSec,
			"updated_at":  now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func ptrTime(t time.Time) *time.Time { return &t }
