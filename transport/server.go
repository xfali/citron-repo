// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package transport

import (
    "context"
    "net"
    "time"
)

const (
    ReadTimeout  = 15 * time.Second
    WriteTimeout = 15 * time.Second
    ReadBufSize  = 4096
)

type TcpTransport struct {
    port     string
    listener net.Listener
    connConf ConnConfig
    connList []*Connect
}

type Opt func(*TcpTransport)

func SetPort(port string) Opt {
    return func(t *TcpTransport) {
        t.port = port
    }
}

func SetReadBufSize(size int) Opt {
    return func(t *TcpTransport) {
        if size <= 0 {
            size = ReadBufSize
        }
        t.connConf.ReadBufSize = size
    }
}

func SetReadChan(readChan chan<- []byte) Opt {
    return func(t *TcpTransport) {
        t.connConf.ReadChan = readChan
    }
}

func SetWriteChan(writeChan <-chan []byte) Opt {
    return func(t *TcpTransport) {
        t.connConf.WriteChan = writeChan
    }
}

func NewTcpTransport(opts ...Opt) *TcpTransport {
    ret := &TcpTransport{}
    for i := range opts {
        opts[i](ret)
    }
    if ret.connConf.ReadBufSize == 0 {
        ret.connConf.ReadBufSize = ReadBufSize
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
    go conn.ProcessLoop()
}
