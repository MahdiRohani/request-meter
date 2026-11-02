package domain

type Counter struct {
	Name  string
	Count int64
}

func NewCounter(name string) *Counter {
	return &Counter{
		Name:  name,
		Count: 0,
	}
}

func (c *Counter) Increment() {
	c.Count++
}
