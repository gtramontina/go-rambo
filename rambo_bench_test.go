package rambo_test

import (
	"github.com/gtramontina/rambo"
	"os"
	"testing"
)

func BenchmarkRambo(b *testing.B) {
	b.Run("transaction", func(b *testing.B) {
		b.Run("serial", func(b *testing.B) { benchTx(b, false) })
		b.Run("parallel", func(b *testing.B) { benchTx(b, true) })
	})

	b.Run("bootstrap", func(b *testing.B) {
		b.Run("1 transaction", func(b *testing.B) { benchBootstrap(b, 1, false) })
		b.Run("10 transactions", func(b *testing.B) { benchBootstrap(b, 10, false) })
		b.Run("100 transactions", func(b *testing.B) { benchBootstrap(b, 100, false) })
		b.Run("1000 transactions", func(b *testing.B) { benchBootstrap(b, 1000, false) })
		b.Run("10000 transactions", func(b *testing.B) { benchBootstrap(b, 10000, false) })
		b.Run("100000 transactions", func(b *testing.B) { benchBootstrap(b, 100000, false) })
	})

	b.Run("bootstrap truncated", func(b *testing.B) {
		b.Run("1 transaction", func(b *testing.B) { benchBootstrap(b, 1, true) })
		b.Run("10 transactions", func(b *testing.B) { benchBootstrap(b, 10, true) })
		b.Run("100 transactions", func(b *testing.B) { benchBootstrap(b, 100, true) })
		b.Run("1000 transactions", func(b *testing.B) { benchBootstrap(b, 1000, true) })
		b.Run("10000 transactions", func(b *testing.B) { benchBootstrap(b, 10000, true) })
		b.Run("100000 transactions", func(b *testing.B) { benchBootstrap(b, 100000, true) })
	})
}

func benchTx(b *testing.B, parallel bool) {
	storageFilePath := os.TempDir() + "/todo/" + b.Name()
	defer func() {
		if err := os.RemoveAll(storageFilePath); err != nil {
			b.Fatal(err)
		}
	}()

	app, err := rambo.Load(storageFilePath, &Calc{}, &Add{})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	if parallel {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := app.Transact(&Add{Amount: 1}); err != nil {
					b.Fatal(err)
				}
			}
		})
	} else {
		for i := 0; i < b.N; i++ {
			if err := app.Transact(&Add{Amount: 1}); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func benchBootstrap(b *testing.B, numberOfTx int, truncated bool) {
	storageFilePath := os.TempDir() + "/todo/" + b.Name()
	defer func() {
		if err := os.RemoveAll(storageFilePath); err != nil {
			b.Fatal(err)
		}
	}()

	app, err := rambo.Load(storageFilePath, &Calc{}, &Add{})
	if err != nil {
		b.Fatal(err)
	}

	for txNumber := 0; txNumber < numberOfTx; txNumber++ {
		if err := app.Transact(&Add{Amount: 1}); err != nil {
			b.Fatal(err)
		}
	}

	if truncated {
		if _, err := rambo.Load(storageFilePath, &Calc{}, &Add{}); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := rambo.Load(storageFilePath, &Calc{}, &Add{}); err != nil {
			b.Fatal(err)
		}
	}
}
