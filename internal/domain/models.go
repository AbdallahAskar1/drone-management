package domain

import "time"

type Principal struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Drone struct {
	ID              uint        `json:"id"`
	PrincipalID     uint        `json:"principal_id"`
	Status          DroneStatus `json:"status"`
	CurrentOrderID  *uint       `json:"current_order_id,omitempty"`
	LastLat         *float64    `json:"last_lat,omitempty"`
	LastLng         *float64    `json:"last_lng,omitempty"`
	LastHeartbeatAt *time.Time  `json:"last_heartbeat_at,omitempty"`
	Version         uint        `json:"version"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type Order struct {
	ID                 uint        `json:"id"`
	EnduserPrincipalID uint        `json:"enduser_principal_id"`
	Status             OrderStatus `json:"status"`
	OriginLat          float64     `json:"origin_lat"`
	OriginLng          float64     `json:"origin_lng"`
	DestLat            float64     `json:"dest_lat"`
	DestLng            float64     `json:"dest_lng"`
	AssignedDroneID    *uint       `json:"assigned_drone_id,omitempty"`
	CurrentLat         *float64    `json:"current_lat,omitempty"`
	CurrentLng         *float64    `json:"current_lng,omitempty"`
	ETASeconds         *int64      `json:"eta_seconds,omitempty"`
	Version            uint        `json:"version"`

	CreatedAt         time.Time  `json:"created_at"`
	ReadyAt           *time.Time `json:"ready_at,omitempty"`
	ReservedAt        *time.Time `json:"reserved_at,omitempty"`
	PickedUpAt        *time.Time `json:"picked_up_at,omitempty"`
	DeliveredAt       *time.Time `json:"delivered_at,omitempty"`
	FailedAt          *time.Time `json:"failed_at,omitempty"`
	HandoffRequiredAt *time.Time `json:"handoff_required_at,omitempty"`
	WithdrawnAt       *time.Time `json:"withdrawn_at,omitempty"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Job struct {
	ID                uint       `json:"id"`
	OrderID           uint       `json:"order_id"`
	Type              JobType    `json:"type"`
	Status            JobStatus  `json:"status"`
	SourceDroneID     *uint      `json:"source_drone_id,omitempty"`
	ReservedByDroneID *uint      `json:"reserved_by_drone_id,omitempty"`
	Version           uint       `json:"version"`
	CreatedAt         time.Time  `json:"created_at"`
	ReservedAt        *time.Time `json:"reserved_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

type OrderEvent struct {
	ID               uint           `json:"id"`
	OrderID          uint           `json:"order_id"`
	EventType        OrderEventType `json:"event_type"`
	FromStatus       *OrderStatus   `json:"from_status,omitempty"`
	ToStatus         *OrderStatus   `json:"to_status,omitempty"`
	ActorPrincipalID *uint          `json:"actor_principal_id,omitempty"`
	Lat              *float64       `json:"lat,omitempty"`
	Lng              *float64       `json:"lng,omitempty"`
	MetadataJSON     string         `json:"metadata,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
}
