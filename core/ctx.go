package core

import (
	"github.com/idealeak/goserver/core/basic"
)

var (
	AppCtx *Ctx = newCtx()
)

type Ctx struct {
	*basic.Object
}

func newCtx() *Ctx {
	ctx := &Ctx{}
	ctx.init()
	return ctx
}

func (ctx *Ctx) init() {
	ctx.Object = basic.NewObject(ObjId_RootId,
		"root",
		basic.Options{
			MaxDone:      1024,
			QueueBacklog: 1024,
		},
		nil)
	ctx.UserData = ctx
}

func LaunchChild(o *basic.Object) {
	AppCtx.LaunchChild(o)
}

func Terminate(o *basic.Object) {
	AppCtx.Terminate(o)
}

func CoreObject() *basic.Object {
	return AppCtx.GetChildById(ObjId_CoreId)
}
