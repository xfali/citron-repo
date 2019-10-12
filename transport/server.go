// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package transport

import (
    "citron-repo/util"
    "context"
    "github.com/xfali/goutils/log"
    "io"
    "net"
    "sync"
    "time"
)

const (
    ReadTimeout  = 15 * time.Second
    WriteTimeout = 15 * time.Second
)

type TcpTransport struct {
    port     string
    listener net.Listener
    stop     bool
    connConf ConnConfig

    connMap sync.Map
}

type Observer interface {
    NotifyClosed(close io.Closer)
}

type Processor interface {
    //read channel
    ReadChan() chan<- []byte
    //write channel
    WriteChan() <-chan []byte
    //close channel
    Closer() util.Closable

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

func SetConfig(conf ConnConfig) Opt {
    return func(t *TcpTransport) {
        t.connConf = conf
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

func (t *TcpTransport) ListenAndServe() error {
    l, err := net.Listen("tcp", t.port)
    if err != nil {
        return err
    }
    t.listener = l

    ctx, cancel := context.WithCancel(context.Background())
    for {
        if t.stop {
            return nil
        }
        c, err := l.Accept()
        if err != nil {
            cancel()
            return err
        }
        t.handleConnect(ctx, c)
    }
}

func (t *TcpTransport) Close() error {
    t.stop = true
    err := t.listener.Close()
    t.connMap.Range(func(key, value interface{}) bool {
        key.(*Connect).Close()
        t.connMap.Delete(key)
        return true
    })
    return err
}

func (t *TcpTransport) handleConnect(ctx context.Context, c net.Conn) {
    conn := NewConnect(t.connConf, c)
    //observer
    conn.RegisterObserver(t)
    t.connMap.Store(conn, conn.conn.RemoteAddr())
    log.Debug("accept %v", c.RemoteAddr())
    go conn.ProcessLoop()
}

func (t *TcpTransport) NotifyClosed(closer io.Closer) {
    t.connMap.Delete(closer)
}
