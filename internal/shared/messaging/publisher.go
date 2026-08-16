package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (

	ExchangeOrders = "orders"

	QueueOrderConfirmed = "order.confirmed"
)

type OrderConfirmedEvent struct {
	AccountName string             `json:"account_name"`
	Items       []OrderConfirmedItem `json:"items"`
	ConfirmedAt time.Time          `json:"confirmed_at"`
}

type OrderConfirmedItem struct {
	ItemID    string  `json:"item_id"`
	ItemTitle string  `json:"item_title"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Discount  float64 `json:"discount"`
}

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(amqpURL string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("amqp dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(
		ExchangeOrders,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	log.Printf("rabbitmq publisher connected, exchange=%s", ExchangeOrders)
	return &Publisher{conn: conn, channel: ch}, nil
}

func (p *Publisher) PublishOrderConfirmed(ctx context.Context, event OrderConfirmedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return p.channel.PublishWithContext(ctx,
		ExchangeOrders,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
