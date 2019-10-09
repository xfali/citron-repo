// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package test

import (
    "bytes"
    "citron-repo/client"
    "citron-repo/ioutil"
    "citron-repo/protocol"
    "citron-repo/transport"
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
    s := transport.NewBinaryServer(
        transport.SetTransport(transport.NewTcpTransport(
            transport.SetPort(":20001"))),
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

func readResponseHeader(c *client.TcpClient, w io.ReadWriter) protocol.ResponseHeader {
    n, er := c.ReceiveN(w, int64(protocol.ResponseHeaderSize))
    if n != int64(protocol.ResponseHeaderSize) {
        panic("error")
    }
    if er != nil {
        panic("error")
    }

    header := protocol.ResponseHeader{}
    client.ReadResponseHeader(&header, w)
    return header
}

func TestBinaryMultiPkg(t *testing.T) {
    log.Level = log.DEBUG
    s := transport.NewBinaryServer(
        transport.SetTransport(transport.NewTcpTransport(
            transport.SetPort(":20001"))),
    )

    go s.ListenAndServe()
    go http.ListenAndServe(":8001", nil)

    time.Sleep(time.Second)

    c := client.Open(":20001")

    b := make([]byte, 32*1024)
    buf := &ioutil.ByteWrapper{
        B: b,
    }

    client.WriteRequestHeader(buf, 3)
    buf.Write([]byte("123"))
    client.WriteRequestHeader(buf, 3)
    buf.Write([]byte("45"))

    c.Send(buf)

    r := &ioutil.ByteWrapper{B: b}
    header := readResponseHeader(c, r)
    t.Log(header)

    r.Reset()
    //body
    c.ReceiveN(r, 3)
    t.Log(string(r.Bytes()))

    buf.Reset()
    //finish send pkg
    buf.Write([]byte("6"))
    time.Sleep(time.Second)
    c.Send(buf)

    //next pkg
    r.Reset()
    header = readResponseHeader(c, r)
    t.Log(header)

    //body
    r.Reset()
    c.ReceiveN(r, 3)
    t.Log(string(r.Bytes()))

    select {
    case <-time.NewTimer(5 * time.Second).C:
        return
    }
}

func TestBinaryMultiPkgTimeout(t *testing.T) {
    log.Level = log.DEBUG
    s := transport.NewBinaryServer(
        transport.SetTransport(transport.NewTcpTransport(
            transport.SetPort(":20001"),
            transport.SetReadTimeout(500*time.Millisecond),
            transport.SetWriteTimeout(500*time.Millisecond))),
    )

    go s.ListenAndServe()
    go http.ListenAndServe(":8001", nil)

    time.Sleep(time.Second)

    c := client.Open(":20001")

    b := make([]byte, 32*1024)
    buf := &ioutil.ByteWrapper{
        B: b,
    }

    client.WriteRequestHeader(buf, 3)
    buf.Write([]byte("123"))
    client.WriteRequestHeader(buf, 3)
    buf.Write([]byte("45"))

    c.Send(buf)

    r := &ioutil.ByteWrapper{B: b}
    header := readResponseHeader(c, r)
    t.Log(header)

    r.Reset()
    //body
    c.ReceiveN(r, 3)
    t.Log(string(r.Bytes()))

    buf.Reset()
    //finish send pkg
    buf.Write([]byte("6"))
    //timeout here
    time.Sleep(time.Second)
    c.Send(buf)

    //next pkg
    r.Reset()
    header = readResponseHeader(c, r)
    t.Log(header)

    //body
    r.Reset()
    c.ReceiveN(r, 3)
    t.Log(string(r.Bytes()))

    select {
    case <-time.NewTimer(5 * time.Second).C:
        return
    }
}
