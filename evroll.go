package evroll

import (
	"fmt"
	"github.com/influx6/immute"
)

type Next func(i interface{})
type Callable func(i interface{}, f func(g interface{}))
type Callabut func(i interface{})
type CallList []Callable

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
	ReverseCallDoneAt(f int)
	onRoller(i interface{}, f func(g interface{}))
	onRevRoller(i interface{}, f func(g interface{}))
}

type Roller struct {
	enders []Callable
	doners []Callable
}

type Eventables interface {
	Emit(interface{})
	Listen(Callabut)
}

type StreamPackers interface {
	Flush()
	Send(interface{})
	Subscribe(bool,bool)
	WeakSubscribe(bool,bool)
}

type Streamable interface {
	Send(interface{})
	Seq() *immute.Sequence
	Drain(Callabut)
	CollectTo(Callabut)
	Collect() []interface{}
	CollectAndStream()
	NotifyDrain()
	Clear()
	Reverse()
	Unreverse()
	BufferSize() int
}

type Streams struct {
	*Roller
	Buffer  *immute.Sequence
	Drains  *EventRoll
	manual  bool
	reverse bool
}

type EventRoll struct {
	Handlers *immute.Sequence
	Id       string
}

//Provides a persistent streams data
type StreamPack struct {
	Buffer  *immute.Sequence
	Adds	*EventRoll
}

func (s *StreamPack) Flush() {
	s.Buffer.Clear()
}

func (s *StreamPack) Send(data interface{}) {
	s.Adds.Emit(data)
}

func (s *StreamPack) WeakSubscribe(reverse bool,manual bool) *Streams {
	sm := NewStream(reverse,manual)

	s.Buffer.Each(func(data interface{}, key interface{}) interface{} {
		sm.Send(data)
		return nil
	}, func(_ int, _ interface{}) {})

	return sm
}

func (s *StreamPack) Subscribe(reverse bool,manual bool) *Streams {
	sm := NewStream(reverse,manual)
	s.Buffer.Each(func(data interface{}, key interface{}) interface{} {
		sm.Send(data)
		return nil
	}, func(_ int, _ interface{}) {})

	s.Adds.Listen(func( data interface{}){
		sm.Send(data)
	})

	return sm
}

func (e *EventRoll) Listen(f ...Callabut) {
	for _, v := range f {
		e.Handlers.Add(v, nil)
	}
}

func (e *EventRoll) Emit(val interface{}) {
	e.Handlers.Each(func(data interface{}, key interface{}) interface{} {
		fn, ok := data.(Callabut)

		if !ok {
			return nil
		}

		fn(val)
		return nil
	}, func(_ int, _ interface{}) {})
}

func (s *Streams) Seq() *immute.Sequence {
	return s.Buffer.Seq()
}

func (s *Streams) Drain(drainer ...Callabut) {
	s.Drains.Listen(drainer...)
}

func (s *Streams) Collect() []interface{} {
	data, ok := s.Buffer.Obj().([]interface{})

	if !ok {
		return nil
	}

	s.Buffer.Clear()
	return data
}

func (s *Streams) CollectAndStream() {
	s.Send(s.Collect())
}

func (s *Streams) CollectTo(fn func(data []interface{})) {
	var data = s.Collect()
	fn(data)
}

func (s *Streams) Send(data interface{}) {
	if s.Size() > 0{
		if !s.manual{
			if s.BufferSize() > 0 {
				s.Buffer.Add(data,nil)
			}else{
				s.Delegate(data)
				s.Drains.Emit(data)
			}
		}else{
			s.Buffer.Add(data,nil)
		}
	}else{
		s.Buffer.Add(data,nil)
	}

	if !s.manual {
		s.Stream()
	}
}

func (s *Streams) Delegate(data interface{}){
	listeners := s.Size()

	if listeners <= 0 {
		return
	}

	if s.reverse {
		s.RevMunch(data)
	} else {
		s.Munch(data)
	}

}

func (s *Streams) Stream() {

	listeners := s.Size()

	if listeners <= 0 {
		return
	}

	if s.BufferSize() <= 0 {
		return
	}

	if s.BufferSize() > 0{
		cur := s.Buffer.Delete(0)

		if s.BufferSize() <= 0 {
			s.Delegate(cur)
			s.Drains.Emit(cur)
		}else{
			s.Delegate(cur)
		}
	}

}

func (s *Streams) BufferSize() int {
	return s.Buffer.Length()
}

func (s *Streams) Reverse() {
	s.reverse = true
}

func (s *Streams) Unreverse() {
	s.reverse = false
}

func (s *Streams) Clear() {
	s.Buffer.Clear()
}

func (s *Streams) NotifyDrain(data interface{}) {
		s.Drains.Emit(data)
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

func (r *Roller) ReverseCallDoneAt(i int, g interface{}) {
	if len(r.doners) <= 0 {
		return
	}
	total := len(r.doners) - 1
	loc := total - i

	if loc >= 0 {
		val := r.doners[loc]
		if val != nil {
			val(g, func(fval interface{}) {

				ind := i + 1
				if fval == nil {
					r.ReverseCallDoneAt(ind, g)
					return
				}

				r.ReverseCallDoneAt(ind, fval)
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
			val(g, func(fval interface{}) {

				ind := i + 1

				if fval == nil {
					r.ReverseCallAt(ind, g)
					return
				}

				r.ReverseCallAt(ind, fval)
			})
		}
	} else {
		r.ReverseCallDoneAt(0, g)
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

//NewRoller creates and return a pointer to a new roller struct ready for middleware style pattern stack
func NewRoller() *Roller {
	roll := &Roller{
		[]Callable{},
		[]Callable{},
	}
	return roll
}

//NewEvent creates a new eventroll for event notification to its callbacks
func NewEvent(id string) *EventRoll {
	list := immute.CreateList(make([]interface{}, 0))
	return &EventRoll{list, id}
}

/*NewStream creates a new Stream struct and accepts two bool values:
reverse bool: indicate wether callback queue be called in reverse or not
manaul bool: indicates wether it should be a push model or a pull model
(push means every addition of data calls the notification of callbacks immediately)
(pull) means the Stream() method is called by the caller when ready to notify callbacks
*/
func NewStream(reverse bool, manaul bool) *Streams {

	//our internal buffer
	list := immute.CreateList(make([]interface{}, 0))

	//drain event handler
	drains := NewEvent("drain")

	sm := &Streams{
		NewRoller(),
		list,
		drains,
		manaul,
		reverse,
	}

	sm.ReceiveDone(func(data interface{}) {
			sm.Stream()
	})

	return sm
}

/*
NewStreamPack returns a new stream pack that allows persistence of stream data
*/
func NewStreamPack() *StreamPack {
	//our internal buffer
	list := immute.CreateList(make([]interface{}, 0))
	data := NewEvent("data")
	sm := &StreamPack{ list,data }

	data.Listen(func(i interface{}){
		list.Add(i,nil)
	})

	return sm
}
