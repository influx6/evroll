package evroll

import (
	"fmt"
	"github.com/influx6/immute"
)

type Next func(i interface{})
type Callable func(i interface{}, f func(g interface{}))
type Callabut func(i interface{})
type CallList []Callable

type Event interface{}
type EventHandler func(e Event)

type RollerInterface interface {
	Size() int
	DecidedDone(r Callable)
	ReceiveDone(r Callabut)
	Decide(r Callable)
	Receive(r Callabut)
	Munch(i interface{})
	RevMunch(i interface{})
	CallAt(f int)
	CallDoneAt(f int)
	ReverseCallAt(f int)
	ReverseDoneCallAt(f int)
	onRoller(i interface{}, f func(g interface{}))
	onRevRoller(i interface{}, f func(g interface{}))
}

type Roller struct {
	enders []Callable
	doners []Callable
}

type Streamable interface {
	Send(interface{})
}

type Streams struct {
	*Roller
	Buffer  *immute.Sequence
	active  bool
	reverse bool
}

func (s *Streams) Init() {
	s.ReceiveDone(func(data interface{}) {
		s.active = false
		s.Stream()
	})
}

func (s *Streams) Send(data interface{}) {
	s.Buffer.Add(data, nil)

	if s.active {
		return
	}

	s.Stream()
}

func (s *Streams) Stream() {
	listeners := s.Size()

	if listeners <= 0 {
		return
	}

	size := s.Buffer.Length()

	if size <= 0 {
		return
	}

	cur := s.Buffer.Delete(0)

	if s.reverse {
		s.RevMunch(cur)
	} else {
		s.Munch(cur)
	}

	s.active = true
}

func (r *Roller) onRoller(i interface{}, next func(g interface{})) {
	r.Munch(i)
	next(nil)
}

func (r *Roller) onRevRoller(i interface{}, next func(g interface{})) {
	r.RevMunch(i)
	next(nil)
}

func (r *Roller) Munch(i interface{}) {
	r.CallAt(0, i)
}

func (r *Roller) RevMunch(i interface{}) {
	r.ReverseCallAt(0, i)
}

func (r *Roller) ReverseDoneCallAt(i int, g interface{}) {
	if len(r.doners) <= 0 {
		return
	}
	total := len(r.doners) - 1
	loc := total - i

	if loc >= 0 {
		val := r.doners[loc]
		if val != nil {
			val(g, func(f interface{}) {

				ind := i + 1
				if f == nil {
					r.ReverseDoneCallAt(ind, g)
					return
				}

				r.ReverseDoneCallAt(ind, f)
			})
		}
	}
}

func (r *Roller) ReverseCallAt(i int, g interface{}) {
	if len(r.enders) <= 0 {
		return
	}
	total := len(r.enders) - 1
	loc := total - i

	if loc >= 0 {
		val := r.enders[loc]
		if val != nil {
			val(g, func(f interface{}) {

				ind := i + 1
				if f == nil {
					r.ReverseCallAt(ind, g)
					return
				}

				r.ReverseCallAt(ind, f)
			})
		}
	} else {
		r.ReverseDoneCallAt(0, g)
	}
}

func (r *Roller) CallAt(i int, g interface{}) {
	if len(r.enders) <= 0 {
		return
	}

	if len(r.enders) > i {
		val := r.enders[i]
		if val != nil {
			val(g, func(f interface{}) {

				ind := i + 1
				if f == nil {
					r.CallAt(ind, g)
					return
				}

				r.CallAt(ind, f)
			})
		}
	} else {
		r.CallDoneAt(0, g)
	}
}

func (r *Roller) CallDoneAt(i int, g interface{}) {
	if len(r.doners) <= 0 {
		return
	}

	if len(r.doners) > i {
		val := r.doners[i]
		if val != nil {
			val(g, func(f interface{}) {

				ind := i + 1
				if f == nil {
					r.CallDoneAt(ind, g)
					return
				}

				r.CallDoneAt(ind, f)
			})
		}
	}
}

func (r *Roller) ReceiveDone(f func(c interface{})) {
	r.DecidedDone(func(i interface{}, next func(t interface{})) {
		f(i)
		next(nil)
	})
}

func (r *Roller) DecidedDone(f ...Callable) {
	r.doners = append(r.doners, f...)
}

func (r *Roller) Receive(f func(c interface{})) {
	r.Decide(func(i interface{}, next func(t interface{})) {
		f(i)
		next(nil)
	})
}

func (r *Roller) Decide(f ...Callable) {
	r.enders = append(r.enders, f...)
}

func (r *Roller) Size() int {
	return len(r.enders)
}

func (r *Roller) String() string {
	return fmt.Sprint(r.enders)
}

func NewRoller() *Roller {
	return &Roller{[]Callable{}, []Callable{}}
}

type EventRoll struct {
	Handlers []EventHandler
	Id       string
}

func (e *EventRoll) Listen(f ...EventHandler) {
	e.Handlers = append(e.Handlers, f...)
}

func (e *EventRoll) Emit(f Event) {
	if len(e.Handlers) <= 0 {
		return
	}

	for _, cur := range e.Handlers {
		cur(f)
	}
}

func NewEvents(id string) *EventRoll {
	return &EventRoll{make([]EventHandler, 0), id}
}

func NewStream(reverse bool) *Streams {
	list := immute.CreateList(make([]interface{}, 0))
	s := &Streams{NewRoller(), list, false, reverse}
	s.Init()
	return s
}
