package testext

import (
	"fmt"
	"sync"
)

// NewFreeAddress returns a "host:port" generator for the given host that begins at the current port.
func NewFreeAddress(host string, startPort uint16) FreeAddress {
	return FreeAddress{
		Host:   host,
		Port:   startPort,
		locker: &sync.Mutex{},
	}
}

// FreeAddress helps you generate unique "host:port" strings for each of your tests.
// You start at a given port and each call to Next() will give you a new host:port one
// higher than the previous.
type FreeAddress struct {
	// Host is the hostname portion of the address to generate.
	Host string
	// Port is the port number to increment each time.
	Port uint16
	// locker makes sure that if we run tests in parallel that we truly get unique values.
	locker *sync.Mutex
}

// Next increments the base port and returns the unique host:port. For instance, the first
// call might return "localhost:9001" and the subsequent one will return "localhost:9002".
func (next *FreeAddress) Next() string {
	next.locker.Lock()
	defer next.locker.Unlock()

	// Format the string first so we don't skip the first port you specified.
	value := fmt.Sprintf("%s:%v", next.Host, next.Port)
	next.Port++

	return value
}
