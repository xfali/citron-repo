// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package transport

import "net"

type TcpTransport struct {
    port string
}

func (t *TcpTransport) Startup() error {
    l, err := net.Listen("tcp4", t.port)
    if err != nil {
        return err
    }
    defer l.Close()

    for {
        c, err := l.Accept()
        if err != nil {
            return err
        }
        go t.handleConnect(c)
    }
}

func (t *TcpTransport) handleConnect(c net.Conn) {
    c.Close()
}