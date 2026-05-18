package handler

import (
	"drone-management/internal/domain"
)

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type TokenRequest struct {
	Name string      `json:"name"`
	Role domain.Role `json:"role"`
}

type TokenResponse struct {
	Token     string            `json:"token"`
	Principal *domain.Principal `json:"principal"`
}

type SubmitOrderRequest struct {
	Origin      LatLng `json:"origin"`
	Destination LatLng `json:"destination"`
}

type OrderResponse struct {
	*domain.Order
	Timeline []*domain.OrderEvent `json:"timeline,omitempty"`
}

type HeartbeatRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type HeartbeatResponse struct {
	Drone *domain.Drone `json:"drone"`
	Order *domain.Order `json:"order,omitempty"`
}

type PatchOrderRequest struct {
	Origin      *LatLng `json:"origin,omitempty"`
	Destination *LatLng `json:"destination,omitempty"`
}
