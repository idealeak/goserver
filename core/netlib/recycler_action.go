package netlib

import (
	"github.com/idealeak/goserver/core/container/recycler"
)

const (
	ActionRecyclerBacklog int = 128
)

var ActionRecycler = recycler.NewRecycler(
	ActionRecyclerBacklog,
	func() interface{} {
		return &action{}
	},
	"action_recycler",
)

func AllocAction() *action {
	a := ActionRecycler.Get()
	return a.(*action)
}

func FreeAction(a *action) {
	ActionRecycler.Give(a)
}
