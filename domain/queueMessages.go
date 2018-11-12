package domain

import (
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type QueueMessages struct {
	mx     sync.Mutex
	items  []*msg.MessageEnvelope
	length int
}

func NewQueueMessages() *QueueMessages {
	return &QueueMessages{
		items: make([]*msg.MessageEnvelope, 0),
	}
}

func (q *QueueMessages) PushMany(m []*msg.MessageEnvelope) {
	q.mx.Lock()
	q.items = append(q.items, m...)
	q.length += len(m)
	q.mx.Unlock()
}

func (q *QueueMessages) Push(m *msg.MessageEnvelope) {
	q.mx.Lock()
	q.items = append(q.items, m)
	q.length++
	q.mx.Unlock()
}

func (q *QueueMessages) Pop() (*msg.MessageEnvelope, error) {
	if q.length > 0 {
		q.mx.Lock()
		m := q.items[0]
		q.length--
		q.items = q.items[1:]
		q.mx.Unlock()
		return m, nil
	}
	return nil, ErrDoesNotExists
}
func (q *QueueMessages) PopAll() []*msg.MessageEnvelope {
	q.mx.Lock()
	m := q.items[:]
	q.length = 0
	q.items = q.items[0:0]
	q.mx.Unlock()
	return m
}

func (q *QueueMessages) Clear() {
	q.mx.Lock()
	q.length = 0
	q.items = q.items[:]
	q.mx.Unlock()
}

func (q *QueueMessages) Length() int {
	return q.length
}

func (q *QueueMessages) GetRawItems() []*msg.MessageEnvelope {
	return q.items
}
