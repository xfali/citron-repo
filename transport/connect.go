// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package transport

import (
    "github.com/xfali/goutils/log"
    "net"
    "sync"
)

type ConnConfig struct {
    ReadBufSize  int
    WriteBufSize int

    ReadChan  chan<- []byte
    WriteChan <-chan []byte
}

type Connect struct {
    conn     net.Conn
    stopChan chan bool
    wait     sync.WaitGroup

    conf ConnConfig
}

func NewConnect(conf ConnConfig, conn net.Conn) *Connect {
    ret := Connect{
        conn:     conn,
        stopChan: make(chan bool),
        conf:     conf,
    }

    return &ret
}

func (c *Connect) ProcessLoop() {
    //read and write
    c.wait.Add(2)

    go c.ReadLoop()
    go c.WriteLoop()

    //wait read and write finished
    c.wait.Wait()
    c.conn.Close()
}

func (c *Connect) ReadLoop() {
    defer c.wait.Done()
    if c.conf.ReadChan == nil {
        return
    }

    data := make([]byte, c.conf.ReadBufSize)
    for {
        n, err := c.conn.Read(data)
        if err != nil {
            log.Error(err.Error())
            //exit goroutine
            close(c.stopChan)
            return
        }

        log.Debug("read %s", string(data[:n]))

        select {
        case <-c.stopChan:
            return
        case c.conf.ReadChan <- data[:n]:
            break
        }
    }
}

func (c *Connect) WriteLoop() {
    defer c.wait.Done()
    if c.conf.WriteChan == nil {
        return
    }
    for {
        var d []byte
        select {
        case <-c.stopChan:
            return
        case d = <-c.conf.WriteChan:
            break
        }

        n, err := c.conn.Write(d)
        if err != nil {
            log.Error(err.Error())
            //exit goroutine
            close(c.stopChan)
            break
        }
        log.Debug("write %s", string(d[:n]))
    }
}

func (c *Connect) Close() {
    close(c.stopChan)
}
