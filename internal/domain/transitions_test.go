package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCanTransitionOrder(t *testing.T) {
	tests := []struct {
		name string
		from OrderStatus
		to   OrderStatus
		want bool
	}{
		{"Created -> ReadyForPickup", OrderStatusCreated, OrderStatusReadyForPickup, true},
		{"Created -> Withdrawn", OrderStatusCreated, OrderStatusWithdrawn, true},
		{"Created -> Delivered", OrderStatusCreated, OrderStatusDelivered, false},
		{"PickedUp -> Delivered", OrderStatusPickedUp, OrderStatusDelivered, true},
		{"PickedUp -> HandoffRequired", OrderStatusPickedUp, OrderStatusHandoffRequired, true},
		{"HandoffRequired -> Reserved", OrderStatusHandoffRequired, OrderStatusReserved, true},
		{"Invalid from state", "UNKNOWN", OrderStatusCreated, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CanTransitionOrder(tt.from, tt.to))
		})
	}
}

func TestCanTransitionDrone(t *testing.T) {
	tests := []struct {
		name string
		from DroneStatus
		to   DroneStatus
		want bool
	}{
		{"Available -> Busy", DroneStatusAvailable, DroneStatusBusy, true},
		{"Busy -> Available", DroneStatusBusy, DroneStatusAvailable, true},
		{"Broken -> Available", DroneStatusBroken, DroneStatusAvailable, true},
		{"Available -> Broken", DroneStatusAvailable, DroneStatusBroken, true},
		{"Broken -> Busy", DroneStatusBroken, DroneStatusBusy, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CanTransitionDrone(tt.from, tt.to))
		})
	}
}
