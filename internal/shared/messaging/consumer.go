package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewConsumer(amqpURL string) (*Consumer, error) {
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

	if _, err := ch.QueueDeclare(
		QueueOrderConfirmed,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.QueueBind(
		QueueOrderConfirmed,
		"",
		ExchangeOrders,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("bind queue: %w", err)
	}

	log.Printf("rabbitmq consumer connected, queue=%s", QueueOrderConfirmed)
	return &Consumer{conn: conn, channel: ch}, nil
}

type OrderConfirmedHandler func(ctx context.Context, event OrderConfirmedEvent) error

func (c *Consumer) Consume(ctx context.Context, handler OrderConfirmedHandler) error {
	msgs, err := c.channel.Consume(
		QueueOrderConfirmed,
		"checkout-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	log.Printf("rabbitmq: waiting for messages in queue=%s", QueueOrderConfirmed)

	for {
		select {
		case <-ctx.Done():
			log.Println("rabbitmq consumer: context cancelled, stopping")
			return nil

		case msg, ok := <-msgs:
			if !ok {
				log.Println("rabbitmq consumer: channel closed")
				return fmt.Errorf("channel closed")
			}

			var event OrderConfirmedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("rabbitmq: failed to unmarshal message: %v", err)

				msg.Nack(false, false)
				continue
			}

			if err := handler(ctx, event); err != nil {
				log.Printf("rabbitmq: handler error: %v — requeuing", err)

				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
			log.Printf("rabbitmq: processed order for account=%s", event.AccountName)
		}
	}
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
