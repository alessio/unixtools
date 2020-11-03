package seq

import (
	"fmt"
	"sync"
)

// Sequence is implemented by types that generate sequence of strings.
type Sequence interface {
	// Items returns a channel containing all the sequence items.
	Items() <-chan string

	// WidthExceeded returns true if the an out of bounds error has occurred.
	WidthExceeded() bool
}

// NewInt creates a new string sequence of integers.
// Strings will not be padded if width is 0.
func NewInt(start int, incr uint, end int, width uint) Sequence {
	step := int(incr)
	if end < start {
		step = -step
	}

	seq := &intSequence{data: make(chan string), step: step, end: end, width: width, widthExceededMutex: sync.RWMutex{}}

	go seq.push(start)

	return seq
}

type intSequence struct {
	data               chan string
	step               int
	end                int
	width              uint
	widthExceeded      bool
	widthExceededMutex sync.RWMutex
}

// Items returns a channel containing all the sequence items.
func (s *intSequence) Items() <-chan string { return s.data }

// WidthExceeded returns true if the an out of bounds error has occurred.
func (s *intSequence) WidthExceeded() bool {
	s.widthExceededMutex.Lock()
	defer s.widthExceededMutex.Unlock()
	return s.widthExceeded
}

func (s *intSequence) push(start int) {
	for cur := start; (s.step > 0 && cur <= s.end) || (s.step < 0 && cur >= s.end); cur += s.step {
		if s.width == 0 {
			s.data <- fmt.Sprintf("%d", cur)
			continue
		}

		next := fmt.Sprintf(fmt.Sprintf("%%0%dd", s.width), cur)
		if int(s.width)-len(next) < 0 {
			func() {
				s.widthExceededMutex.RLock()
				defer s.widthExceededMutex.RUnlock()
				s.widthExceeded = true
			}()
			break
		}

		s.data <- next
	}

	close(s.data)
}
