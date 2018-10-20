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

// Promise is a generator for the PromiseStruct and it only accepts functions with no input arguments
func Promise(fn interface{}) PromiseStruct {
	if reflect.TypeOf(fn).NumIn() > 0 {
		fErr := fmt.Errorf("function argument should have no arguments")
		return PromiseStruct{
			Status: rejected,
			Result: []reflect.Value{reflect.ValueOf(fErr)},
		}
	}
	p := promise
	return p.runFunc(fn)
}

// Then runs a given function with the piped in parameters as long as the previous promise was successful
func (p PromiseStruct) Then(fn interface{}) PromiseStruct {
	// Rejected promises should skip chained thens
	if p.Status == rejected {
		return p
	}

	return p.runFunc(fn)
}

// Catch runs a given function if any promise before it failed to complete successfully
func (p PromiseStruct) Catch(fn interface{}) PromiseStruct {
	// fulfilled promises should skip chained catches
	if p.Status == fulfilled {
		return p
	}
	return p.runFunc(fn)
}

// All takes several functions and runs them in parallel
// The next chaining function will need to take the input of all function returns in order
func (p PromiseStruct) All(fn ...interface{}) PromiseStruct {
	// Rejected promises should skip chained thens
	if p.Status == rejected {
		return p
	}
	done := make(chan PromiseStruct)
	defer close(done)
	i := 0

	for _, f := range fn {
		go func(fptr interface{}, ind int) {
			done <- PromiseStruct{p.Status, p.Result, ind}.runFunc(fptr)
		}(f, i)
		i++
	}

	allStatuses := fulfilled
	results := make([][]reflect.Value, i, i)

	for j := 0; j < i; j++ {
		resolution := <-done
		if resolution.Status == rejected && allStatuses == rejected {
			results[resolution.Order] = resolution.Result
		} else if resolution.Status == rejected {
			allStatuses = rejected
			results[resolution.Order] = resolution.Result
		} else if resolution.Status != rejected && allStatuses == fulfilled {
			results[resolution.Order] = resolution.Result
		}
	}

	flattendResults := make([]reflect.Value, 0)
	for _, res := range results {
		flattendResults = append(flattendResults, res...)
	}

	return PromiseStruct{
		Status: allStatuses,
		Result: flattendResults,
	}
}

// Map takes an array and applies a given function to each element in the array
// The return of Map is similar to All, the next chained method will need to handle all passed returns in order
func (p PromiseStruct) Map(s interface{}, fn interface{}) PromiseStruct {
	// Rejected promises should skip chained thens
	if p.Status == rejected {
		return p
	}

	// Make sure a slice is received and that it is of the correct type
	xs := reflect.ValueOf(s)
	xstype := xs.Type()
	if xstype.Kind() != reflect.Slice {
		sErr := fmt.Errorf("first argument should be %s but got %s", reflect.Slice, xstype.Kind())
		return PromiseStruct{
			Status: rejected,
			Result: []reflect.Value{reflect.ValueOf(sErr)},
		}
	}

	done := make(chan PromiseStruct)
	defer close(done)
	i := 0
	for i = 0; i < xs.Len(); i++ {
		arg := []reflect.Value{xs.Index(i)}
		go func(fptr interface{}, result []reflect.Value, ind int) {
			done <- PromiseStruct{p.Status, result, ind}.runFunc(fptr)
		}(fn, arg, i)
	}

	allStatuses := fulfilled
	results := make([][]reflect.Value, i, i)

	for j := 0; j < i; j++ {
		resolution := <-done
		if resolution.Status == rejected && allStatuses == rejected {
			results[resolution.Order] = resolution.Result
		} else if resolution.Status == rejected {
			allStatuses = rejected
			results[resolution.Order] = resolution.Result
		} else if resolution.Status != rejected && allStatuses == fulfilled {
			results[resolution.Order] = resolution.Result
		}
	}

	flattendResults := make([]reflect.Value, 0)
	for _, res := range results {
		flattendResults = append(flattendResults, res...)
	}

	return PromiseStruct{
		Status: allStatuses,
		Result: flattendResults,
	}
}

// ThenMap is like Map but takes the array from the previously executed promise
func (p PromiseStruct) ThenMap(fn interface{}) PromiseStruct {
	// Rejected promises should skip chained thens
	if p.Status == rejected {
		return p
	}

	mapSlice := p.Result[0].Interface()
	return p.Map(mapSlice, fn)
}

// Reduce takes a slice and give a function and initial value creates a single value
func (p PromiseStruct) Reduce(s interface{}, fn interface{}, init interface{}) PromiseStruct {
	// Rejected promises should skip chained thens
	if p.Status == rejected {
		return p
	}

	// Make sure a slice is received and that it is of the correct type
	xs := reflect.ValueOf(s)
	xstype := xs.Type()
	if xstype.Kind() != reflect.Slice {
		sErr := fmt.Errorf("first argument should be %s but got %s", reflect.Slice, xstype.Kind())
		return PromiseStruct{
			Status: rejected,
			Result: []reflect.Value{reflect.ValueOf(sErr)},
		}
	}
	fmt.Println()
	// Make sure init argument is of the same type of the slice elements
	xsElem := reflect.TypeOf(s).Elem()
	if reflect.ValueOf(init).IsValid() && xsElem != reflect.TypeOf(init) {
		initErr := fmt.Errorf("Initializing argument should be %s but got %s", xsElem, reflect.TypeOf(init))
		return PromiseStruct{
			Status: rejected,
			Result: []reflect.Value{reflect.ValueOf(initErr)},
		}
	}

	// Setup reduce with initial argumets
	// if init is nil use the first two values of s, index 1 and slice length
	// if not use init and the first value of s, index 0, slice length
	ind := 0
	args := make([]reflect.Value, 0, 0)
	if !reflect.ValueOf(init).IsValid() {
		args = append(args, xs.Index(0), xs.Index(1), reflect.ValueOf(1), reflect.ValueOf(xs.Len()))
		ind = 1
	} else {
		args = append(args, reflect.ValueOf(init), xs.Index(0), reflect.ValueOf(0), reflect.ValueOf(xs.Len()))
	}

	for i := ind; i < xs.Len(); i++ {
		if i != ind {
			args[1] = xs.Index(i)
			args[2] = reflect.ValueOf(i)
		}

		accumulator := func(fptr interface{}, result []reflect.Value) PromiseStruct {
			return PromiseStruct{Status: p.Status, Result: result}.runFunc(fptr)
		}(fn, args)

		args[0] = accumulator.Result[0]

		if accumulator.Status == rejected {
			return accumulator
		}
	}

	return PromiseStruct{
		Status: fulfilled,
		Result: []reflect.Value{args[0]},
	}
}

//ThenReduce is like Reduce but the slice comes from a previous promise
func (p PromiseStruct) ThenReduce(fn interface{}, init interface{}) PromiseStruct {
	// Rejected promises should skip chained thens
	if p.Status == rejected {
		return p
	}

	redSlice := p.Result[0].Interface()
	return p.Reduce(redSlice, fn, init)
}
