// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package transport

import (
    "citron-repo/ioutil"
    "github.com/xfali/goutils/log"
    "net"
    "sync"
)

type ConnConfig struct {
    ReadBufSize  int
    WriteBufSize int

    factory ListenerFactory
}

type Connect struct {
    conn     net.Conn
    stopChan chan bool
    wait     sync.WaitGroup

    readChan    chan<- []byte
    writeChan   <-chan []byte
    closeChan   <-chan bool
    readBufSize int
}

func NewConnect(conf ConnConfig, conn net.Conn) *Connect {
    l := conf.factory()
    r, w, c := l()
    ret := Connect{
        conn:        conn,
        stopChan:    make(chan bool),
        readBufSize: conf.ReadBufSize,
        readChan:    r,
        writeChan:   w,
        closeChan:   c,
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
    log.Info("connect closed %v", c.conn.RemoteAddr())
    c.conn.Close()
}

func (c *Connect) ReadLoop() {
    defer c.wait.Done()
    if c.readChan == nil {
        return
    }

    data := make([]byte, c.readBufSize)
    for {
        n, err := c.conn.Read(data)
        if err != nil {
            log.Error(err.Error())
            //exit goroutine
            close(c.stopChan)
            return
        }

        log.Debug("read %s %d", string(data[:n]), ioutil.GetGoroutineID())

        select {
        case <-c.stopChan:
            return
        case <-c.closeChan:
            return
        case c.readChan <- data[:n]:
            break
        }
    }
}

func (c *Connect) WriteLoop() {
    defer c.wait.Done()
    if c.writeChan == nil {
        return
    }
    for {
        var d []byte
        select {
        case <-c.stopChan:
            return
        case <-c.closeChan:
            return
        case d = <-c.writeChan:
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
