# Rambo

<sup><b>⚠️ Warning:</b> <em>This is still an experiment.</em></sup>

Rambo is a Go implementation of the [Prevalent System](https://en.wikipedia.org/wiki/System_prevalence) design pattern, in which business objects are kept live in memory and transactions are journaled for system recovery. A prevalent system needs enough memory to hold its entire state in RAM (the "prevalent hypothesis").

This version is heavily inspired on [Prevayler's](https://prevayler.org/) author [Klaus Wuestefeld's](https://twitter.com/klauswuestefeld) "[PrevaylerJr](https://bit.ly/3tGxw3P)" version.

The name "Rambo" comes from the book "[Java RAMBO Manifesto: RAM-based Objects, Prevayler, and Raw Speed](https://amzn.to/3IbPsaC)", by [Peter Wayner](https://twitter.com/peterwayner).

## Requirements

As mentioned above, as the prevalent hypothesis assumes your entire system state fits in memory, this is the primary requirement: you must have enough RAM to hold your entire system state at once.

Given commands are journaled, they will be replayed in order to restore the system state. For this to work, the commands' execution on the system must be deterministic. If you rely on current date/time, for example, don't query the operating system while executing the command. Instead, pass it down as a member of the command so that the current date/time gets journaled with it.

Lastly, your system state and commands must be fully serializable. Rambo uses Go's [`gob`](https://pkg.go.dev/encoding/gob) package to encode/decode structs, so either all of your system's and commands' members are exported, or you implement both [`GobEncoder`](https://pkg.go.dev/encoding/gob#GobEncoder) and [`GobDecoder`](https://pkg.go.dev/encoding/gob#GobDecoder) interfaces accordingly.

## Limitations

Rambo does not implement any kind of replication, so systems using Rambo cannot be load-balanced.

Also, pay special attention to the usage of pointers. As mentioned above, Rambo uses Go's [`gob`](https://pkg.go.dev/encoding/gob) package to encode/decode structs, and it does not handle pointers as, perhaps, one would expect. From [`gob`'s documentation](https://pkg.go.dev/encoding/gob):

> (…) Pointers are not transmitted, but the things they point to are transmitted; that is, the values are flattened.

Here's a short example of how you can get tripped by this: <https://go.dev/play/p/JVsA5ke6Gnc>.

## Sample Usage

```go
type Calc struct{ Total float64 } // System
type Add struct{ Amount float64 } // implements the rambo.Command interface
type Sub struct{ Amount float64 } // implements the rambo.Command interface
type Total struct{}               // implements the rambo.Query interface

func (c *Add) ExecuteOn(calc *Calc) error { calc.Total += c.Amount; return nil }
func (c *Sub) ExecuteOn(calc *Calc) error { calc.Total -= c.Amount; return nil }
func (q *Total) QueryOn(calc *Calc) any   { return calc.Total }

func main() {
	// Wrap `&Calc{}` with Rambo's prevalent layer;
	app, err := rambo.Load(
		"calc.journal", // ← The file to journal your changes to;
		&Calc{},        // ← Your in-memory system;
		&Add{},         // ← From here onwards, command samples;
		&Sub{},         //   These are required by Go's `gob`
	)
	if err != nil {
		panic(err)
	}

	// Execute commands via the prevalent layer…
	if err := app.Transact(&Add{Amount: 20}); err != nil {
		panic(err)
	}
	if err := app.Transact(&Sub{Amount: 7.5}); err != nil {
		panic(err)
	}

	// Query the system via the prevalent layer with a Query struct…
	result := app.Query(&Total{}).(float64)
	fmt.Printf("%g\n", result) // 12.5

	// …or with an anonymous function
	result = app.QueryFn(func (system *System) any { return system.Total })
	fmt.Printf("%g\n", result) // 12.5
}
```

## References

* [System Prevalence on Wikipedia](https://en.wikipedia.org/wiki/System_prevalence)
* [Prevayler Mailing List](https://sourceforge.net/p/prevayler/mailman/prevayler-discussion/)
* ["PrevaylerJr"](https://gist.github.com/klauswuestefeld/1103582) by [Klaus Wuestefeld](https://twitter.com/klauswuestefeld)
* ["An introduction to object prevalence"](https://web.archive.org/web/20061214033736/http://www-128.ibm.com/developerworks/library/wa-objprev/) by [Carlos Villela](https://twitter.com/cv)
* ["Prevalent Systems: A Pattern Language for Persistence"](https://web.archive.org/web/20170610140344/http://hillside.net/sugarloafplop/papers/5.pdf) by [Ralph E. Johnson](https://en.wikipedia.org/wiki/Ralph_Johnson_(computer_scientist)) and [Klaus Wuestefeld](https://twitter.com/klauswuestefeld)
* ["Object Prevalence"](https://web.archive.org/web/20170628061414/http://www.advogato.org/article/398.html) by [Klaus Wuestefeld](https://twitter.com/klauswuestefeld)
