package service

import (
	"context"

	"drone-management/internal/database"
	"drone-management/internal/domain"
	"drone-management/internal/repo"
	"drone-management/internal/utils"

	"gorm.io/gorm"
)

type OrderService struct {
	db     *gorm.DB
	orders *repo.OrderRepo
	jobs   *repo.JobRepo
	events *repo.EventRepo
	clock  utils.Clock
}

func NewOrderService(db *gorm.DB, o *repo.OrderRepo, j *repo.JobRepo, e *repo.EventRepo, clock utils.Clock) *OrderService {
	return &OrderService{db: db, orders: o, jobs: j, events: e, clock: clock}
}

type SubmitOrderInput struct {
	EnduserPrincipalID   uint
	OriginLat, OriginLng float64
	DestLat, DestLng     float64
}

func (s *OrderService) Submit(ctx context.Context, in SubmitOrderInput) (*domain.Order, error) {
	if !validLatLng(in.OriginLat, in.OriginLng) || !validLatLng(in.DestLat, in.DestLng) {
		return nil, domain.ErrInvalidInput
	}
	now := s.clock.Now()
	var created *domain.Order
	err := database.WithTx(s.db, func(tx *gorm.DB) error {
		o, err := s.orders.Create(ctx, tx, repo.CreateOrderInput{
			EnduserPrincipalID: in.EnduserPrincipalID,
			OriginLat:          in.OriginLat,
			OriginLng:          in.OriginLng,
			DestLat:            in.DestLat,
			DestLng:            in.DestLng,
		}, now)
		if err != nil {
			return err
		}
		if _, err := s.jobs.CreateTx(ctx, tx, o.ID, domain.JobTypeOriginPickup, nil, now); err != nil {
			return err
		}
		actor := in.EnduserPrincipalID
		toStatus := o.Status
		if err := s.events.LogTx(ctx, tx, repo.LogInput{
			OrderID:          o.ID,
			EventType:        domain.EventOrderCreated,
			ToStatus:         &toStatus,
			ActorPrincipalID: &actor,
		}, now); err != nil {
			return err
		}
		created = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *OrderService) GetOwn(ctx context.Context, principalID, orderID uint) (*domain.Order, []*domain.OrderEvent, error) {
	o, err := s.orders.ByID(ctx, orderID)
	if err != nil {
		return nil, nil, err
	}
	if o.EnduserPrincipalID != principalID {
		return nil, nil, domain.ErrForbidden
	}
	events, err := s.events.ByOrder(ctx, o.ID)
	if err != nil {
		return nil, nil, err
	}
	return o, events, nil
}

func (s *OrderService) Withdraw(ctx context.Context, principalID, orderID uint) (*domain.Order, error) {
	now := s.clock.Now()
	var out *domain.Order
	err := database.WithTx(s.db, func(tx *gorm.DB) error {
		o, err := s.orders.ByIDTx(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if o.EnduserPrincipalID != principalID {
			return domain.ErrForbidden
		}
		if !domain.CanTransitionOrder(o.Status, domain.OrderStatusWithdrawn) {
			return domain.ErrInvalidTransition
		}
		from := o.Status
		if err := s.orders.TransitionStatusTx(ctx, tx, o.ID, o.Version, from, domain.OrderStatusWithdrawn, repo.OrderTransitionPatch{}, now); err != nil {
			return err
		}
		if err := s.jobs.CancelOpenForOrderTx(ctx, tx, o.ID, now); err != nil {
			return err
		}
		actor := principalID
		to := domain.OrderStatusWithdrawn
		if err := s.events.LogTx(ctx, tx, repo.LogInput{
			OrderID:          o.ID,
			EventType:        domain.EventOrderWithdrawn,
			FromStatus:       &from,
			ToStatus:         &to,
			ActorPrincipalID: &actor,
		}, now); err != nil {
			return err
		}
		updated, err := s.orders.ByIDTx(ctx, tx, o.ID)
		if err != nil {
			return err
		}
		out = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func validLatLng(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
