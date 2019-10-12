// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package transport

import (
    "citron-repo/ioutil"
    "citron-repo/util"
    "github.com/xfali/goutils/log"
    "net"
    "sync"
    "time"
)

type ConnConfig struct {
    readTimeout  time.Duration
    writeTimeout time.Duration

    observer Observer
    factory  ProcessorFactory
}

type Connect struct {
    conn     net.Conn
    stopChan util.Closable
    wait     sync.WaitGroup

    p Processor
    o []Observer

    //read timeout
    rt time.Duration
    //write timeout
    wt time.Duration
}

func NewConnect(conf ConnConfig, conn net.Conn) *Connect {
    p := conf.factory()
    ret := Connect{
        conn:     conn,
        stopChan: util.NewSafeCloseChan(),
        p:        p,
        rt:       conf.readTimeout,
        wt:       conf.writeTimeout,
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
    c.notifyClose()
}

func (c *Connect) RegisterObserver(o Observer) {
    c.o = append(c.o, o)
}

func (c *Connect) ReadLoop() {
    defer c.wait.Done()
    if c.p.ReadChan() == nil {
        return
    }

    for {
        data := c.p.AcquireReadBuf()
        if c.rt > 0 {
            c.conn.SetReadDeadline(time.Now().Add(c.rt))
        }
        n, err := c.conn.Read(data)
        if err != nil {
            log.Error(err.Error())
            //exit goroutine
            c.stopChan.Close()
            return
        }

        log.Debug("read %s %d", string(data[:n]), ioutil.GetGoroutineID())

        select {
        case <-c.stopChan.C():
            return
        case <-c.p.Closer().C():
            return
        case c.p.ReadChan() <- data[:n]:
            break
        }
    }
}

func (c *Connect) WriteLoop() {
    defer c.wait.Done()
    if c.p.WriteChan() == nil {
        return
    }

    for {
        var d []byte
        select {
        case <-c.stopChan.C():
            return
        case <-c.p.Closer().C():
            return
        case d = <-c.p.WriteChan():
            break
        }

        if c.wt > 0 {
            c.conn.SetWriteDeadline(time.Now().Add(c.wt))
        }
        n, err := c.conn.Write(d)
        if err != nil {
            log.Error(err.Error())
            //exit goroutine
            c.stopChan.Close()
            break
        }
        log.Debug("write %s", string(d[:n]))

        c.p.ReleaseWriteBuf(d)
    }
}

func (c *Connect) Close() error {
    c.stopChan.Close()
    return nil
}

func (c *Connect) notifyClose() {
    for _, o := range c.o {
        o.NotifyClosed(c)
    }
    //close processor
    c.p.Closer().Close()
}
