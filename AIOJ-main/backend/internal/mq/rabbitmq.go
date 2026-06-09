package mq

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// SubmitTask is the payload pushed by the HTTP handler and consumed by the
// worker. We enqueue rather than persisting in the handler so MySQL writes
// happen off the request path (see skill.md point 5).
type SubmitTask struct {
	SubmissionID uint64    `json:"submission_id"`
	UserID       uint64    `json:"user_id"`
	ProblemID    uint64    `json:"problem_id"`
	ProblemTitle string    `json:"problem_title"`
	Language     string    `json:"language"`
	Code         string    `json:"code"`
	EnqueuedAt   time.Time `json:"enqueued_at"`
}

// Broker wraps a single RabbitMQ connection + channel. When disabled, it
// degrades to an in-process buffered channel so the backend keeps working
// in local/dev scenarios without RabbitMQ installed.
type Broker struct {
	enabled bool
	queue   string

	mu      sync.Mutex
	conn    *amqp.Connection
	chPub   *amqp.Channel
	chConsu *amqp.Channel

	fallback chan []byte
}

// NewBroker opens the connection. When enabled=false or the URL is empty,
// the returned Broker uses an in-memory channel instead.
func NewBroker(url, queue string, enabled bool) (*Broker, error) {
	b := &Broker{enabled: enabled, queue: queue}
	if !enabled || url == "" {
		b.fallback = make(chan []byte, 1024)
		log.Println("[mq] running in in-memory fallback mode (RabbitMQ disabled)")
		return b, nil
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	chPub, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if _, err := chPub.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		_ = conn.Close()
		return nil, err
	}
	b.conn = conn
	b.chPub = chPub
	log.Printf("[mq] connected to %s (queue=%s)", url, queue)
	return b, nil
}

// Publish enqueues a task for asynchronous processing.
func (b *Broker) Publish(ctx context.Context, task *SubmitTask) error {
	body, err := json.Marshal(task)
	if err != nil {
		return err
	}
	if !b.enabled {
		select {
		case b.fallback <- body:
			return nil
		default:
			return errors.New("mq fallback channel full")
		}
	}
	b.mu.Lock()
	ch := b.chPub
	b.mu.Unlock()
	return ch.PublishWithContext(ctx, "", b.queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
	})
}

// Consume returns a channel that yields raw task payloads. The caller is
// expected to json.Unmarshal each item into *SubmitTask.
func (b *Broker) Consume(ctx context.Context) (<-chan []byte, error) {
	if !b.enabled {
		return b.fallback, nil
	}
	b.mu.Lock()
	if b.chConsu == nil {
		ch, err := b.conn.Channel()
		if err != nil {
			b.mu.Unlock()
			return nil, err
		}
		if err := ch.Qos(8, 0, false); err != nil {
			b.mu.Unlock()
			return nil, err
		}
		b.chConsu = ch
	}
	ch := b.chConsu
	b.mu.Unlock()

	deliveries, err := ch.ConsumeWithContext(ctx, b.queue, "toj-worker", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	out := make(chan []byte, 16)
	go func() {
		defer close(out)
		for d := range deliveries {
			select {
			case out <- d.Body:
				_ = d.Ack(false)
			case <-ctx.Done():
				_ = d.Nack(false, true)
				return
			}
		}
	}()
	return out, nil
}

// Close releases resources.
func (b *Broker) Close() {
	if b.conn != nil {
		_ = b.conn.Close()
	}
}
