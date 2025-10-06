package cond

import "strconv"

type item struct {
	cmd     string
	fooStr  string
	barStr  string
	fooItem *item
	barItem *item
}

type task struct {
	b, e int
	s    []rune
}

type cond struct {
	iAction int
	sAction string
	f, s    string
}

func (c *cond) String() string {
	if c == nil {
		return "cond: пусто"
	}
	return "cond: cmd=" + c.sAction + "/" + strconv.Itoa(c.iAction) + "/. First=!" + c.f + "!, Second=!" + c.s + "!"
}

func (t task) String() string {
	return "from " + strconv.Itoa(t.b) + " to " + strconv.Itoa(t.e) + "=>" + string(t.s)
}
