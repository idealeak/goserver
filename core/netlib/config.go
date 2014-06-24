// config
package netlib

import (
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"time"
)

var Config = Configuration{}

type Configuration struct {
	SrvInfo    ServerInfo
	IoServices []SessionConfig
}

type ServerInfo struct {
	Name   string
	Type   int
	Id     int
	AreaID int
	Banner []string
}

type SessionConfig struct {
	Id              int
	Type            int
	AreaId          int
	Name            string
	Ip              string
	Port            int
	MaxDone         int
	MaxPend         int
	MaxPacket       int
	MaxConn         int
	ExtraConn       int
	RcvBuff         int
	SndBuff         int
	SoLinger        int
	WriteTimeout    time.Duration
	ReadTimeout     time.Duration
	IdleTimeout     time.Duration
	KeepAlive       bool
	NoDelay         bool
	IsClient        bool
	IsAutoReconn    bool
	IsInnerLink     bool
	AuthKey         string //Authentication Key
	EncoderName     string //ProtocolEncoder name
	DecoderName     string //ProtocolDecoder name
	FilterChain     []string
	HandlerChain    []string
	IsKeepAlive     bool
	SupportFragment bool
	AllowMultiConn  bool
	encoder         ProtocolEncoder
	decoder         ProtocolDecoder
	sfc             *SessionFilterChain
	shc             *SessionHandlerChain
}

func (c *Configuration) Name() string {
	return "netlib"
}

func (c *Configuration) Init() error {
	for _, str := range c.SrvInfo.Banner {
		logger.Logger.Info(str)
	}

	for i := 0; i < len(c.IoServices); i++ {
		c.IoServices[i].Init()
	}
	return nil
}

func (c *Configuration) Close() error {
	return nil
}

func (sc *SessionConfig) Init() {
	if sc.EncoderName == "" {
		sc.encoder = GetProtocolEncoder(DefaultProtocolEncoderName)
	} else {
		sc.encoder = GetProtocolEncoder(sc.EncoderName)
	}
	if sc.DecoderName == "" {
		sc.decoder = GetProtocolDecoder(DefaultProtocoDecoderName)
	} else {
		sc.decoder = GetProtocolDecoder(sc.DecoderName)
	}

	for i := 0; i < len(sc.FilterChain); i++ {
		creator := GetSessionFilterCreator(sc.FilterChain[i])
		if creator != nil {
			if sc.sfc == nil {
				sc.sfc = NewSessionFilterChain()
			}
			if sc.sfc != nil {
				sc.sfc.AddLast(creator())
			}
		}
	}

	for i := 0; i < len(sc.HandlerChain); i++ {
		creator := GetSessionHandlerCreator(sc.HandlerChain[i])
		if creator != nil {
			if sc.shc == nil {
				sc.shc = NewSessionHandlerChain()
			}
			if sc.shc != nil {
				sc.shc.AddLast(creator())
			}
		}
	}
	if sc.IdleTimeout <= 0 {
		sc.IdleTimeout = 5 * time.Second
	} else {
		sc.IdleTimeout = sc.IdleTimeout * time.Second
	}
	sc.WriteTimeout = sc.WriteTimeout * time.Second
	sc.ReadTimeout = sc.ReadTimeout * time.Second
}

func (sc *SessionConfig) GetFilter(name string) SessionFilter {
	if sc.sfc != nil {
		return sc.sfc.GetFilter(name)
	}
	return nil
}

func (sc *SessionConfig) GetHandler(name string) SessionHandler {
	if sc.shc != nil {
		return sc.shc.GetHandler(name)
	}
	return nil
}
func init() {
	core.RegistePackage(&Config)
}
