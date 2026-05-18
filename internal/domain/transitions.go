package domain

var orderTransitions = map[OrderStatus]map[OrderStatus]struct{}{
	OrderStatusCreated: {
		OrderStatusReadyForPickup: {},
		OrderStatusWithdrawn:      {},
	},
	OrderStatusReadyForPickup: {
		OrderStatusReserved:  {},
		OrderStatusWithdrawn: {},
	},
	OrderStatusReserved: {
		OrderStatusPickedUp:       {},
		OrderStatusReadyForPickup: {},
	},
	OrderStatusPickedUp: {
		OrderStatusDelivered:       {},
		OrderStatusFailed:          {},
		OrderStatusHandoffRequired: {},
	},
	OrderStatusHandoffRequired: {
		OrderStatusReserved: {},
		OrderStatusFailed:   {},
	},
}

func CanTransitionOrder(from, to OrderStatus) bool {
	allowed, ok := orderTransitions[from]
	if !ok {
		return false
	}
	_, ok = allowed[to]
	return ok
}

var droneTransitions = map[DroneStatus]map[DroneStatus]struct{}{
	DroneStatusAvailable: {
		DroneStatusBusy:   {},
		DroneStatusBroken: {},
	},
	DroneStatusBusy: {
		DroneStatusAvailable: {},
		DroneStatusBroken:    {},
	},
	DroneStatusBroken: {
		DroneStatusAvailable: {},
	},
}

func CanTransitionDrone(from, to DroneStatus) bool {
	allowed, ok := droneTransitions[from]
	if !ok {
		return false
	}
	_, ok = allowed[to]
	return ok
}
