package graphs

// A Set is a container that contains each element just once.
type Set map[interface{}]struct{}

// NewSet creates a new empty set.
func NewSet() *Set {
	return &Set{}
}

// Add adds an element to the set. It returns true if the
// element has been added and false if the set already contained
// that element.
func (s *Set) Add(element interface{}) bool {
	_, exists := (*s)[element]
	(*s)[element] = struct{}{}
	return !exists
}

// Len returns the number of elements.
func (s *Set) Len() int {
	return len(*s)
}

// Iter returns a channel where all elements of the set
// are sent to.
func (s *Set) Iter() chan interface{} {
	ch := make(chan interface{})
	go func() {
		for v := range *s {
			ch <- v
		}
		close(ch)
	}()
	return ch
}
