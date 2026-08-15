package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer слушает очередь RabbitMQ и вызывает handler для каждого сообщения.
// Checkout использует его чтобы получать события OrderConfirmed от Basket.
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewConsumer создаёт подключение к RabbitMQ и объявляет очередь.
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

	// Объявляем тот же exchange что и Publisher (idempotent — не создаст дубль).
	if err := ch.ExchangeDeclare(
		ExchangeOrders,
		"fanout",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	// Объявляем очередь для checkout.
	// durable=true — очередь выживает после перезапуска RabbitMQ.
	if _, err := ch.QueueDeclare(
		QueueOrderConfirmed,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	// Привязываем очередь к exchange — теперь все события попадут в неё.
	if err := ch.QueueBind(
		QueueOrderConfirmed,
		"",             // routing key (fanout игнорирует)
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

// OrderConfirmedHandler — функция обработки события подтверждения заказа.
// Checkout передаёт сюда логику создания Order в БД.
type OrderConfirmedHandler func(ctx context.Context, event OrderConfirmedEvent) error

// Consume запускает бесконечный цикл чтения сообщений из очереди.
// Должен вызываться в отдельной горутине: go consumer.Consume(ctx, handler)
//
// При успехе вызывает Ack (сообщение удаляется из очереди).
// При ошибке вызывает Nack с requeue=false (сообщение уходит в dead-letter).
func (c *Consumer) Consume(ctx context.Context, handler OrderConfirmedHandler) error {
	msgs, err := c.channel.Consume(
		QueueOrderConfirmed,
		"checkout-consumer", // consumer tag (уникальный идентификатор)
		false,               // auto-ack=false — подтверждаем вручную после обработки
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
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
				// Сообщение невалидно — отклоняем без повтора
				msg.Nack(false, false)
				continue
			}

			if err := handler(ctx, event); err != nil {
				log.Printf("rabbitmq: handler error: %v — requeuing", err)
				// Ошибка бизнес-логики — возвращаем в очередь для повтора
				msg.Nack(false, true)
				continue
			}

			// Успех — подтверждаем получение, сообщение удаляется из очереди
			msg.Ack(false)
			log.Printf("rabbitmq: processed order for account=%s", event.AccountName)
		}
	}
}

// Close закрывает соединение.
func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
