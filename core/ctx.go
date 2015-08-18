package core

import (
	"fmt"
	"os"

	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/utils"
)

const (
	HOOK_BEFORE_START int = iota
	HOOK_AFTER_STOP
	HOOK_MAX
)

var (
	AppCtx *Ctx = newCtx()
	hooks  [HOOK_MAX][]hookfunc
)

type hookfunc func() error

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
	ctx.Object.Waitor = utils.NewWaitor()
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

func RegisteHook(hookpos int, f hookfunc) {
	if hookpos < 0 || hookpos > HOOK_MAX {
		return
	}
	hooks[hookpos] = append(hooks[hookpos], f)
}

func ExecuteHook(hookpos int) error {
	if hookpos < 0 || hookpos > HOOK_MAX {
		return nil
	}
	var err error
	for _, h := range hooks[hookpos] {
		err = h()
		if err != nil {
			return err
		}
	}
	return nil
}

func WritePid() {
	f, err := os.OpenFile(".pid", os.O_CREATE|os.O_EXCL|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
		return
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("%v", os.Getpid()))
}
