package cardinal

import (
	"errors"
	"fmt"
	"time"
)

// should be able to chain then functions piping the values from one to the other
func ExamplePromise() {
	stuff := Promise(func() int {
		return 3
	}).
		Then(func(x int) int {
			fmt.Println(x)
			return x * x
		}).
		Then(func(x int) {
			fmt.Println(x)
		})
	fmt.Println(stuff.Result)
	// Output:
	// 3
	// 9
	// []
}

// should not skip Then and not run Catch if error returned is nil
func ExamplePromiseStruct_Then() {
	stuff := Promise(func() int {
		return 3
	}).
		Then(func(x int) (int, error) {
			fmt.Println(x)
			return x * x, nil
		}).
		Then(func(x int) (int, error) {
			fmt.Println(x)
			return x * x, nil
		})
	fmt.Println(stuff.Result[0])
	// Output:
	// 3
	// 9
	// 81
}

// should  skip Then and  run Catch if error returned is not nil
func ExamplePromiseStruct_Catch() {
	stuff := Promise(func() int {
		return 3
	}).
		Then(func(x int) (int, error) {
			fmt.Println(x)
			return x * x, errors.New("should show up in Catch")
		}).
		Then(func(x int) (int, error) {
			fmt.Println(x)
			return x * x, nil
		}).
		Catch(func(e error) error {
			fmt.Println(e)
			return e
		})
	fmt.Println(stuff.Result[0])
	// Output:
	// 3
	// should show up in Catch
	// should show up in Catch
}

// should type check and throw a sensible error
func ExamplePromiseStruct() {
	stuff := Promise(func() int {
		return 3
	}).
		Then(func(x int) (int, error) {
			fmt.Println(x)
			return x * x, nil
		}).
		Then(func(x float64) (float64, error) {
			fmt.Println(x)
			return x * x, nil
		}).
		Catch(func(e error) error {
			fmt.Println(e)
			return e
		})
	fmt.Println(stuff.Result[0])
	// Output:
	// 3
	// Args should be float64 but got int
	// Args should be float64 but got int
}

// should type check and throw a sensible error
func ExamplePromiseStruct_All() {
	stuff := Promise(func() int {
		return 3
	}).
		All(
			func(x int) (int, error) {
				fmt.Println(x)
				time.Sleep(50 * time.Millisecond) //Guarantee Order
				return x + 3, nil
			},
			func(x int) (int, error) {
				fmt.Println(x)
				return x * x, nil
			})
	fmt.Println(stuff.Result[0])
	fmt.Println(stuff.Result[1])
	// Output:
	// 3
	// 3
	// 6
	// 9
}

// should properly pass the result of All to a Then
func ExamplePromiseStruct_All_then() {
	stuff := Promise(func() int {
		return 3
	}).
		All(
			func(x int) (int, error) {
				fmt.Println(x)
				return x * x, nil
			},
			func(x int) (int, error) {
				fmt.Println(x)
				return x * x, nil
			}).
		Then(func(x int, y int) int {
			fmt.Println(x, y)
			return x * y
		})
	fmt.Println(stuff.Result[0])
	// Output:
	// 3
	// 3
	// 9 9
	// 81
}

// Should map a given array with a given function
func ExamplePromiseStruct_Map() {
	fruits := []string{"apples", "bananas", "oranges", "cherries"}
	stuff := Promise(func() {}).
		Map(fruits, func(s string) string {
			return s + " are not a fruit"
		}).
		Then(func(a string, b string, c string, d string) string {
			fmt.Println(a)
			return b + " and " + d
		})
	fmt.Println(stuff.Result[0])
	// Output:
	// apples are not a fruit
	// bananas are not a fruit and cherries are not a fruit
}

// should map a piped array with a given function
func ExamplePromiseStruct_ThenMap() {
	fruits := []string{"apples", "bananas", "oranges", "cherries"}
	stuff := Promise(func() []string { return fruits }).
		Map(fruits, func(s string) string {
			return s + " are not a fruit"
		}).
		Then(func(a string, b string, c string, d string) string {
			fmt.Println(a)
			return b + " and " + d
		})
	fmt.Println(stuff.Result[0])
	// Output:
	// apples are not a fruit
	// bananas are not a fruit and cherries are not a fruit
}

//should reduce a given array with a given function
func ExamplePromiseStruct_Reduce() {
	fruits := []string{"apples", "bananas", "oranges", "cherries"}
	stuff := Promise(func() {}).
		Reduce(fruits, func(a string, b string, i int, l int) string {
			return a + " " + b
		}, nil).
		Then(func(s string) string {
			fmt.Println(s)
			return s
		})
	fmt.Println(stuff.Result[0])
	// Output:
	// apples bananas oranges cherries
	// apples bananas oranges cherries
}

func ExamplePromiseStruct_ThenReduce() {
	fruits := []string{"apples", "bananas", "oranges", "cherries"}
	stuff := Promise(func() []string { return fruits }).
		ThenReduce(func(a string, b string, i int, l int) string {
			return a + " " + b
		}, "grapes")
	fmt.Println(stuff.Result[0])
	// Output:
	// grapes apples bananas oranges cherries
}
