// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package test

import (
    "bytes"
    "citron-repo/client"
    "citron-repo/transport/binary"
    "fmt"
    "github.com/xfali/goutils/log"
    "os"
    "testing"
)

func TestServer(t *testing.T) {
    log.Level = log.DEBUG
    s := binary.NewBinaryServer()
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
