package domain

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleEnduser Role = "enduser"
	RoleDrone   Role = "drone"
)

func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleEnduser, RoleDrone:
		return true
	}
	return false
}

type OrderStatus string

const (
	OrderStatusCreated         OrderStatus = "CREATED"
	OrderStatusReadyForPickup  OrderStatus = "READY_FOR_PICKUP"
	OrderStatusReserved        OrderStatus = "RESERVED"
	OrderStatusPickedUp        OrderStatus = "PICKED_UP"
	OrderStatusHandoffRequired OrderStatus = "HANDOFF_REQUIRED"
	OrderStatusDelivered       OrderStatus = "DELIVERED"
	OrderStatusFailed          OrderStatus = "FAILED"
	OrderStatusWithdrawn       OrderStatus = "WITHDRAWN"
)

func (s OrderStatus) Terminal() bool {
	switch s {
	case OrderStatusDelivered, OrderStatusFailed, OrderStatusWithdrawn:
		return true
	}
	return false
}

type DroneStatus string

const (
	DroneStatusAvailable DroneStatus = "AVAILABLE"
	DroneStatusBusy      DroneStatus = "BUSY"
	DroneStatusBroken    DroneStatus = "BROKEN"
)

type JobType string

const (
	JobTypeOriginPickup  JobType = "ORIGIN_PICKUP"
	JobTypeHandoffPickup JobType = "HANDOFF_PICKUP"
)

func (t JobType) Valid() bool {
	switch t {
	case JobTypeOriginPickup, JobTypeHandoffPickup:
		return true
	}
	return false
}

type JobStatus string

const (
	JobStatusOpen      JobStatus = "OPEN"
	JobStatusReserved  JobStatus = "RESERVED"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusCancelled JobStatus = "CANCELLED"
)

type OrderEventType string

const (
	EventOrderCreated       OrderEventType = "ORDER_CREATED"
	EventOrderReserved      OrderEventType = "ORDER_RESERVED"
	EventOrderPickedUp      OrderEventType = "ORDER_PICKED_UP"
	EventOrderDelivered     OrderEventType = "ORDER_DELIVERED"
	EventOrderFailed        OrderEventType = "ORDER_FAILED"
	EventOrderWithdrawn     OrderEventType = "ORDER_WITHDRAWN"
	EventOrderHandoffOpened OrderEventType = "ORDER_HANDOFF_OPENED"
	EventOrderUpdated       OrderEventType = "ORDER_UPDATED"
	EventDroneBroken        OrderEventType = "DRONE_BROKEN"
	EventDroneFixed         OrderEventType = "DRONE_FIXED"
	EventHeartbeat          OrderEventType = "HEARTBEAT"
)
