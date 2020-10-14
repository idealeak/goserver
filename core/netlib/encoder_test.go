package netlib

import (
	"testing"
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/idealeak/goserver/core/builtin/protocol"
)

//func BenchmarkMarshalPacket(b *testing.B) {
//	runtime.GOMAXPROCS(1)

//	c, err := net.Dial("tcp", "192.168.1.106:9999")
//	if err != nil {
//		log.Fatal(err)
//	}
//	if tcpconn, ok := c.(*net.TCPConn); ok {
//		tcpconn.SetLinger(5)
//		tcpconn.SetNoDelay(false)
//		tcpconn.SetKeepAlive(false)
//		tcpconn.SetReadBuffer(102400)
//		tcpconn.SetWriteBuffer(10240000)
//	}
//	sc := &SessionConfig{}
//	s := newTcpSession(1, c, sc, nil)

//	pck := &protocol.SSPacketAuth{AuthKey: proto.String("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"), Timestamp: proto.Int64(time.Now().Unix())}
//	proto.SetDefaults(pck)
//	tNow := time.Now()
//	b.StartTimer()

//	w := bytes.NewBuffer(nil)

//	for i := 0; i < b.N; i++ {
//		//for j := 0; j < 100; j++ {
//		//	b, err := MarshalPacket(pck)
//		//	if err == nil {
//		//		binary.Write(w, binary.LittleEndian, b)
//		//	}
//		//}
////		pck2 := &protocol.SSPacketAuth{AuthKey: proto.String("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" /*hex.EncodeToString(w.Bytes())*/), Timestamp: proto.Int64(time.Now().Unix())}
////		DefaultBuiltinProtocolEncoder.Encode(s, pck2, s.conn)
//		w.Reset()
//		//Gpb.Marshal(pck)
//	}

//	b.StopTimer()
//	fmt.Println("==========", time.Now().Sub(tNow), "  ==", b.N)
//}

func BenchmarkTypetest(b *testing.B) {
	pck := &protocol.SSPacketAuth{AuthKey: proto.String("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"), Timestamp: proto.Int64(time.Now().Unix())}
	proto.SetDefaults(pck)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		typetest(pck)
	}
	b.StopTimer()
}

func BenchmarkGetPacketId(b *testing.B) {
	pck := &protocol.SSPacketAuth{AuthKey: proto.String("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"), Timestamp: proto.Int64(time.Now().Unix())}
	proto.SetDefaults(pck)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		getPacketId(pck)
	}
	b.StopTimer()
}
