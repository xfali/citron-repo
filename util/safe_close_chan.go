// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package util

import (
    "sync/atomic"
)

type SafeCloseChan struct {
    closed   int32
    stopChan chan bool
}

func NewSafeCloseChan() *SafeCloseChan {
    ret := SafeCloseChan{
        closed:   0,
        stopChan: make(chan bool),
    }
    return &ret
}

func (c *SafeCloseChan) C() <-chan bool {
    return c.stopChan
}

//关闭监听channel，首次关闭channel返回nil，已经关闭返回error:"已关闭"
func (c *SafeCloseChan) Close() error {
    if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
        close(c.stopChan)
        return nil
    }
    return CLOSED
}

//是否已关闭
func (c *SafeCloseChan) IsClosed() bool {
    return atomic.LoadInt32(&c.closed) == 1
}
