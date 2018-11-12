package domain

import (
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type QueueUpdates struct {
	mx     sync.Mutex
	items  []*msg.UpdateContainer
	length int
}

func NewQueueUpdates() *QueueUpdates {
	return &QueueUpdates{
		items: make([]*msg.UpdateContainer, 0),
	}
}

func (q *QueueUpdates) PushMany(m []*msg.UpdateContainer) {
	q.mx.Lock()
	q.items = append(q.items, m...)
	q.length += len(m)
	q.mx.Unlock()
}

func (q *QueueUpdates) Push(m *msg.UpdateContainer) {
	q.mx.Lock()
	q.items = append(q.items, m)
	q.length++
	q.mx.Unlock()
}

func (q *QueueUpdates) Pop() (*msg.UpdateContainer, error) {
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
func (q *QueueUpdates) PopAll() []*msg.UpdateContainer {
	q.mx.Lock()
	m := q.items[:]
	q.length = 0
	q.items = make([]*msg.UpdateContainer, 0)
	q.mx.Unlock()
	return m
}

func (q *QueueUpdates) Clear() {
	q.mx.Lock()
	q.length = 0
	q.items = make([]*msg.UpdateContainer, 0)
	q.mx.Unlock()
}

func (q *QueueUpdates) Length() int {
	return q.length
}

func (q *QueueUpdates) GetRawItems() []*msg.UpdateContainer {
	return q.items
}
