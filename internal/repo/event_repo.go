package repo

import (
	"context"
	"time"

	"drone-management/internal/domain"

	"gorm.io/gorm"
)

type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{db: db}
}

type LogInput struct {
	OrderID          uint
	EventType        domain.OrderEventType
	FromStatus       *domain.OrderStatus
	ToStatus         *domain.OrderStatus
	ActorPrincipalID *uint
	Lat              *float64
	Lng              *float64
	Metadata         string
}

func (r *EventRepo) LogTx(ctx context.Context, tx *gorm.DB, in LogInput, now time.Time) error {
	var meta *string
	if in.Metadata != "" {
		meta = &in.Metadata
	}
	row := OrderEventRow{
		OrderID:          in.OrderID,
		EventType:        in.EventType,
		FromStatus:       in.FromStatus,
		ToStatus:         in.ToStatus,
		ActorPrincipalID: in.ActorPrincipalID,
		Lat:              in.Lat,
		Lng:              in.Lng,
		MetadataJSON:     meta,
		CreatedAt:        now,
	}
	return tx.WithContext(ctx).Create(&row).Error
}

func (r *EventRepo) ByOrder(ctx context.Context, orderID uint) ([]*domain.OrderEvent, error) {
	var rows []OrderEventRow
	if err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.OrderEvent, 0, len(rows))
	for i := range rows {
		out = append(out, eventToDomain(&rows[i]))
	}
	return out, nil
}
