package rambo

import (
	"encoding/gob"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Rambo is an implementation of the Prevalent System design pattern, in which
// business objects are kept live in memory and transactions are journaled for
// system recovery. A prevalent system needs enough memory to hold its entire
// state in RAM (the "prevalent hypothesis").
//
// This version is heavily inspired on Prevayler's author Klaus Wuestefeld's
// "PrevaylerJr" version (https://bit.ly/3tGxw3P).
//
// The name "Rambo" comes from a book called "Java RAMBO Manifesto: RAM-based
// Objects, Prevayler, and Raw Speed" (https://amzn.to/3IbPsaC) by Peter Wayner.
type Rambo[System any] struct {
	system  *System
	encoder *gob.Encoder
	mutex   sync.Mutex
}

// Load restores the system state from the disk and wraps the given system in a
// prevalent layer.
//
// The given initial system state is hydrated by loading the last persisted
// state from the file located at the given path and applying any journaled
// transactions. If no persisted state is found (very first time running the
// system), then given initial system state is used.
//
// IMPORTANT: don't forget to provide samples (empty struct pointers) of the
// commands as the last set of arguments. This is required for deserialization.
func Load[System any](storageFilePath string, initial *System, commandSamples ...any) (*Rambo[System], error) {
	gob.Register(initial)
	for _, commandSample := range commandSamples {
		gob.Register(commandSample)
	}

	tempFilePath := storageFilePath + ".tmp"

	filePath := tempFilePath
	if _, err := os.Stat(storageFilePath); !errors.Is(err, os.ErrNotExist) {
		filePath = storageFilePath
	}

	file, err := openFile(filePath)
	if err != nil {
		return nil, err
	}

	system, err := restoreSystem(initial, file)
	if err != nil {
		return nil, err
	}

	file, err = openFile(tempFilePath)
	encoder := gob.NewEncoder(file)

	if err != nil {
		return nil, err
	}

	err = encoder.Encode(system)
	if err != nil {
		return nil, err
	}

	err = os.Rename(tempFilePath, storageFilePath)
	if err != nil {
		return nil, err
	}

	return &Rambo[System]{
		system:  system,
		encoder: encoder,
	}, nil
}

// Command describes a transactional system state change. Implementations of
// this interface are used with the Rambo.Transact method.
type Command[System any] interface {
	ExecuteOn(*System) error
}

// Transact journals the given command and then executes it. Beware that the
// journaled commands will be replayed during bootstrap – during Load – so don't
// perform any third-party interactions from commands (e.g. sending emails o
// charging credit cards). Commands should be deterministic, meaning that they
// should always produce the same result when replayed.
func (p *Rambo[System]) Transact(command Command[System]) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.encoder.Encode(&command)
	if err != nil {
		return err
	}

	return command.ExecuteOn(p.system)
}

// Query describes a query to run against the system. Queries are not journaled.
// Implementations of this interface are used with Rambo.Query method.
type Query[System any] interface {
	QueryOn(*System) any
}

// Query executes the given query against the system. The output from the query
// is returned and must be type-casted in order to be used.
//
// See also: Rambo.QueryFn
func (p *Rambo[System]) Query(query Query[System]) any {
	return query.QueryOn(p.system)
}

// QueryFn executes the given query function against the system. The output from
// the query is returned and must be type-casted in order to be used.
//
// See also: Rambo.Query
func (p *Rambo[System]) QueryFn(fn func(*System) any) any {
	return fn(p.system)
}

func restoreSystem[System any](system *System, file *os.File) (*System, error) {
	decoder := gob.NewDecoder(file)

	err := decoder.Decode(&system)
	if err == io.EOF {
		return system, nil
	}
	if err != nil {
		return nil, err
	}

	for {
		var command Command[System]
		err := decoder.Decode(&command)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		_ = command.ExecuteOn(system)
	}

	return system, nil
}

func openFile(path string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(path), 0770)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
}
