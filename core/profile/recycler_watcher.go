package profile

import (
	"github.com/idealeak/goserver/core/container/recycler"
)

const (
	WatcherRecyclerBacklog int = 128
)

var WatcherRecycler = recycler.NewRecycler(
	WatcherRecyclerBacklog,
	func() interface{} {
		return &TimeWatcher{}
	},
	"watcher_recycler",
)

func AllocWatcher() *TimeWatcher {
	t := WatcherRecycler.Get()
	return t.(*TimeWatcher)
}

func FreeWatcher(t *TimeWatcher) {
	WatcherRecycler.Give(t)
}
