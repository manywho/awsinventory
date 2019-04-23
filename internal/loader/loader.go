package loader

import (
	"sync"

	"github.com/itmecho/awsinventory/internal/inventory"
)

// Loader is responsible for loading data from the AWS API and storing it.
type Loader struct {
	lock   sync.Mutex
	Data   []inventory.Row
	Errors chan error
}

// NewLoader returns a new Loader with a preloaded AWS session and lock
func NewLoader() *Loader {
	return &Loader{
		lock:   sync.Mutex{},
		Data:   make([]inventory.Row, 0),
		Errors: make(chan error, 1),
	}
}

func (l *Loader) appendData(data []inventory.Row) {
	l.lock.Lock()
	defer l.lock.Unlock()

	for _, row := range data {
		l.Data = append(l.Data, row)
	}
}
