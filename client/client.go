// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package client

import (
    "io"
    "net"
)

type TcpClient struct {
    conn net.Conn
}

func (c *TcpClient) Close() error {
    if c.conn != nil {
        return c.conn.Close()
    }
    return nil
}

func Open(addr string) *TcpClient {
    c := TcpClient{}
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return nil
    }
    c.conn = conn

    return &c
}

func (c *TcpClient) Send(reader io.Reader) (int64, error) {
    return io.Copy(c.conn, reader)
}

func (c *TcpClient) Receive(writer io.Writer) (int64, error) {
    return io.Copy(writer, c.conn)
}
