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
    "io"
    "net/http"
    _ "net/http/pprof"
    "strings"
    "testing"
    "time"
)

func TestBinary(t *testing.T) {
    log.Level = log.DEBUG
    s := binary.NewBinaryServer(
        binary.SetPort(":20001"),
    )

    go s.ListenAndServe()
    go http.ListenAndServe(":8001", nil)

    c := client.NewBinaryClient(":20001")
    buf := bytes.NewBuffer(make([]byte, 1024))
    time.Sleep(time.Second)
    for i := 0; i < 100; i++ {
        buf.Reset()
        msg := fmt.Sprintf("test %d", i)
        err := c.Send(int64(len(msg)), strings.NewReader(msg))
        if err != nil {
            t.Fatal(err)
        }
        r, err := c.Receive()
        if err != nil {
            t.Fatal(err)
        }
        io.Copy(buf, r)
        t.Logf("%s\n", string(buf.Bytes()))
    }

    select {
    case <-time.NewTimer(10 * time.Second).C:
        return
    }
}
