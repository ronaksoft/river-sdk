package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_CalendarGetEvents int64 = 1010730154

type poolCalendarGetEvents struct {
	pool sync.Pool
}

func (p *poolCalendarGetEvents) Get() *CalendarGetEvents {
	x, ok := p.pool.Get().(*CalendarGetEvents)
	if !ok {
		return &CalendarGetEvents{}
	}
	return x
}

func (p *poolCalendarGetEvents) Put(x *CalendarGetEvents) {
	x.From = 0
	x.To = 0
	x.Filter = 0
	p.pool.Put(x)
}

var PoolCalendarGetEvents = poolCalendarGetEvents{}

const C_CalendarSetEvent int64 = 3405460640

type poolCalendarSetEvent struct {
	pool sync.Pool
}

func (p *poolCalendarSetEvent) Get() *CalendarSetEvent {
	x, ok := p.pool.Get().(*CalendarSetEvent)
	if !ok {
		return &CalendarSetEvent{}
	}
	return x
}

func (p *poolCalendarSetEvent) Put(x *CalendarSetEvent) {
	x.Name = ""
	x.Date = 0
	x.StartRange = 0
	x.Duration = 0
	x.Recurring = false
	x.Period = 0
	x.AllDay = false
	x.Team = false
	x.Global = false
	p.pool.Put(x)
}

var PoolCalendarSetEvent = poolCalendarSetEvent{}

const C_CalendarEditEvent int64 = 2440838922

type poolCalendarEditEvent struct {
	pool sync.Pool
}

func (p *poolCalendarEditEvent) Get() *CalendarEditEvent {
	x, ok := p.pool.Get().(*CalendarEditEvent)
	if !ok {
		return &CalendarEditEvent{}
	}
	return x
}

func (p *poolCalendarEditEvent) Put(x *CalendarEditEvent) {
	x.EventID = 0
	x.Name = ""
	x.Date = 0
	x.StartRange = 0
	x.Duration = 0
	x.Recurring = false
	x.Period = 0
	x.AllDay = false
	x.Policy = 0
	p.pool.Put(x)
}

var PoolCalendarEditEvent = poolCalendarEditEvent{}

const C_CalendarRemoveEvent int64 = 3761579510

type poolCalendarRemoveEvent struct {
	pool sync.Pool
}

func (p *poolCalendarRemoveEvent) Get() *CalendarRemoveEvent {
	x, ok := p.pool.Get().(*CalendarRemoveEvent)
	if !ok {
		return &CalendarRemoveEvent{}
	}
	return x
}

func (p *poolCalendarRemoveEvent) Put(x *CalendarRemoveEvent) {
	x.EventID = 0
	p.pool.Put(x)
}

var PoolCalendarRemoveEvent = poolCalendarRemoveEvent{}

const C_CalendarEvent int64 = 1185062169

type poolCalendarEvent struct {
	pool sync.Pool
}

func (p *poolCalendarEvent) Get() *CalendarEvent {
	x, ok := p.pool.Get().(*CalendarEvent)
	if !ok {
		return &CalendarEvent{}
	}
	return x
}

func (p *poolCalendarEvent) Put(x *CalendarEvent) {
	x.ID = 0
	x.Name = ""
	x.Recurring = false
	x.Period = 0
	x.AllDay = false
	p.pool.Put(x)
}

var PoolCalendarEvent = poolCalendarEvent{}

const C_CalendarEventInstance int64 = 3586847608

type poolCalendarEventInstance struct {
	pool sync.Pool
}

func (p *poolCalendarEventInstance) Get() *CalendarEventInstance {
	x, ok := p.pool.Get().(*CalendarEventInstance)
	if !ok {
		return &CalendarEventInstance{}
	}
	return x
}

func (p *poolCalendarEventInstance) Put(x *CalendarEventInstance) {
	x.ID = 0
	x.EventID = 0
	x.Start = 0
	x.End = 0
	x.Colour = ""
	p.pool.Put(x)
}

var PoolCalendarEventInstance = poolCalendarEventInstance{}

func init() {
	registry.RegisterConstructor(1010730154, "CalendarGetEvents")
	registry.RegisterConstructor(3405460640, "CalendarSetEvent")
	registry.RegisterConstructor(2440838922, "CalendarEditEvent")
	registry.RegisterConstructor(3761579510, "CalendarRemoveEvent")
	registry.RegisterConstructor(1185062169, "CalendarEvent")
	registry.RegisterConstructor(3586847608, "CalendarEventInstance")
}

func (x *CalendarGetEvents) DeepCopy(z *CalendarGetEvents) {
	z.From = x.From
	z.To = x.To
	z.Filter = x.Filter
}

func (x *CalendarSetEvent) DeepCopy(z *CalendarSetEvent) {
	z.Name = x.Name
	z.Date = x.Date
	z.StartRange = x.StartRange
	z.Duration = x.Duration
	z.Recurring = x.Recurring
	z.Period = x.Period
	z.AllDay = x.AllDay
	z.Team = x.Team
	z.Global = x.Global
}

func (x *CalendarEditEvent) DeepCopy(z *CalendarEditEvent) {
	z.EventID = x.EventID
	z.Name = x.Name
	z.Date = x.Date
	z.StartRange = x.StartRange
	z.Duration = x.Duration
	z.Recurring = x.Recurring
	z.Period = x.Period
	z.AllDay = x.AllDay
	z.Policy = x.Policy
}

func (x *CalendarRemoveEvent) DeepCopy(z *CalendarRemoveEvent) {
	z.EventID = x.EventID
}

func (x *CalendarEvent) DeepCopy(z *CalendarEvent) {
	z.ID = x.ID
	z.Name = x.Name
	z.Recurring = x.Recurring
	z.Period = x.Period
	z.AllDay = x.AllDay
}

func (x *CalendarEventInstance) DeepCopy(z *CalendarEventInstance) {
	z.ID = x.ID
	z.EventID = x.EventID
	z.Start = x.Start
	z.End = x.End
	z.Colour = x.Colour
}

func (x *CalendarGetEvents) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CalendarGetEvents, x)
}

func (x *CalendarSetEvent) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CalendarSetEvent, x)
}

func (x *CalendarEditEvent) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CalendarEditEvent, x)
}

func (x *CalendarRemoveEvent) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CalendarRemoveEvent, x)
}

func (x *CalendarEvent) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CalendarEvent, x)
}

func (x *CalendarEventInstance) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CalendarEventInstance, x)
}

func (x *CalendarGetEvents) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CalendarSetEvent) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CalendarEditEvent) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CalendarRemoveEvent) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CalendarEvent) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CalendarEventInstance) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CalendarGetEvents) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CalendarSetEvent) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CalendarEditEvent) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CalendarRemoveEvent) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CalendarEvent) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CalendarEventInstance) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
