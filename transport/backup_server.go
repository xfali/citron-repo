// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package transport

import "github.com/xfali/goutils/log"

type Server struct {
    readChan  chan []byte
    writeChan chan []byte
    stopChan  chan bool
    transport *TcpTransport
}

func NewServer() *Server {
    s := Server{
        readChan:  make(chan []byte),
        writeChan: make(chan []byte),
        stopChan:  make(chan bool),
    }

    tcp := NewTcpTransport(
        SetPort(":20000"),
        SetReadChan(s.readChan),
        SetWriteChan(s.writeChan),
    )
    s.transport = tcp

    return &s
}

func (s *Server) ListenAndServe() {
    go s.transport.Startup()
    s.Process()
}

func (s *Server) Process() {
    for {
        select {
        case <-s.stopChan:
            return
        case d := <-s.readChan:
            log.Debug("receive : %s", string(d))
            s.writeChan <- []byte("server reply: " + string(d))
        }
    }
}

func (s *Server) Close() error {
    close(s.stopChan)
    return s.transport.Close()
}
