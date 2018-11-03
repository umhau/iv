package apl

import (
	"fmt"
	"unicode"

	"github.com/ktye/iv/apl/scan"
)

// Identifier is a the Value of a variable identifier.
// It evaluates to itself, not the stored Value.
type Identifier string

func (id Identifier) String(a *Apl) string {
	return string(id)
}

func (id Identifier) Eval(a *Apl) (Value, error) {
	return id, nil
}

// Assign assigns a value to a variable with the given name.
func (a *Apl) Assign(name string, v Value) error {
	ok, isfunc := isVarname(name)
	if ok == false {
		return fmt.Errorf("variable name is not allowed: %s", name)
	}

	// Assignment to the special variable ⎕ prints the value.
	if name == "⎕" {
		fmt.Fprintf(a.stdout, "%s\n", v.String(a))
		return nil
	}

	if _, ok := v.(Function); ok && isfunc != true {
		return fmt.Errorf("cannot assign a function to an uppercase variable")
	} else if ok == false && isfunc == true {
		return fmt.Errorf("only functions can be assigned to lowercase variables")
	}

	a.vars[name] = v
	return nil
}

// Lookup returns the value stored under the given variable name.
// It returns nil, if the variable does not exist.
func (a *Apl) Lookup(name string) Value {
	v, ok := a.vars[name]
	if ok == false {
		return nil
	}
	return v
}

// NumVar contains the identifier to a value.
// The name is upper case and does not evaluate to a function.
// NumVar evaluates to the stored value or to an Identifier if it is undeclared.
type numVar struct {
	name string
}

func (v numVar) String(a *Apl) string {
	return v.name
}

func (v numVar) Eval(a *Apl) (Value, error) {
	x := a.Lookup(v.name)
	if x == nil {
		return Identifier(v.name), nil
	}
	return x, nil
}

// FnValue contains the identifier to a function value.
// It's name is lowercase.
// FnVar evaluates to the stored value or to an Identifier if it is undeclared.
type fnVar string

func (f fnVar) String(a *Apl) string {
	return string(f)
}

func (f fnVar) Eval(a *Apl) (Value, error) {
	return f, nil
}

func (f fnVar) Call(a *Apl, l, r Value) (Value, error) {
	x := a.Lookup(string(f))
	if x == nil {
		fmt.Println(f, "is not assigned")
		return Identifier(f), nil
	}
	fn, ok := x.(Function)
	if ok == false {
		return nil, fmt.Errorf("value in function variable is not a function: %T", x)
	}
	if fn == nil {
		return nil, fmt.Errorf("value in function variable %s is nil", string(f))
	}
	return fn.Call(a, l, r)
}

// isVarname returns if the string is allowed as a variable name and
// referes to a number or function value.
func isVarname(s string) (ok, isfunc bool) {
	if s == "" {
		return false, false
	}
	special := false
	upper := false
	for i, r := range s {
		if scan.AllowedInVarname(r, i == 0) == false {
			return false, false
		}
		if i == 0 && scan.IsSpecial(r) {
			special = true
		}
		if i == 0 && unicode.IsUpper(r) {
			upper = true
		}
	}
	if special {
		return true, false
	}
	return true, upper == false
}
