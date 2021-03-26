package socker

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/Esbiya/loguru"
	"github.com/panjf2000/ants/v2"
	"github.com/silenceper/pool"
)

var retMap *sync.Map

type (
	Callback struct {
		sign string
		Err  error
		Done bool
		Body chan interface{}
	}
	Config struct {
		Mode        string
		Addr        string
		InitCap     int
		MaxIdle     int
		MaxCap      int
		IdleTimeout time.Duration
	}
	Client struct {
		gpool       *ants.Pool
		connectPool pool.Pool
		writeChan   chan []byte
		close       chan struct{}
	}
)

func (c *Callback) Then(callback func(b interface{})) *Callback {
	callback(<-c.Body)
	return c
}

func (c *Callback) Catch(callback func(err error)) *Callback {
	callback(c.Err)
	return c
}

func (c *Callback) Close() {
	close(c.Body)
	retMap.Delete(c.sign)
}

func NewClient(cfg *Config) (*Client, error) {
	retMap = &sync.Map{}
	poolConfig := &pool.Config{
		Factory:     func() (interface{}, error) { return net.Dial(cfg.Mode, cfg.Addr) },
		Close:       func(v interface{}) error { return v.(net.Conn).Close() },
		Ping:        func(v interface{}) error { _, err := v.(net.Conn).Read(make([]byte, 0)); return err },
		InitialCap:  cfg.InitCap,     // 资源池初始连接数
		MaxIdle:     cfg.MaxIdle,     // 最大空闲连接数
		MaxCap:      cfg.MaxCap,      // 最大并发连接数
		IdleTimeout: cfg.IdleTimeout, // 连接最大空闲时间，超过该时间的连接将会关闭，可避免空闲时连接EOF，自动失效的问题
	}
	connectPool, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil, err
	}
	gpool, err := ants.NewPool(1 << 18)
	c := &Client{
		gpool:       gpool,
		connectPool: connectPool,
		writeChan:   make(chan []byte),
		close:       make(chan struct{}),
	}
	return c, err
}

func DefaultUDSClient() *Client {
	c, _ := NewClient(&Config{
		Mode:        UDS,
		Addr:        DefaultUDSAddr,
		InitCap:     100,
		MaxIdle:     100,
		MaxCap:      1 << 18,
		IdleTimeout: 5 * time.Second,
	})
	return c
}

func DefaultTCPClient() *Client {
	c, _ := NewClient(&Config{
		Mode:        TCP,
		Addr:        DefaultTCPAddr,
		InitCap:     100,
		MaxIdle:     100,
		MaxCap:      1 << 18,
		IdleTimeout: 5 * time.Second,
	})
	return c
}

func (c *Client) read(conn net.Conn) {
	defer c.connectPool.Put(conn)
	for {
		// 读取消息长度
		s := make([]byte, 4)
		_, err := conn.Read(s)
		if err == io.EOF {
			return
		}
		if err != nil {
			continue
		}
		size := BytesToInt(s)

		// 读取消息主体
		d := make([]byte, size)
		l, err := conn.Read(d)
		if err != nil {
			loguru.Error("read data error: %v", err)
			continue
		}

		msg := Message{}
		err = msg.Parse(d[:l])
		if err != nil {
			loguru.Error(err)
		}
		if v, ok := retMap.Load(msg.Sign); ok {
			v.(chan interface{}) <- msg.Body
		}
		if msg.Done {
			return
		}
	}
}

func (c *Client) loopTask() {
	for data := range c.writeChan {
		select {
		case <-c.close:
			return
		default:
			c1, _ := c.connectPool.Get()
			conn := c1.(net.Conn)
			conn.Write(data)
			_ = c.gpool.Submit(func() {
				c.read(conn)
			})
		}
	}
}

func (c *Client) Send(api string, context interface{}) *Callback {
	sign := GenUUIDStr()
	message := NewMessage(api, sign, context)

	c.writeChan <- MergeBytes(IntToBytes(message.length), message.bytes)

	out := make(chan interface{})
	retMap.Store(sign, out)
	return &Callback{
		sign: sign,
		Body: out,
	}
}

func (c *Client) Start() {
	go c.loopTask()
}

func (c *Client) Close() {
	close(c.close)
	c.gpool.Release()
	c.connectPool.Release()
}
