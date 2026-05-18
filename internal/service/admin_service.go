package service

import (
	"context"

	"drone-management/internal/database"
	"drone-management/internal/domain"
	"drone-management/internal/repo"
	"drone-management/internal/utils"

	"gorm.io/gorm"
)

type AdminService struct {
	db     *gorm.DB
	orders *repo.OrderRepo
	drones *repo.DroneRepo
	events *repo.EventRepo
	clock  utils.Clock
}

func NewAdminService(db *gorm.DB, o *repo.OrderRepo, d *repo.DroneRepo, e *repo.EventRepo, clock utils.Clock) *AdminService {
	return &AdminService{db: db, orders: o, drones: d, events: e, clock: clock}
}

func (s *AdminService) ListOrders(ctx context.Context, f repo.ListFilter) ([]*domain.Order, error) {
	return s.orders.List(ctx, f)
}

func (s *AdminService) ListDrones(ctx context.Context) ([]*domain.Drone, error) {
	return s.drones.List(ctx)
}

type PatchOrderInput struct {
	OriginLat *float64
	OriginLng *float64
	DestLat   *float64
	DestLng   *float64
}

func (s *AdminService) PatchOrder(ctx context.Context, orderID, actorID uint, in PatchOrderInput) (*domain.Order, error) {
	if in.OriginLat == nil && in.OriginLng == nil && in.DestLat == nil && in.DestLng == nil {
		return nil, domain.ErrInvalidInput
	}
	if (in.OriginLat == nil) != (in.OriginLng == nil) {
		return nil, domain.ErrInvalidInput
	}
	if (in.DestLat == nil) != (in.DestLng == nil) {
		return nil, domain.ErrInvalidInput
	}
	if in.OriginLat != nil && !validLatLng(*in.OriginLat, *in.OriginLng) {
		return nil, domain.ErrInvalidInput
	}
	if in.DestLat != nil && !validLatLng(*in.DestLat, *in.DestLng) {
		return nil, domain.ErrInvalidInput
	}

	now := s.clock.Now()
	var out *domain.Order
	err := database.WithTx(s.db, func(tx *gorm.DB) error {
		order, err := s.orders.ByIDTx(ctx, tx, orderID)
		if err != nil {
			return err
		}
		// disallow once goods are in flight (PICKED_UP) or terminal
		switch order.Status {
		case domain.OrderStatusPickedUp,
			domain.OrderStatusDelivered,
			domain.OrderStatusFailed,
			domain.OrderStatusWithdrawn:
			return domain.ErrInvalidTransition
		}
		if err := s.orders.UpdateOriginDestination(ctx, order.ID, in.OriginLat, in.OriginLng, in.DestLat, in.DestLng, order.Version, now); err != nil {
			return err
		}
		actor := actorID
		if err := s.events.LogTx(ctx, tx, repo.LogInput{
			OrderID:          order.ID,
			EventType:        domain.EventOrderUpdated,
			ActorPrincipalID: &actor,
		}, now); err != nil {
			return err
		}
		refreshed, err := s.orders.ByIDTx(ctx, tx, order.ID)
		if err != nil {
			return err
		}
		out = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
