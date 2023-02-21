package balance

import "errors"

type RoundRobinBalance struct {
	curIndex int
	rss      []string
}

func (r *RoundRobinBalance) Set(s []string) error {
	if len(s) == 0 {
		return errors.New("input []string")
	}

	r.rss = s
	return nil
}
func (r *RoundRobinBalance) next() string {
	if len(r.rss) == 0 {
		return ""
	}
	lens := len(r.rss)
	if r.curIndex >= lens {
		r.curIndex = 0
	}

	curAddr := r.rss[r.curIndex]
	r.curIndex = (r.curIndex + 1) % lens
	return curAddr
}
func (r *RoundRobinBalance) Get() string {
	return r.next()
}
