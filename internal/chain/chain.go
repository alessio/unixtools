package chain

type Pipeline[T any] interface {
	Kickoff(V T) error
}

type Worker[T any] interface {
	Work(T) error
	Next(T) error
}

type Dispatcher[T any] interface {
	Next(T) error
	SetNext(Worker[T])
}

type dispatcher[T any] struct {
	next Worker[T]
}

func (d *dispatcher[T]) Next(i T) error {
	if d.next == nil {
		return nil
	}

	if err := d.next.Work(i); err != nil {
		return err
	}

	return d.next.Next(i)
}

func (d *dispatcher[T]) SetNext(n Worker[T]) {
	d.next = n
}
