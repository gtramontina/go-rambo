package rambo_test

import (
	"github.com/gtramontina/rambo"
	"os"
	"sync"
	"testing"
)

func TestRambo(t *testing.T) {
	tmpTestPath := os.TempDir() + "/todo"
	defer func() { assertNoError(t, os.RemoveAll(tmpTestPath)) }()

	t.Run("executes transactions and queries", func(t *testing.T) {
		storageFilePath := tmpTestPath + "/" + t.Name()
		app, err := rambo.Load(storageFilePath, &Calc{}, &Add{}, &Sub{}, &Mul{}, &Div{})
		assertNoError(t, err)

		assertNoError(t, app.Transact(&Add{Amount: 2}))
		assertEq(t, app.Query(&Total{}).(float64), 2)

		assertNoError(t, app.Transact(&Mul{Amount: 5}))
		assertEq(t, app.QueryFn(func(calc *Calc) any { return calc.Total }).(float64), 10)
	})

	t.Run("persists state", func(t *testing.T) {
		storageFilePath := tmpTestPath + "/" + t.Name()
		app, err := rambo.Load(storageFilePath, &Calc{}, &Add{}, &Sub{}, &Mul{}, &Div{})
		assertNoError(t, err)

		assertNoError(t, app.Transact(&Add{Amount: 1.23}))

		// Reload…
		app, err = rambo.Load(storageFilePath, &Calc{}, &Add{}, &Sub{}, &Mul{}, &Div{})
		assertNoError(t, err)
		assertEq(t, app.Query(&Total{}).(float64), 1.23)
	})

	t.Run("multiple transaction types", func(t *testing.T) {
		storageFilePath := tmpTestPath + "/" + t.Name()
		app, err := rambo.Load(storageFilePath, &Calc{}, &Add{}, &Sub{}, &Mul{}, &Div{})
		assertNoError(t, err)

		assertNoError(t, app.Transact(&Add{Amount: 10}))
		assertNoError(t, app.Transact(&Sub{Amount: 5}))
		assertNoError(t, app.Transact(&Mul{Amount: 4}))
		assertNoError(t, app.Transact(&Div{Amount: 8}))

		// Reload…
		app, err = rambo.Load(storageFilePath, &Calc{}, &Add{}, &Sub{}, &Mul{}, &Div{})
		assertNoError(t, err)
		assertEq(t, app.Query(&Total{}).(float64), 2.5)
	})

	t.Run("concurrent transactions", func(t *testing.T) {
		storageFilePath := tmpTestPath + "/" + t.Name()
		app, err := rambo.Load(storageFilePath, &Calc{}, &Add{}, &Sub{}, &Mul{}, &Div{})
		assertNoError(t, err)

		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			for i := 0; i < 1000; i++ {
				assertNoError(t, app.Transact(&Add{Amount: 3}))
			}
			wg.Done()
		}()

		go func() {
			for i := 0; i < 1000; i++ {
				assertNoError(t, app.Transact(&Sub{Amount: 1}))
			}
			wg.Done()
		}()

		wg.Wait()
		assertEq(t, app.Query(&Total{}).(float64), 2000)
	})
}
