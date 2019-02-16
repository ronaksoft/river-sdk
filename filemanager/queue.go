package filemanager

import (
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/domain"
)

type QueueParts struct {
	mx     sync.Mutex
	items  []int64
	length int
}

func NewQueueParts() *QueueParts {
	return &QueueParts{
		items: make([]int64, 0),
	}
}

func (q *QueueParts) PushMany(m []int64) {
	q.mx.Lock()
	q.items = append(q.items, m...)
	q.length += len(m)
	q.mx.Unlock()
}

func (q *QueueParts) Push(m int64) {
	q.mx.Lock()
	q.items = append(q.items, m)
	q.length++
	q.mx.Unlock()
}

func (q *QueueParts) Pop() (int64, error) {
	if q.length > 0 {
		q.mx.Lock()
		m := q.items[0]
		q.length--
		q.items = q.items[1:]
		q.mx.Unlock()
		return m, nil
	}
	return -1, domain.ErrDoesNotExists
}
func (q *QueueParts) PopAll() []int64 {
	q.mx.Lock()
	m := q.items[:]
	q.length = 0
	q.items = make([]int64, 0)
	q.mx.Unlock()
	return m
}

func (q *QueueParts) Clear() {
	q.mx.Lock()
	q.length = 0
	q.items = make([]int64, 0)
	q.mx.Unlock()
}

func (q *QueueParts) Length() int {
	return q.length
}

func (q *QueueParts) GetRawItems() []int64 {
	return q.items[:]
}
