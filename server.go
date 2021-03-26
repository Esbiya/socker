package socker

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/Esbiya/loguru"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
)

type (
	Conn   gnet.Conn
	Action gnet.Action
	Server struct {
		*gnet.EventServer
		mode         string
		addr         string
		router       *Router
		pool         *ants.Pool
		clients      int
		connected    int32
		disconnected int32
		timeout      time.Duration
		multicore    bool
	}
)

const (
	None     Action = iota
	Close           // 连接关闭
	ShutDown        // 服务关闭
	Continue        // 消息继续
	Done            // 消息结束
)

const (
	DefaultAntsPoolSize = 1 << 18

	ExpiryDuration = 10 * time.Second

	Nonblocking = true
)

const (
	TCP = "tcp"
	UDS = "uds"
	UDP = "udp"
)

const (
	DefaultUDSAddr = "/tmp/us.socket"

	DefaultTCPAddr = ":20124"
)

func (u *Server) Router() *Router {
	return u.router
}

func (u *Server) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	loguru.Info("server is listening on %s (multi-cores: %t, loops: %d)", srv.Addr.String(), srv.Multicore, srv.NumEventLoop)
	return
}

func (u *Server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	c.SetContext(c)
	atomic.AddInt32(&u.connected, 1)
	// msg := NewMessage("client.init", "client.init.sign", "hello world")
	// out = MergeBytes(IntToBytes(msg.length), msg.bytes)
	// if c.LocalAddr() == nil {
	// 	panic("nil local addr")
	// }
	// if c.RemoteAddr() == nil {
	// 	panic("nil local addr")
	// }
	return
}

func (u *Server) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	if c.Context() != c {
		panic("invalid context")
	}

	atomic.AddInt32(&u.disconnected, 1)
	if atomic.LoadInt32(&u.connected) == atomic.LoadInt32(&u.disconnected) &&
		atomic.LoadInt32(&u.disconnected) == int32(u.clients) {
		action = gnet.Shutdown
	}

	return
}

func NewServer(mode, addr string, multicore bool, poolSize int, timeout time.Duration) *Server {
	options := ants.Options{ExpiryDuration: ExpiryDuration, Nonblocking: Nonblocking}
	pool, _ := ants.NewPool(poolSize, ants.WithOptions(options))
	return &Server{
		mode: mode,
		addr: addr,
		router: &Router{
			Handlers: map[string]Handler{},
		},
		pool:      pool,
		timeout:   timeout,
		multicore: multicore,
	}
}

// 记录日志
func (u *Server) RecordLog() {
	err := loguru.Enable(loguru.FileLog)
	if err != nil {
		loguru.Error("log write file error: %v", err)
		os.Exit(1)
	}
}

func (u *Server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	messages := ProcessMessages(frame)

	for _, message := range messages {
		if message.err != nil {
			message.reset(true, map[string]interface{}{
				"code": 400,
				"msg":  "message decode err! ",
			})
			_ = c.AsyncWrite(message.out())
		} else {
			loguru.Debug("receive message - length: %d, body: %s", message.bodyLength, message.BodyStringify())
			_ = u.pool.Submit(func() {
				next := u.router.Get(message.Api)
				var out interface{}
				for next != nil {
					out, next = next(message.ToData())
					message.reset(false, out)
					loguru.Debug("reply   message - length: %d, body: %s", message.bodyLength, message.BodyStringify())
					_ = c.AsyncWrite(message.out())
				}
			})
		}
	}
	return
}

func (u *Server) version() {
	fmt.Println(loguru.Fuchsia(`                    __`))
	fmt.Println(loguru.Fuchsia(`   _________  _____/ /_____  _____`))
	fmt.Println(loguru.Fuchsia(`  / ___/ __ \/ ___/ //_/ _ \/ ___/`))
	fmt.Println(loguru.Fuchsia(` (__  ) /_/ / /__/ ,< /  __/ /`))
	fmt.Println(loguru.Fuchsia(fmt.Sprintf(`/____/\____/\___/_/|_|\___/_/       socker v0.0.1 %s/%s`, runtime.GOOS, runtime.GOARCH)))
}

func (u *Server) Run() {
	u.version()
	defer u.pool.Release()
	log.Fatal(gnet.Serve(u, fmt.Sprintf("%s://%s", u.mode, u.addr), gnet.WithMulticore(u.multicore)))
}

func DefaultUDSServer() *Server {
	return NewServer(UDS, DefaultUDSAddr, true, DefaultAntsPoolSize, ExpiryDuration)
}

func DefaultTCPServer() *Server {
	return NewServer(TCP, DefaultTCPAddr, true, DefaultAntsPoolSize, ExpiryDuration)
}
