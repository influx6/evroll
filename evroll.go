package evroll

import (
	"fmt"
)

type Next func(i interface{})
type Callable func(i interface{}, f func(g interface{}))
type CallList []Callable

type RollerInterface interface {
	End(r Callable)
	Munch(i interface{})
	RevMunch(i interface{})
	CallAt(f int)
	ReverseCallAt(f int)
	onRoller(i interface{}, f func(g interface{}))
	onRevRoller(i interface{}, f func(g interface{}))
}

type Roller struct {
	enders []Callable
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
	}
}

func (r *Roller) Or(f func(i interface{})) {
	r.End(func(i interface{}, next func(t interface{})) {
		f(i)
		next(nil)
	})
}

func (r *Roller) End(f ...Callable) {
	r.enders = append(r.enders, f...)
}

func (r *Roller) size() int {
	return len(r.enders)
}

func (r *Roller) String() string {
	return fmt.Sprint(r.enders)
}

func NewRoller() *Roller {
	return &Roller{[]Callable{}}
}
