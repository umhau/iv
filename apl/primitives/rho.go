package primitives

import (
	"fmt"

	"github.com/ktye/iv/apl"
	. "github.com/ktye/iv/apl/domain"
)

func init() {
	register(primitive{
		symbol: "⍴",
		doc:    "shape",
		Domain: Monadic(nil),
		fn:     rho1,
	})
	register(primitive{
		symbol: "⍴",
		doc:    "reshape",
		Domain: Dyadic(Split(ToVector(ToIndexArray(nil)), ToArray(nil))),
		fn:     rho2,
		sel:    selection(rho2),
	})
	register(primitive{
		symbol: "⍴",
		doc:    "reshape channel",
		Domain: Dyadic(Split(ToVector(ToIndexArray(nil)), IsChannel(nil))),
		fn:     rhoChannel,
	})
}

// Rho1 returns the shape of R.
func rho1(a *apl.Apl, _, R apl.Value) (apl.Value, error) {
	// Report a table as a two dimensional array.
	if t, ok := R.(apl.Table); ok == true {
		return apl.IntArray{
			Dims: []int{2},
			Ints: []int{t.Rows, len(t.K)},
		}, nil
	}
	// An object returns the number of keys.
	if o, ok := R.(apl.Object); ok == true {
		n := len(o.Keys())
		return apl.IntArray{Dims: []int{1}, Ints: []int{n}}, nil
	}

	if _, ok := R.(apl.Array); ok == false {
		return apl.EmptyArray{}, nil
	}
	// Shape of an empty array is 0, rank is 1
	if _, ok := R.(apl.EmptyArray); ok {
		return apl.IntArray{Ints: []int{0}, Dims: []int{1}}, nil
	}
	ar := R.(apl.Array)
	shape := ar.Shape()
	ret := apl.IntArray{
		Ints: make([]int, len(shape)),
		Dims: []int{len(shape)},
	}
	for i, n := range shape {
		ret.Ints[i] = n
	}
	return ret, nil
}

// Rho2 is dyadic reshape, L is empty or index array, R is array.
func rho2(a *apl.Apl, L, R apl.Value) (apl.Value, error) {
	// L is empty, returns empty.
	if apl.ArraySize(L.(apl.Array)) == 0 {
		return apl.EmptyArray{}, nil
	}

	if _, ok := R.(apl.Object); ok {
		return nil, fmt.Errorf("cannot reshape %T", R)
	}

	l := L.(apl.IntArray)
	shape := make([]int, len(l.Ints))
	copy(shape, l.Ints)

	if rs, ok := R.(apl.Reshaper); ok {
		return rs.Reshape(shape), nil
	}
	return nil, fmt.Errorf("cannot reshape %T", R)
}

// rhoChannel returns a channel and sends arrays with the shape of L.
// Values are read from a channel R.
func rhoChannel(a *apl.Apl, L, R apl.Value) (apl.Value, error) {
	if L.(apl.Array).Size() == 0 {
		return R, nil
	}
	al := L.(apl.IntArray)
	size := prod(al.Ints)
	newarray := func() apl.MixedArray {
		s := make([]int, len(al.Ints))
		copy(s, al.Ints)
		return apl.MixedArray{
			Dims:   s,
			Values: make([]apl.Value, size),
		}
	}
	res := newarray()

	in := R.(apl.Channel)
	out := apl.NewChannel()
	go func() {
		p := 0
		defer close(out[0])
		push := func(v apl.Value) {
			res.Values[p] = v
			p++
			if p == size {
				select {
				case _, ok := <-out[1]:
					if ok == false {
						close(in[1])
						return
					}
				case out[0] <- res:
					res = newarray()
					p = 0
				}
			}
		}
		for {
			select {
			case _, ok := <-out[1]:
				if ok == false {
					close(in[1])
					return
				}
			case v, ok := <-in[0]:
				if ok == false {
					return
				}
				if ar, ok := v.(apl.Array); ok {
					for i := 0; i < ar.Size(); i++ {
						push(ar.At(i))
					}
				} else {
					push(v)
				}
			}
		}
	}()
	return out, nil
}

func prod(shape []int) int {
	if len(shape) == 0 {
		return 0
	}
	n := shape[0]
	for i := 1; i < len(shape); i++ {
		n *= shape[i]
	}
	return n
}
