package evroll

import (
	"fmt"

	"github.com/influx6/immute"
)

//Next provides a nice shorthand for function with one argument
type Next func(i interface{})

//Callable defines a function with a interface and function as arguments
type Callable func(i interface{}, f func(g interface{}))

//Callabut defines a function with one argument
type Callabut func(i interface{})

//CallList short type for list of functions
type CallList []Callable

//RollerInterface defines the core member functions for middleware style structures
type RollerInterface interface {
	Size() int
	RemoveCallback(int) bool
	RemoveDoneCallback(int) bool
	DecidedDoneOnce(r Callable)
	DecidedDone(r Callable)
	ReceiveDoneOnce(r Callabut)
	ReceiveDone(r Callabut)
	DecideOnce(r Callable)
	Decide(r Callable)
	Receive(r Callabut)
	ReceiveOnce(r Callabut)
	Munch(i interface{})
	RevMunch(i interface{})
	CallAt(f int, i interface{})
	CallOnceAt(f int, i interface{})
	CallDoneAt(f int, i interface{})
	CallOnceDoneAt(f int, i interface{})
	ReverseCallAt(f int, i interface{})
	ReverseOnceCallAt(f int, i interface{})
	ReverseCallDoneAt(f int, i interface{})
	ReverseOnceCallDoneAt(f int)
	onRoller(i interface{}, f func(g interface{}))
	onRevRoller(i interface{}, f func(g interface{}))
}

//Roller is the core structure meeting the RollerInterface
type Roller struct {
	enders     []Callable
	doners     []Callable
	onceDoners []Callable
	onceEnders []Callable
}

//Eventables interface defines the core event member methods
type Eventables interface {
	Emit(interface{})
	Listen(Callabut)
}

//StreamPackers interface defines the core streampack member methods
type StreamPackers interface {
	Flush()
	Send(interface{})
	Subscribe(bool, bool)
	WeakSubscribe(bool, bool)
}

//Streamable interface defines the core stream member functions
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

//Streams struct is the core structure for streams
type Streams struct {
	*Roller
	Buffer  *immute.Sequence
	Drains  *EventRoll
	manual  bool
	reverse bool
}

//EventRoll provides an event struct for event like purposes
type EventRoll struct {
	Handlers *immute.Sequence
	ID       string
	fired    bool
	cache    interface{}
}

//StreamPack Provides a persistent streams data
type StreamPack struct {
	Buffer *immute.Sequence
	Adds   *EventRoll
}

//Flush flushes the streampack data
func (s *StreamPack) Flush() {
	s.Buffer.Clear()
}

//Send emits a data into the streampack buffer
func (s *StreamPack) Send(data interface{}) {
	s.Adds.Emit(data)
}

//WeakSubscribe creates a new stream from the streapack but only streams the
//current streampack contents after which it disconnects from the streampack
func (s *StreamPack) WeakSubscribe(reverse bool, manual bool) *Streams {
	sm := NewStream(reverse, manual)

	s.Buffer.Each(func(data interface{}, key interface{}) interface{} {
		sm.Send(data)
		return nil
	}, func(_ int, _ interface{}) {})

	return sm
}

//Subscribe allows creation of a stream from the streampack which will still continue to
//be connected after all streampack data have been emitted to allow continues update
func (s *StreamPack) Subscribe(reverse bool, manual bool) *Streams {
	sm := NewStream(reverse, manual)
	s.Buffer.Each(func(data interface{}, key interface{}) interface{} {
		sm.Send(data)
		return nil
	}, func(_ int, _ interface{}) {})

	s.Adds.Listen(func(data interface{}) {
		sm.Send(data)
	})

	return sm
}

//Listen adds a callback waiting for the next firing of the event
func (e *EventRoll) Listen(f ...Callabut) {
	for _, v := range f {
		e.Handlers.Add(v, nil)
	}
}

//After lets you bind for when an event as fired or will be called after an event is fired
//after which the callback is added into the events callback list
func (e *EventRoll) After(f ...Callabut) {
	if e.fired {
		for _, v := range f {
			e.Handlers.Add(v, nil)
			v(e.cache)
		}
		return
	}

	for _, v := range f {
		e.Handlers.Add(v, nil)
	}
}

//Emit executes all callbacks with the val supplied
func (e *EventRoll) Emit(val interface{}) {
	if e.Handlers.Length() <= 0 {
		return
	}

	e.cache = val
	e.fired = true
	e.Handlers.Each(func(data interface{}, key interface{}) interface{} {
		fn, ok := data.(Callabut)

		if !ok {
			return nil
		}

		fn(val)
		return nil
	}, func(_ int, _ interface{}) {})
}

//Seq returns the buffer as a immute.Sequence for sequence operations without affecting the stream data
func (s *Streams) Seq() *immute.Sequence {
	return s.Buffer.Seq()
}

//Drain lets you add a callback into the drain event
func (s *Streams) Drain(drainer ...Callabut) {
	s.Drains.Listen(drainer...)
}

//Collect returns a list of the streams buffer content
func (s *Streams) Collect() []interface{} {
	data, ok := s.Buffer.Obj().([]interface{})

	if !ok {
		return nil
	}

	s.Buffer.Clear()
	return data
}

//CollectAndStream copys all streams buffer content into a list and streams that lists instead
func (s *Streams) CollectAndStream() {
	s.Send(s.Collect())
}

//CollectTo copys all stream data into a list and calls the supplied callback with it
func (s *Streams) CollectTo(fn func(data []interface{})) {
	var data = s.Collect()
	fn(data)
}

//Send handles sending data either into the stream's buffer when set to or when having pending elements
func (s *Streams) Send(data interface{}) {
	if s.Size() > 0 {
		if !s.manual {
			if s.BufferSize() > 0 {
				s.Buffer.Add(data, nil)
			} else {
				s.Delegate(data)
				s.Drains.Emit(data)
			}
		} else {
			s.Buffer.Add(data, nil)
		}
	} else {
		s.Buffer.Add(data, nil)
	}

	if !s.manual {
		s.Stream()
	}
}

//Delegate sends a data into the callbacks directly without including the stream buffer
func (s *Streams) Delegate(data interface{}) {
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

//Stream is the manual mechanic for forcing streaming operation
func (s *Streams) Stream() {

	listeners := s.Size()

	if listeners <= 0 {
		return
	}

	if s.BufferSize() <= 0 {
		return
	}

	if s.BufferSize() > 0 {
		cur := s.Buffer.Delete(0)

		if s.BufferSize() <= 0 {
			s.Delegate(cur)
			s.Drains.Emit(cur)
		} else {
			s.Delegate(cur)
		}
	}

}

//BufferSize returns the total length of items in the buffer
func (s *Streams) BufferSize() int {
	return s.Buffer.Length()
}

//Reverse sets the stream to call its callbacks in reverse
func (s *Streams) Reverse() {
	s.reverse = true
}

//Unreverse allows resetting the reverse callback execution property of a stream
func (s *Streams) Unreverse() {
	s.reverse = false
}

//Clear flushes the stream buffer
func (s *Streams) Clear() {
	s.Buffer.Clear()
}

//NotifyDrain fires off an event to the drain event roller
func (s *Streams) NotifyDrain(data interface{}) {
	s.Drains.Emit(data)
}

//onRoller provides a function mask for binding itself into another roller
func (r *Roller) onRoller(i interface{}, next func(g interface{})) {
	r.Munch(i)
	next(nil)
}

//onRevRoller provides a function mask for binding itself into another roller in reverse mode
func (r *Roller) onRevRoller(i interface{}, next func(g interface{})) {
	r.RevMunch(i)
	next(nil)
}

//Munch execs the callback and once callback list in FIFS(First in First Serve) order
func (r *Roller) Munch(i interface{}) {
	r.CallAt(0, i)
	r.CallOnceAt(0, i)
}

//RevMunch execs the callback and once callback list in reverse
func (r *Roller) RevMunch(i interface{}) {
	r.ReverseCallAt(0, i)
	r.ReverseOnceCallAt(0, i)
}

//ReverseCallDoneAt exec all callbacks in the done list in reverse
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

//ReverseOnceCallDoneAt exec all callbacks in the done once list in reverse
func (r *Roller) ReverseOnceCallDoneAt(i int, g interface{}) {
	if len(r.onceDoners) <= 0 {
		return
	}
	total := len(r.onceDoners) - 1
	loc := total - i

	if loc >= 0 {
		val := r.onceDoners[loc]
		if val != nil {
			val(g, func(fval interface{}) {

				ind := i + 1
				if fval == nil {
					r.ReverseOnceCallDoneAt(ind, g)
					return
				}

				r.ReverseOnceCallDoneAt(ind, fval)
			})
		}
	} else {
		r.onceDoners = make([]Callable, 0)
	}
}

//ReverseOnceCallAt exec all callbacks in the callback list in reverse
func (r *Roller) ReverseOnceCallAt(i int, g interface{}) {
	if len(r.onceEnders) <= 0 {
		return
	}
	total := len(r.onceEnders) - 1
	loc := total - i

	if loc >= 0 {
		val := r.onceEnders[loc]
		if val != nil {
			val(g, func(fval interface{}) {

				ind := i + 1

				if fval == nil {
					r.ReverseOnceCallAt(ind, g)
					return
				}

				r.ReverseOnceCallAt(ind, fval)
			})
		}
	} else {
		r.onceEnders = make([]Callable, 0)
		r.ReverseOnceCallDoneAt(0, g)
	}
}

//ReverseCallAt exec all callbacks in the callback list in reverse
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

//CallOnceAt exec all callbacks in the once callback list
func (r *Roller) CallOnceAt(i int, g interface{}) {
	if len(r.onceEnders) <= 0 {
		return
	}

	if len(r.onceEnders) > i {
		val := r.onceEnders[i]
		if val != nil {
			val(g, func(f interface{}) {

				ind := i + 1
				if f == nil {
					r.CallOnceAt(ind, g)
					return
				}

				r.CallOnceAt(ind, f)
			})
		}
	} else {
		r.onceEnders = make([]Callable, 0)
		r.CallOnceDoneAt(0, g)
	}
}

//CallAt exec all callbacks in the callback list
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

//CallOnceDoneAt exec all callbacks in the once done list
func (r *Roller) CallOnceDoneAt(i int, g interface{}) {
	if len(r.onceDoners) <= 0 {
		return
	}

	if len(r.onceDoners) > i {
		val := r.onceDoners[i]
		if val != nil {
			val(g, func(f interface{}) {

				ind := i + 1
				if f == nil {
					r.CallOnceDoneAt(ind, g)
					return
				}

				r.CallOnceDoneAt(ind, f)
			})
		}
	} else {
		r.onceDoners = make([]Callable, 0)
	}
}

//CallDoneAt exec all callbacks in the done list
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

//ReceiveDone adds a callback to the done list
func (r *Roller) ReceiveDone(f func(c interface{})) {
	r.DecidedDone(func(i interface{}, next func(t interface{})) {
		f(i)
		next(nil)
	})
}

//ReceiveDoneOnce adds a callback to the done once list
func (r *Roller) ReceiveDoneOnce(f func(c interface{})) {
	r.DecidedDoneOnce(func(i interface{}, next func(t interface{})) {
		f(i)
		next(nil)
	})
}

//DecidedDone adds a callback to the  done list
func (r *Roller) DecidedDone(f ...Callable) {
	r.doners = append(r.doners, f...)
}

//DecidedDoneOnce adds a callback to the once done list
func (r *Roller) DecidedDoneOnce(f ...Callable) {
	r.onceDoners = append(r.onceDoners, f...)
}

//Receive adds a standard style callback
func (r *Roller) Receive(f func(c interface{})) {
	r.Decide(func(i interface{}, next func(t interface{})) {
		f(i)
		next(nil)
	})
}

//ReceiveOnce adds a standard style callback to the once list
func (r *Roller) ReceiveOnce(f func(c interface{})) {
	r.DecideOnce(func(i interface{}, next func(t interface{})) {
		f(i)
		next(nil)
	})
}

//Decide adds a middleware style callback
func (r *Roller) Decide(f ...Callable) {
	r.enders = append(r.enders, f...)
}

//DecideOnce adds a middleware style callback to the once list
func (r *Roller) DecideOnce(f ...Callable) {
	r.onceEnders = append(r.onceEnders, f...)
}

//RemoveCallback removes a callback at position index of the callbacks
func (r *Roller) RemoveCallback(ind int) bool {
	if r.Size() <= ind || ind <= -1 {
		return false
	}

	r.enders = append(r.enders[:ind], r.enders[ind+1:]...)
	return true
}

//RemoveDoneCallback removes a callback at position index of the done callbacks
func (r *Roller) RemoveDoneCallback(ind int) bool {
	if len(r.doners) <= ind || ind <= -1 {
		return false
	}

	r.doners = append(r.doners[:ind], r.doners[ind+1:]...)
	return true
}

//RemoveOnceCallback removes a callback at position index of the once callbacks
func (r *Roller) RemoveOnceCallback(ind int) bool {
	if len(r.onceEnders) <= ind || ind <= -1 {
		return false
	}

	r.onceEnders = append(r.onceEnders[:ind], r.onceEnders[ind+1:]...)
	return true
}

//RemoveOnceDoneCallback removes a callback at position index of the done once callbacks
func (r *Roller) RemoveOnceDoneCallback(ind int) bool {
	if len(r.onceDoners) <= ind || ind <= -1 {
		return false
	}

	r.onceDoners = append(r.onceDoners[:ind], r.onceDoners[ind+1:]...)
	return true
}

//Size returns the total number of callbacks
func (r *Roller) Size() int {
	return len(r.enders)
}

//String returns a string representation of the callback list
func (r *Roller) String() string {
	return fmt.Sprint("Total Roller Callback:", len(r.enders))
}

//NewRoller creates and return a pointer to a new roller struct ready for middleware style pattern stack
func NewRoller() *Roller {
	roll := &Roller{
		[]Callable{},
		[]Callable{},
		[]Callable{},
		[]Callable{},
	}
	return roll
}

//NewEvent creates a new eventroll for event notification to its callbacks
func NewEvent(id string) *EventRoll {
	list := immute.CreateList(make([]interface{}, 0))
	return &EventRoll{list, id, false, nil}
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
	sm := &StreamPack{list, data}

	data.Listen(func(i interface{}) {
		list.Add(i, nil)
	})

	return sm
}
