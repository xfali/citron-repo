// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package client

import (
    "citron-repo/ioutil"
    "citron-repo/protocol"
    "encoding/binary"
    "errors"
    "github.com/xfali/goutils/log"
    "io"
)

const (
    MagicCode       = 0xC100
    Version         = 0x01
    ReadBufferSize  = 32 * 1024
    WriteBufferSize = 32 * 1024
)

type BinaryClient struct {
    sendBuffer []byte
    recvBuffer []byte
    client     *TcpClient
}

func NewBinaryClient(addr string) *BinaryClient {
    ret := &BinaryClient{
        sendBuffer: make([]byte, WriteBufferSize),
        recvBuffer: make([]byte, ReadBufferSize),
        client:     Open(addr),
    }
    return ret
}

func (c *BinaryClient) Close() error {
    if c.client != nil {
        return c.client.Close()
    }
    return nil
}

func WriteRequestHeader(w io.Writer, length int64) error {
    return binary.Write(w, binary.BigEndian, protocol.RequestHeader{
        MagicCode: MagicCode,
        Version:   Version,
        Length:    length,
    })
}

func ReadResponseHeader(resp *protocol.ResponseHeader, r io.Reader) error {
    return binary.Read(r, binary.BigEndian, resp)
}

func (c *BinaryClient) Send(length int64, body io.Reader) (err error) {
    w := &ioutil.ByteWrapper{B: c.sendBuffer}
    err = binary.Write(w, binary.BigEndian, protocol.RequestHeader{
        MagicCode: MagicCode,
        Version:   Version,
        Length:    length,
    })
    if err != nil {
        return err
    }
    _, err = c.client.Send(w)
    if err != nil {
        return err
    }
    if body != nil {
        _, err = c.client.Send(body)
        if err != nil {
            return err
        }
    }

    return nil
}

func (c *BinaryClient) Receive() (body io.Reader, err error) {
    r := &ioutil.ByteWrapper{B: c.recvBuffer}
    n, er := c.client.ReceiveN(r, int64(protocol.ResponseHeaderSize))
    if n != int64(protocol.ResponseHeaderSize) {
        return nil, errors.New("read header error")
    }
    if er != nil {
        err = er
        return
    }

    header := protocol.ResponseHeader{}
    err = binary.Read(r, binary.BigEndian, &header)
    if err != nil {
        return
    }

    log.Debug("header is %v", header)
    if header.MagicCode != MagicCode {
        err = errors.New("Magic Code Not Match ")
        return
    }
    if header.Version != Version {
        err = errors.New("Version Not Match ")
        return
    }

    return io.LimitReader(c.client.conn, header.Length), nil
}
