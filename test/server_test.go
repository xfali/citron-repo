// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package test

import (
    "bytes"
    "citron-repo/client"
    "citron-repo/transport"
    "citron-repo/util"
    "fmt"
    "github.com/xfali/goutils/log"
    "os"
    "sync"
    "testing"
    "time"
)

func TestServer(t *testing.T) {
    log.Level = log.DEBUG
    s := transport.NewBinaryServer(
        transport.SetTransport(transport.NewTcpTransport(
            transport.SetPort("20001"))),
    )
    s.ListenAndServe()
}

func TestClient(t *testing.T) {
    c := client.Open("127.0.0.1:20000")
    if c != nil {
        for i := 0; i < 1; i++ {
            c.Send(bytes.NewReader([]byte(fmt.Sprintf("test %d", i))))
            c.Receive(os.Stdout)
        }
    }
}

func TestClose(t *testing.T) {
    x := util.NewSafeCloseChan()
    w := sync.WaitGroup{}
    w.Add(1)
    go func() {
        defer w.Done()
        select {
        case <- x.C():
            return
        }
    }()

    <- time.NewTimer(time.Second).C
    t.Log(x.Close())
    t.Log(x.Close())
    t.Log(x.Close())

    w.Wait()
    t.Log("done")
}
