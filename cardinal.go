// Package cardinal is a Promise library written in Go using the reflect package.
//
// Before getting started there a few basic rules you need to understand:
//
// 	1. To reject a promise return a non-nil error
//
// 	2. Nil errors will not be piped into the next chained function
package cardinal

import (
	"fmt"
	"reflect"
)

// PromiseStruct holds the data for running the Promise
type PromiseStruct struct {
	Status stat            // Describes the success of the previously ran promise
	Result []reflect.Value // Holds the function output arguments of the previously ran promise
	Order  int             // For promise methods that return multiple promises, this is used to keep their return arguments in order
}

// Define Promise statuses
type stat int

const (
	fulfilled stat = 0
	pending   stat = 10
	rejected  stat = 20
)

// Initialize Promise Object
var promise = PromiseStruct{Status: pending}

// Function for running any function described as an interface{}
// All promise methods must call this function
func (p PromiseStruct) runFunc(fn interface{}) PromiseStruct {
	// Get pointer to passed in function
	fnPtr := reflect.ValueOf(fn)
	fnTyp := reflect.TypeOf(fn)

	// fn must be a function
	if fnPtr.Kind() != reflect.Func {
		message := fmt.Errorf("was expecting a %s but got %s", reflect.Func, fnPtr.Kind())
		return PromiseStruct{
			Status: rejected,
			Result: []reflect.Value{reflect.ValueOf(message)},
		}
	}

	// Error interface definition
	reflectError := reflect.TypeOf((*error)(nil)).Elem()

	// All arguments must be of the same type
	for i := 0; i < fnTyp.NumIn(); i++ {
		// Although they are naturally coerced, reflect errors and errors passed in wont be the exact same type so they need to be excluded from this check
		if fnTyp.In(i) != p.Result[i].Type() && !fnTyp.In(i).Implements(reflectError) && !p.Result[i].Type().Implements(reflectError) {
			message := fmt.Errorf("Args should be %s but got %s", fnTyp.In(i), p.Result[i].Type())
			return PromiseStruct{
				Status: rejected,
				Result: []reflect.Value{reflect.ValueOf(message)},
			}
		}
	}

	// Call function
	y := fnPtr.Call(p.Result)

	// Create return structure
	resolveStatus := fulfilled
	var values []reflect.Value

	// Add each return value to the reflect Value slice or return just the error
	for i := 0; i < fnTyp.NumOut(); i++ {
		if fnTyp.Out(i).Implements(reflectError) && !y[i].IsNil() {
			values = []reflect.Value{y[i]}
			resolveStatus = rejected
			break
		} else if !fnTyp.Out(i).Implements(reflectError) {
			values = append(values, y[i])
		}
	}

	return PromiseStruct{
		Status: resolveStatus,
		Result: values,
		Order:  p.Order,
	}
}
