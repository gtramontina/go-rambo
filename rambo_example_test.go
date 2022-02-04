package rambo_test

import (
	"fmt"
	"github.com/gtramontina/rambo"
	"os"
)

func ExampleLoad() {
	tmpTestPath := os.TempDir() + "/todo"
	defer func() {
		if err := os.RemoveAll(tmpTestPath); err != nil {
			panic(err)
		}
	}()

	storageFilePath := tmpTestPath + "/example.journal"
	app, err := rambo.Load(storageFilePath, &Calc{}, &Add{}, &Sub{}, &Mul{}, &Div{})
	if err != nil {
		panic(err)
	}

	if err := app.Transact(&Add{Amount: 10}); err != nil {
		panic(err)
	}
	if err := app.Transact(&Sub{Amount: 2.5}); err != nil {
		panic(err)
	}
	if err := app.Transact(&Mul{Amount: 4.25}); err != nil {
		panic(err)
	}
	if err := app.Transact(&Div{Amount: 1.25}); err != nil {
		panic(err)
	}

	total := app.Query(&Total{}).(float64)
	fmt.Printf("Total: %g\n", total)
	// Output:
	// Total: 25.5
}
