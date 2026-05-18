package repo

import (
	"time"

	"drone-management/internal/domain"
)

type PrincipalRow struct {
	ID        uint        `gorm:"primaryKey"`
	Name      string      `gorm:"size:128;not null;index"`
	Role      domain.Role `gorm:"size:16;not null;index"`
	CreatedAt time.Time
}

func (PrincipalRow) TableName() string { return "principals" }

type DroneRow struct {
	ID              uint               `gorm:"primaryKey"`
	PrincipalID     uint               `gorm:"not null;uniqueIndex"`
	Status          domain.DroneStatus `gorm:"size:16;not null;index"`
	CurrentOrderID  *uint              `gorm:"index"`
	LastLat         *float64
	LastLng         *float64
	LastHeartbeatAt *time.Time
	Version         uint `gorm:"not null;default:0"`
	UpdatedAt       time.Time
}

func (DroneRow) TableName() string { return "drones" }

type OrderRow struct {
	ID                 uint               `gorm:"primaryKey"`
	EnduserPrincipalID uint               `gorm:"not null;index:idx_orders_enduser"`
	Status             domain.OrderStatus `gorm:"size:24;not null;index"`
	OriginLat          float64            `gorm:"not null"`
	OriginLng          float64            `gorm:"not null"`
	DestLat            float64            `gorm:"not null"`
	DestLng            float64            `gorm:"not null"`
	AssignedDroneID    *uint              `gorm:"index"`
	CurrentLat         *float64
	CurrentLng         *float64
	ETASeconds         *int64
	Version            uint `gorm:"not null;default:0"`

	CreatedAt         time.Time
	ReadyAt           *time.Time
	ReservedAt        *time.Time
	PickedUpAt        *time.Time
	DeliveredAt       *time.Time
	FailedAt          *time.Time
	HandoffRequiredAt *time.Time
	WithdrawnAt       *time.Time
	UpdatedAt         time.Time
}

func (OrderRow) TableName() string { return "orders" }

type JobRow struct {
	ID                uint             `gorm:"primaryKey"`
	OrderID           uint             `gorm:"not null;index"`
	Type              domain.JobType   `gorm:"size:24;not null;index:idx_jobs_status_type"`
	Status            domain.JobStatus `gorm:"size:16;not null;index:idx_jobs_status_type"`
	SourceDroneID     *uint
	ReservedByDroneID *uint
	Version           uint `gorm:"not null;default:0"`
	CreatedAt         time.Time
	ReservedAt        *time.Time
	CompletedAt       *time.Time
}

func (JobRow) TableName() string { return "jobs" }

type OrderEventRow struct {
	ID               uint                  `gorm:"primaryKey"`
	OrderID          uint                  `gorm:"not null;index:idx_events_order"`
	EventType        domain.OrderEventType `gorm:"size:32;not null"`
	FromStatus       *domain.OrderStatus   `gorm:"size:24"`
	ToStatus         *domain.OrderStatus   `gorm:"size:24"`
	ActorPrincipalID *uint
	Lat              *float64
	Lng              *float64
	MetadataJSON     *string   `gorm:"type:jsonb"`
	CreatedAt        time.Time `gorm:"index:idx_events_order"`
}

func (OrderEventRow) TableName() string { return "order_events" }

func principalToDomain(r *PrincipalRow) *domain.Principal {
	return &domain.Principal{
		ID:        r.ID,
		Name:      r.Name,
		Role:      r.Role,
		CreatedAt: r.CreatedAt,
	}
}

func droneToDomain(r *DroneRow) *domain.Drone {
	return &domain.Drone{
		ID:              r.ID,
		PrincipalID:     r.PrincipalID,
		Status:          r.Status,
		CurrentOrderID:  r.CurrentOrderID,
		LastLat:         r.LastLat,
		LastLng:         r.LastLng,
		LastHeartbeatAt: r.LastHeartbeatAt,
		Version:         r.Version,
		UpdatedAt:       r.UpdatedAt,
	}
}

func orderToDomain(r *OrderRow) *domain.Order {
	return &domain.Order{
		ID:                 r.ID,
		EnduserPrincipalID: r.EnduserPrincipalID,
		Status:             r.Status,
		OriginLat:          r.OriginLat,
		OriginLng:          r.OriginLng,
		DestLat:            r.DestLat,
		DestLng:            r.DestLng,
		AssignedDroneID:    r.AssignedDroneID,
		CurrentLat:         r.CurrentLat,
		CurrentLng:         r.CurrentLng,
		ETASeconds:         r.ETASeconds,
		Version:            r.Version,
		CreatedAt:          r.CreatedAt,
		ReadyAt:            r.ReadyAt,
		ReservedAt:         r.ReservedAt,
		PickedUpAt:         r.PickedUpAt,
		DeliveredAt:        r.DeliveredAt,
		FailedAt:           r.FailedAt,
		HandoffRequiredAt:  r.HandoffRequiredAt,
		WithdrawnAt:        r.WithdrawnAt,
		UpdatedAt:          r.UpdatedAt,
	}
}

func jobToDomain(r *JobRow) *domain.Job {
	return &domain.Job{
		ID:                r.ID,
		OrderID:           r.OrderID,
		Type:              r.Type,
		Status:            r.Status,
		SourceDroneID:     r.SourceDroneID,
		ReservedByDroneID: r.ReservedByDroneID,
		Version:           r.Version,
		CreatedAt:         r.CreatedAt,
		ReservedAt:        r.ReservedAt,
		CompletedAt:       r.CompletedAt,
	}
}

func eventToDomain(r *OrderEventRow) *domain.OrderEvent {
	var meta string
	if r.MetadataJSON != nil {
		meta = *r.MetadataJSON
	}
	return &domain.OrderEvent{
		ID:               r.ID,
		OrderID:          r.OrderID,
		EventType:        r.EventType,
		FromStatus:       r.FromStatus,
		ToStatus:         r.ToStatus,
		ActorPrincipalID: r.ActorPrincipalID,
		Lat:              r.Lat,
		Lng:              r.Lng,
		MetadataJSON:     meta,
		CreatedAt:        r.CreatedAt,
	}
}
