package queue

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"remote_judge/internal/domain"
)

// RabbitMQQueue 提供基于 RabbitMQ 的队列实现。
type RabbitMQQueue struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string
	consumer  chan domain.SubmissionMessage
	once      sync.Once
}

// NewRabbitMQQueue 创建 RabbitMQ 队列。
func NewRabbitMQQueue(url, queueName string) (*RabbitMQQueue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	return &RabbitMQQueue{
		conn:      conn,
		ch:        ch,
		queueName: queueName,
		consumer:  make(chan domain.SubmissionMessage, 128),
	}, nil
}

// Publish 将一条提交消息发送到 RabbitMQ。
func (q *RabbitMQQueue) Publish(ctx context.Context, msg domain.SubmissionMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return q.ch.PublishWithContext(ctx, "", q.queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

// Consume 启动 RabbitMQ 消费并暴露本地消息流。
func (q *RabbitMQQueue) Consume(ctx context.Context) (<-chan domain.SubmissionMessage, error) {
	deliveries, err := q.ch.Consume(q.queueName, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	q.once.Do(func() {
		go func() {
			defer close(q.consumer)
			for {
				select {
				case <-ctx.Done():
					return
				case delivery, ok := <-deliveries:
					if !ok {
						return
					}
					var msg domain.SubmissionMessage
					if json.Unmarshal(delivery.Body, &msg) == nil {
						q.consumer <- msg
					}
				}
			}
		}()
	})
	return q.consumer, nil
}

// Close 关闭 RabbitMQ 连接。
func (q *RabbitMQQueue) Close() error {
	var firstErr error
	if q.ch != nil {
		firstErr = q.ch.Close()
	}
	if q.conn != nil {
		if err := q.conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if errors.Is(firstErr, amqp.ErrClosed) {
		return nil
	}
	return firstErr
}
