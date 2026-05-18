package service

import (
	"context"
	"errors"

	"drone-management/internal/database"
	"drone-management/internal/domain"
	"drone-management/internal/repo"
	"drone-management/internal/utils"

	"gorm.io/gorm"
)

type DroneService struct {
	db         *gorm.DB
	drones     *repo.DroneRepo
	orders     *repo.OrderRepo
	jobs       *repo.JobRepo
	events     *repo.EventRepo
	principals *repo.PrincipalRepo
	clock      utils.Clock
	avgSpeedMS float64
}

func NewDroneService(db *gorm.DB, d *repo.DroneRepo, o *repo.OrderRepo, j *repo.JobRepo, e *repo.EventRepo, p *repo.PrincipalRepo, clock utils.Clock, avgSpeedMS float64) *DroneService {
	return &DroneService{db: db, drones: d, orders: o, jobs: j, events: e, principals: p, clock: clock, avgSpeedMS: avgSpeedMS}
}

func (s *DroneService) Self(ctx context.Context, principalID uint) (*domain.Drone, error) {
	return s.drones.ByPrincipalID(ctx, principalID)
}

func (s *DroneService) ListOpenJobs(ctx context.Context, jobType *domain.JobType) ([]*domain.Job, error) {
	return s.jobs.ListOpen(ctx, jobType)
}

// ReserveJob: drone reserves an OPEN job. Order transitions
// READY_FOR_PICKUP→RESERVED (origin) or HANDOFF_REQUIRED→RESERVED (handoff).
// Drone transitions AVAILABLE→BUSY.
func (s *DroneService) ReserveJob(ctx context.Context, principalID, jobID uint) (*domain.Job, *domain.Order, error) {
	now := s.clock.Now()
	var resJob *domain.Job
	var resOrder *domain.Order
	err := database.WithTx(s.db, func(tx *gorm.DB) error {
		drone, err := s.drones.ByPrincipalID(ctx, principalID)
		if err != nil {
			return err
		}
		// re-read drone under tx to get fresh version
		drone, err = s.drones.ByIDTx(ctx, tx, drone.ID)
		if err != nil {
			return err
		}
		if drone.Status == domain.DroneStatusBroken {
			return domain.ErrDroneBroken
		}
		if drone.Status != domain.DroneStatusAvailable {
			return domain.ErrDroneBusy
		}
		job, err := s.jobs.ByIDTx(ctx, tx, jobID)
		if err != nil {
			return err
		}
		if job.Status != domain.JobStatusOpen {
			return domain.ErrAlreadyReserved
		}
		order, err := s.orders.ByIDTx(ctx, tx, job.OrderID)
		if err != nil {
			return err
		}

		var orderFrom domain.OrderStatus
		switch job.Type {
		case domain.JobTypeOriginPickup:
			if order.Status != domain.OrderStatusReadyForPickup {
				return domain.ErrInvalidTransition
			}
			orderFrom = domain.OrderStatusReadyForPickup
		case domain.JobTypeHandoffPickup:
			if order.Status != domain.OrderStatusHandoffRequired {
				return domain.ErrInvalidTransition
			}
			orderFrom = domain.OrderStatusHandoffRequired
		default:
			return domain.ErrInvalidInput
		}

		if err := s.jobs.ReserveTx(ctx, tx, job.ID, job.Version, drone.ID, now); err != nil {
			return err
		}
		droneID := drone.ID
		if err := s.orders.TransitionStatusTx(ctx, tx, order.ID, order.Version, orderFrom, domain.OrderStatusReserved, repo.OrderTransitionPatch{
			AssignDroneID: &droneID,
		}, now); err != nil {
			return err
		}
		if err := s.drones.TransitionStatusTx(ctx, tx, drone.ID, drone.Version, domain.DroneStatusAvailable, domain.DroneStatusBusy, false, nil, now); err != nil {
			return err
		}
		actor := principalID
		from := orderFrom
		to := domain.OrderStatusReserved
		if err := s.events.LogTx(ctx, tx, repo.LogInput{
			OrderID:          order.ID,
			EventType:        domain.EventOrderReserved,
			FromStatus:       &from,
			ToStatus:         &to,
			ActorPrincipalID: &actor,
		}, now); err != nil {
			return err
		}
		// Refresh
		updatedJob, err := s.jobs.ByIDTx(ctx, tx, job.ID)
		if err != nil {
			return err
		}
		updatedOrder, err := s.orders.ByIDTx(ctx, tx, order.ID)
		if err != nil {
			return err
		}
		resJob = updatedJob
		resOrder = updatedOrder
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return resJob, resOrder, nil
}

// Pickup transitions the drone's currently reserved order RESERVED→PICKED_UP
// and completes the reservation job.
func (s *DroneService) Pickup(ctx context.Context, principalID, orderID uint) (*domain.Order, error) {
	now := s.clock.Now()
	var out *domain.Order
	err := database.WithTx(s.db, func(tx *gorm.DB) error {
		drone, err := s.drones.ByPrincipalID(ctx, principalID)
		if err != nil {
			return err
		}
		drone, err = s.drones.ByIDTx(ctx, tx, drone.ID)
		if err != nil {
			return err
		}
		order, err := s.orders.ByIDTx(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if order.AssignedDroneID == nil || *order.AssignedDroneID != drone.ID {
			return domain.ErrForbidden
		}
		if order.Status != domain.OrderStatusReserved {
			return domain.ErrInvalidTransition
		}
		// find the RESERVED job for this drone+order
		jobs, err := s.jobs.ByOrderTx(ctx, tx, order.ID)
		if err != nil {
			return err
		}
		var reservedJob *domain.Job
		for _, j := range jobs {
			if j.Status == domain.JobStatusReserved && j.ReservedByDroneID != nil && *j.ReservedByDroneID == drone.ID {
				reservedJob = j
				break
			}
		}
		if reservedJob == nil {
			return domain.ErrConflict
		}

		from := order.Status
		patch := repo.OrderTransitionPatch{
			CurrentLat: &order.OriginLat,
			CurrentLng: &order.OriginLng,
		}
		if drone.LastLat != nil && drone.LastLng != nil {
			lat, lng := *drone.LastLat, *drone.LastLng
			patch.CurrentLat = &lat
			patch.CurrentLng = &lng
		}
		etaSec := utils.ETASeconds(*patch.CurrentLat, *patch.CurrentLng, order.DestLat, order.DestLng, s.avgSpeedMS)
		patch.ETASeconds = &etaSec

		if err := s.orders.TransitionStatusTx(ctx, tx, order.ID, order.Version, from, domain.OrderStatusPickedUp, patch, now); err != nil {
			return err
		}
		if err := s.jobs.CompleteTx(ctx, tx, reservedJob.ID, reservedJob.Version, now); err != nil {
			return err
		}
		oid := order.ID
		if err := s.drones.TransitionStatusTx(ctx, tx, drone.ID, drone.Version, drone.Status, domain.DroneStatusBusy, false, &oid, now); err != nil {
			// Drone may already be BUSY (its status was set by Reserve). We only need to update current_order_id.
			if !errors.Is(err, domain.ErrConflict) {
				return err
			}
			// Update only current_order_id without status-check
			res := tx.WithContext(ctx).Model(&repo.DroneRow{}).
				Where("id = ?", drone.ID).
				Updates(map[string]any{
					"current_order_id": oid,
					"updated_at":       now,
				})
			if res.Error != nil {
				return res.Error
			}
		}
		actor := principalID
		to := domain.OrderStatusPickedUp
		if err := s.events.LogTx(ctx, tx, repo.LogInput{
			OrderID:          order.ID,
			EventType:        domain.EventOrderPickedUp,
			FromStatus:       &from,
			ToStatus:         &to,
			ActorPrincipalID: &actor,
			Lat:              patch.CurrentLat,
			Lng:              patch.CurrentLng,
		}, now); err != nil {
			return err
		}
		updated, err := s.orders.ByIDTx(ctx, tx, order.ID)
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

func (s *DroneService) MarkDelivered(ctx context.Context, principalID, orderID uint) (*domain.Order, error) {
	return s.completeCarryingOrder(ctx, principalID, orderID, domain.OrderStatusDelivered, domain.EventOrderDelivered)
}

func (s *DroneService) MarkFailed(ctx context.Context, principalID, orderID uint) (*domain.Order, error) {
	return s.completeCarryingOrder(ctx, principalID, orderID, domain.OrderStatusFailed, domain.EventOrderFailed)
}

func (s *DroneService) completeCarryingOrder(ctx context.Context, principalID, orderID uint, to domain.OrderStatus, eventType domain.OrderEventType) (*domain.Order, error) {
	now := s.clock.Now()
	var out *domain.Order
	err := database.WithTx(s.db, func(tx *gorm.DB) error {
		drone, err := s.drones.ByPrincipalID(ctx, principalID)
		if err != nil {
			return err
		}
		drone, err = s.drones.ByIDTx(ctx, tx, drone.ID)
		if err != nil {
			return err
		}
		order, err := s.orders.ByIDTx(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if order.AssignedDroneID == nil || *order.AssignedDroneID != drone.ID {
			return domain.ErrForbidden
		}
		if order.Status != domain.OrderStatusPickedUp {
			return domain.ErrInvalidTransition
		}
		if !domain.CanTransitionOrder(order.Status, to) {
			return domain.ErrInvalidTransition
		}
		from := order.Status
		patch := repo.OrderTransitionPatch{ClearETA: true}
		if err := s.orders.TransitionStatusTx(ctx, tx, order.ID, order.Version, from, to, patch, now); err != nil {
			return err
		}
		if err := s.drones.TransitionStatusTx(ctx, tx, drone.ID, drone.Version, drone.Status, domain.DroneStatusAvailable, true, nil, now); err != nil {
			return err
		}
		actor := principalID
		toCopy := to
		if err := s.events.LogTx(ctx, tx, repo.LogInput{
			OrderID:          order.ID,
			EventType:        eventType,
			FromStatus:       &from,
			ToStatus:         &toCopy,
			ActorPrincipalID: &actor,
			Lat:              drone.LastLat,
			Lng:              drone.LastLng,
		}, now); err != nil {
			return err
		}
		updated, err := s.orders.ByIDTx(ctx, tx, order.ID)
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

// MarkSelfBroken: drone reports broken. If carrying an order, transition order
// PICKED_UP→HANDOFF_REQUIRED and open a HANDOFF_PICKUP job for it.
func (s *DroneService) MarkSelfBroken(ctx context.Context, principalID uint) (*domain.Drone, error) {
	return s.markBroken(ctx, principalID, nil /* admin actor unspecified */, true)
}

func (s *DroneService) AdminMarkBroken(ctx context.Context, droneID, actorID uint) (*domain.Drone, error) {
	return s.markBrokenByDroneID(ctx, droneID, &actorID, false)
}

func (s *DroneService) AdminMarkFixed(ctx context.Context, droneID, actorID uint) (*domain.Drone, error) {
	now := s.clock.Now()
	var out *domain.Drone
	err := database.WithTx(s.db, func(tx *gorm.DB) error {
		drone, err := s.drones.ByIDTx(ctx, tx, droneID)
		if err != nil {
			return err
		}
		if drone.Status != domain.DroneStatusBroken {
			return domain.ErrInvalidTransition
		}
		if err := s.drones.TransitionStatusTx(ctx, tx, drone.ID, drone.Version, domain.DroneStatusBroken, domain.DroneStatusAvailable, false, nil, now); err != nil {
			return err
		}
		// NOTE: by design we do NOT cancel the outstanding handoff job.
		updated, err := s.drones.ByIDTx(ctx, tx, drone.ID)
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

func (s *DroneService) markBroken(ctx context.Context, principalID uint, actorOverride *uint, selfReport bool) (*domain.Drone, error) {
	drone, err := s.drones.ByPrincipalID(ctx, principalID)
	if err != nil {
		return nil, err
	}
	actor := principalID
	if actorOverride != nil {
		actor = *actorOverride
	}
	return s.markBrokenByDroneID(ctx, drone.ID, &actor, selfReport)
}

func (s *DroneService) markBrokenByDroneID(ctx context.Context, droneID uint, actorID *uint, selfReport bool) (*domain.Drone, error) {
	now := s.clock.Now()
	var out *domain.Drone
	err := database.WithTx(s.db, func(tx *gorm.DB) error {
		drone, err := s.drones.ByIDTx(ctx, tx, droneID)
		if err != nil {
			return err
		}
		if drone.Status == domain.DroneStatusBroken {
			return domain.ErrInvalidTransition
		}
		carryingOrderID := drone.CurrentOrderID

		// 1. drone → BROKEN, clear current_order_id
		if err := s.drones.TransitionStatusTx(ctx, tx, drone.ID, drone.Version, drone.Status, domain.DroneStatusBroken, true, nil, now); err != nil {
			return err
		}

		// 2. if carrying, order → HANDOFF_REQUIRED + new HANDOFF_PICKUP job
		if carryingOrderID != nil {
			order, err := s.orders.ByIDTx(ctx, tx, *carryingOrderID)
			if err != nil {
				return err
			}
			if order.Status == domain.OrderStatusPickedUp {
				from := order.Status
				to := domain.OrderStatusHandoffRequired
				patch := repo.OrderTransitionPatch{
					ClearDrone: true,
					ClearETA:   true,
				}
				// pickup location for handoff = drone's last known lat/lng
				if drone.LastLat != nil && drone.LastLng != nil {
					lat, lng := *drone.LastLat, *drone.LastLng
					patch.CurrentLat = &lat
					patch.CurrentLng = &lng
				}
				if err := s.orders.TransitionStatusTx(ctx, tx, order.ID, order.Version, from, to, patch, now); err != nil {
					return err
				}
				srcDrone := drone.ID
				if _, err := s.jobs.CreateTx(ctx, tx, order.ID, domain.JobTypeHandoffPickup, &srcDrone, now); err != nil {
					return err
				}
				if err := s.events.LogTx(ctx, tx, repo.LogInput{
					OrderID:          order.ID,
					EventType:        domain.EventOrderHandoffOpened,
					FromStatus:       &from,
					ToStatus:         &to,
					ActorPrincipalID: actorID,
					Lat:              patch.CurrentLat,
					Lng:              patch.CurrentLng,
				}, now); err != nil {
					return err
				}
			}
		}
		// 3. broker-event always logged (no order: write a sentinel event with order_id=0 omitted? skip if no order)
		if carryingOrderID != nil {
			if err := s.events.LogTx(ctx, tx, repo.LogInput{
				OrderID:          *carryingOrderID,
				EventType:        domain.EventDroneBroken,
				ActorPrincipalID: actorID,
			}, now); err != nil {
				return err
			}
		}
		updated, err := s.drones.ByIDTx(ctx, tx, drone.ID)
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

// Heartbeat updates the drone's lat/lng + heartbeat timestamp and, if it is
// carrying an order, propagates location and ETA to that order.
func (s *DroneService) Heartbeat(ctx context.Context, principalID uint, lat, lng float64) (*domain.Drone, *domain.Order, error) {
	if !validLatLng(lat, lng) {
		return nil, nil, domain.ErrInvalidInput
	}
	now := s.clock.Now()
	var outDrone *domain.Drone
	var outOrder *domain.Order
	err := database.WithTx(s.db, func(tx *gorm.DB) error {
		drone, err := s.drones.ByPrincipalID(ctx, principalID)
		if err != nil {
			return err
		}
		drone, err = s.drones.ByIDTx(ctx, tx, drone.ID)
		if err != nil {
			return err
		}
		res := tx.WithContext(ctx).Model(&repo.DroneRow{}).
			Where("id = ?", drone.ID).
			Updates(map[string]any{
				"last_lat":          lat,
				"last_lng":          lng,
				"last_heartbeat_at": now,
				"updated_at":        now,
			})
		if res.Error != nil {
			return res.Error
		}
		if drone.CurrentOrderID != nil {
			order, err := s.orders.ByIDTx(ctx, tx, *drone.CurrentOrderID)
			if err != nil {
				return err
			}
			if order.Status == domain.OrderStatusPickedUp {
				eta := utils.ETASeconds(lat, lng, order.DestLat, order.DestLng, s.avgSpeedMS)
				if err := s.orders.UpdateLocationTx(ctx, tx, order.ID, lat, lng, eta, now); err != nil {
					return err
				}
				refreshed, err := s.orders.ByIDTx(ctx, tx, order.ID)
				if err != nil {
					return err
				}
				outOrder = refreshed
			}
		}
		updated, err := s.drones.ByIDTx(ctx, tx, drone.ID)
		if err != nil {
			return err
		}
		outDrone = updated
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return outDrone, outOrder, nil
}

// AssignedOrder returns the order currently assigned to a drone, or nil if none.
func (s *DroneService) AssignedOrder(ctx context.Context, principalID uint) (*domain.Drone, *domain.Order, error) {
	drone, err := s.drones.ByPrincipalID(ctx, principalID)
	if err != nil {
		return nil, nil, err
	}
	if drone.CurrentOrderID == nil {
		return drone, nil, nil
	}
	order, err := s.orders.ByID(ctx, *drone.CurrentOrderID)
	if err != nil {
		return drone, nil, err
	}
	return drone, order, nil
}
