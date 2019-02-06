package domain

import (
	"reflect"

	"github.com/ktye/iv/apl"
)

// ToNumber accepts scalars and single size arrays.
// and converts them to scalars if they contain one of the types:
// apl.Bool, apl.Int, apl.Float or apl.Complex.
func ToNumber(child SingleDomain) SingleDomain {
	return number{child, true}
}

// IsNumber accepts scalars if they contain of of the types:
// apl.Bool, apl.Int, apl.Float or apl.Complex
func IsNumber(child SingleDomain) SingleDomain {
	return number{child, false}
}

type number struct {
	child   SingleDomain
	convert bool
}

func (n number) To(a *apl.Apl, V apl.Value) (apl.Value, bool) {
	v := V
	if ar, ok := V.(apl.Array); ok {
		if n.convert == false {
			return V, false
		}
		if n := apl.ArraySize(ar); n != 1 {
			return V, false
		}
		v = ar.At(0)
	}
	if b, ok := V.(apl.Bool); ok {
		return a.Tower.FromBool(b), true
	}
	if i, ok := V.(apl.Index); ok {
		return a.Tower.FromIndex(int(i)), true
	}
	if _, ok := a.Tower.Numbers[reflect.TypeOf(v)]; ok {
		return v, true
	}
	return V, false
}
func (n number) String(a *apl.Apl) string {
	name := "number"
	if n.convert {
		name = "tonumber"
	}
	if n.child == nil {
		return name
	}
	return name + " " + n.child.String(a)
}

// ToIndex converts the scalar to an Index.
func ToIndex(child SingleDomain) SingleDomain {
	return index{child}
}

type index struct {
	child SingleDomain
}

func (idx index) To(a *apl.Apl, V apl.Value) (apl.Value, bool) {
	if b, ok := V.(apl.Bool); ok {
		if b == true {
			return apl.Index(1), true
		}
		return apl.Index(0), true
	}
	if n, ok := V.(apl.Index); ok {
		return n, true
	}
	if n, ok := V.(apl.Number); ok == false {
		return V, false
	} else {
		if i, ok := n.ToIndex(); ok == false {
			return V, false
		} else {
			return propagate(a, apl.Index(i), idx.child)
		}
	}
}
func (idx index) String(a *apl.Apl) string {
	if idx.child == nil {
		return "index"
	} else {
		return "index " + idx.child.String(a)
	}
}
