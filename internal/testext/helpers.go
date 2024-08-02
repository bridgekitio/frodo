package testext

import "sync"

// Sequence is a convenient way to capture a series of values in a specified order that you can use
// to determine if code was fired in a specific sequence/order.
type Sequence struct {
	lock   sync.Mutex
	values []string
	wg     *sync.WaitGroup
}

// Append writes the next value for the piece of code that executed.
func (seq *Sequence) Append(value string) {
	seq.lock.Lock()
	defer seq.lock.Unlock()
	seq.values = append(seq.values, value)
}

// Value returns the value at the specific index. If you haven't appended that much yet, then this
// will return "".
func (seq *Sequence) Value(index int) string {
	if index >= len(seq.values) {
		return ""
	}
	return seq.values[index]
}

// Values returns all the values that you collected during the test case.
func (seq *Sequence) Values() []string {
	return seq.values
}

// Last returns the most recent value from the sequence, so you can check them as they come in.
func (seq *Sequence) Last() string {
	if len(seq.values) <= 0 {
		return ""
	}
	return seq.values[len(seq.values)-1]
}

// WaitGroup returns the wait group that helps synchronize the async routines filling this sequence.
func (seq *Sequence) WaitGroup() *sync.WaitGroup {
	return seq.wg
}

// Reset erases all current values in the sequence, allowing you to re-use this sequence multiple
// times within the same test case.
func (seq *Sequence) Reset() {
	seq.ResetWithWorkers(0)
}

// ResetWithWorkers erases all current values in the sequence and update the underlying wait group
// to expect this many 'Done()' invocations. This allows you to re-use this sequence multiple times
// within the same test case.
func (seq *Sequence) ResetWithWorkers(waitGroupCount int) {
	seq.values = nil
	seq.wg = &sync.WaitGroup{}
	seq.wg.Add(waitGroupCount)
}
