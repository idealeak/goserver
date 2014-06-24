package main

import (
	"time"

	_ "github.com/idealeak/goserver/core/builtin/action"

	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/builtin/filter"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
	"github.com/idealeak/goserver/examples/protocol"
)

var (
	Config         = Configuration{}
	PressureModule = &PressureTest{}
	StartCnt       = 0
)

type Configuration struct {
	Count    int
	Connects netlib.SessionConfig
}

func (this *Configuration) Name() string {
	return "pressure"
}

func (this *Configuration) Init() error {
	this.Connects.Init()
	f := this.Connects.GetFilter(filter.AuthenticationFilterName)
	if f != nil {
		f.(*filter.AuthenticationFilter).SessionAuthHandler = func(s *netlib.Session, bSuc bool) {
			if bSuc {
				packet := &protocol.CSPacketPing{
					TimeStamb: proto.Int64(time.Now().Unix()),
					Message:   []byte("=1234567890abcderghijklmnopqrstuvwxyz="),
				}
				//for i := 0; i < 1024*32; i++ {
				//	packet.Message = append(packet.Message, byte('x'))
				//}
				proto.SetDefaults(packet)
				s.Send(packet)
			} else {
				logger.Logger.Info("SessionAuthHandler auth failed")
			}
		}
	}
	return nil
}

func (this *Configuration) Close() error {
	return nil
}

type PressureTest struct {
}

func (this PressureTest) ModuleName() string {
	return "pressure-module"
}

func (this *PressureTest) Init() {
	cfg := Config.Connects
	for i := 0; i < Config.Count; i++ {
		cfg.Id += i
		netlib.Connect(core.CoreObject(), &cfg)
	}
}

func (this *PressureTest) Update() {
	return
}

func (this *PressureTest) Shutdown(ownerAck chan<- interface{}) {
	ownerAck <- this.ModuleName()
}

func init() {
	module.RegisteModule(PressureModule, time.Second*30, 50)
	core.RegistePackage(&Config)
}

type openTimer struct {
}

func (t *openTimer) OnTimer(h timer.TimerHandle, ud interface{}) bool {

	if StartCnt >= Config.Count {

	} else {
		logger.Logger.Info("Start ", StartCnt, " Times Connect")
		cfg := Config.Connects
		cfg.Id = cfg.Id - StartCnt
		netlib.Connect(core.CoreObject(), &cfg)
		StartCnt = StartCnt + 1
	}

	return true
}
