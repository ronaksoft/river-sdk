package domain

import (
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// QueueMessages message envelop queue for network debouncer
type QueueMessages struct {
	mx     sync.Mutex
	items  []*msg.MessageEnvelope
	length int
}

// NewQueueMessages create new instance
func NewQueueMessages() *QueueMessages {
	return &QueueMessages{
		items: make([]*msg.MessageEnvelope, 0),
	}
}

// PushMany insert items to queue
func (q *QueueMessages) PushMany(m []*msg.MessageEnvelope) {
	q.mx.Lock()
	q.items = append(q.items, m...)
	q.length += len(m)
	q.mx.Unlock()
}

// Push insert item to queue
func (q *QueueMessages) Push(m *msg.MessageEnvelope) {
	q.mx.Lock()
	q.items = append(q.items, m)
	q.length++
	q.mx.Unlock()
}

// Pop pickup item from queue
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

// PopAll pick all items from queue
func (q *QueueMessages) PopAll() []*msg.MessageEnvelope {
	q.mx.Lock()
	m := q.items[:]
	q.length = 0
	q.items = make([]*msg.MessageEnvelope, 0)
	q.mx.Unlock()
	return m
}

// Clear queue
func (q *QueueMessages) Clear() {
	q.mx.Lock()
	q.length = 0
	q.items = make([]*msg.MessageEnvelope, 0)
	q.mx.Unlock()
}

// Length  size of queue
func (q *QueueMessages) Length() int {
	return q.length
}

// GetRawItems return underlying slice
func (q *QueueMessages) GetRawItems() []*msg.MessageEnvelope {
	return q.items
}
