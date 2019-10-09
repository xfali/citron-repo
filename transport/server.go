// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package transport

import (
    "context"
    "github.com/xfali/goutils/log"
    "net"
    "time"
)

const (
    ReadTimeout  = 15 * time.Second
    WriteTimeout = 15 * time.Second
)

type TcpTransport struct {
    port     string
    listener net.Listener
    connConf ConnConfig
    connList []*Connect
}

type Processor interface {
    //read channel
    ReadChan() chan<- []byte
    //write channel
    WriteChan() <-chan []byte
    //close channel
    CloseChan() <-chan bool

    //获得读缓存，必须保证获得的buf操作是线程安全的
    AcquireReadBuf() []byte
    ////归还缓存，必须保证归还的buf操作是线程安全的
    //ReleaseReadBuf(d []byte)
    ////获得写缓存，必须保证获得的buf操作是线程安全的
    //AcquireWriteBuf() []byte
    //归还缓存，必须保证归还的buf操作是线程安全的
    ReleaseWriteBuf([]byte)
}

type ProcessorFactory func() Processor

type Opt func(*TcpTransport)

func SetPort(port string) Opt {
    return func(t *TcpTransport) {
        t.port = port
    }
}

func SetReadTimeout(duration time.Duration) Opt {
    return func(t *TcpTransport) {
        t.connConf.readTimeout = duration
    }
}

func SetWriteTimeout(duration time.Duration) Opt {
    return func(t *TcpTransport) {
        t.connConf.writeTimeout = duration
    }
}

func SetListenerFactory(factory ProcessorFactory) Opt {
    return func(t *TcpTransport) {
        t.connConf.factory = factory
    }
}

func NewTcpTransport(opts ...Opt) *TcpTransport {
    ret := &TcpTransport{}
    for i := range opts {
        opts[i](ret)
    }
    return ret
}

func (t *TcpTransport) Startup() error {
    l, err := net.Listen("tcp", t.port)
    if err != nil {
        return err
    }
    t.listener = l

    ctx, cancel := context.WithCancel(context.Background())
    for {
        c, err := l.Accept()
        if err != nil {
            cancel()
            return err
        }
        t.handleConnect(ctx, c)
    }
}

func (t *TcpTransport) Close() error {
    err := t.listener.Close()
    for _, conn := range t.connList {
        conn.Close()
    }
    return err
}

func (t *TcpTransport) handleConnect(ctx context.Context, c net.Conn) {
    conn := NewConnect(t.connConf, c)
    t.connList = append(t.connList, conn)
    log.Debug("accept %v", c.RemoteAddr())
    go conn.ProcessLoop()
}
