package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// --- Имена очередей и exchange ---

const (
	// ExchangeOrders — fanout exchange для событий заказов.
	// fanout рассылает сообщение всем подписанным очередям.
	ExchangeOrders = "orders"

	// QueueOrderConfirmed — очередь событий подтверждения заказа.
	// Checkout слушает эту очередь чтобы создать Order.
	QueueOrderConfirmed = "order.confirmed"
)

// --- Событие ---

// OrderConfirmedEvent — событие которое Basket публикует когда пользователь
// нажимает "Оформить заказ". Checkout подписывается и создаёт заказ в БД.
type OrderConfirmedEvent struct {
	AccountName string             `json:"account_name"`
	Items       []OrderConfirmedItem `json:"items"`
	ConfirmedAt time.Time          `json:"confirmed_at"`
}

// OrderConfirmedItem — позиция в событии заказа.
type OrderConfirmedItem struct {
	ItemID    string  `json:"item_id"`
	ItemTitle string  `json:"item_title"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Discount  float64 `json:"discount"` // % скидки от promotion-сервиса
}

// --- Publisher ---

// Publisher публикует события в RabbitMQ.
// Basket использует его чтобы сообщить другим сервисам о подтверждении заказа.
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewPublisher создаёт подключение к RabbitMQ и возвращает Publisher.
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

	// Объявляем exchange типа fanout — рассылает всем подписчикам.
	// durable=true — exchange выживает после перезапуска RabbitMQ.
	if err := ch.ExchangeDeclare(
		ExchangeOrders, // имя
		"fanout",       // тип
		true,           // durable
		false,          // auto-delete
		false,          // internal
		false,          // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	log.Printf("rabbitmq publisher connected, exchange=%s", ExchangeOrders)
	return &Publisher{conn: conn, channel: ch}, nil
}

// PublishOrderConfirmed отправляет событие о подтверждении заказа.
func (p *Publisher) PublishOrderConfirmed(ctx context.Context, event OrderConfirmedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return p.channel.PublishWithContext(ctx,
		ExchangeOrders, // exchange
		"",             // routing key (fanout игнорирует)
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // сообщение переживёт перезапуск RabbitMQ
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

// Close закрывает соединение с RabbitMQ.
func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
