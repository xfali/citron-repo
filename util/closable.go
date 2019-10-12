// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package util

import "errors"

var CLOSED = errors.New("CLOSED")

//线程安全的关闭监听器
type Closable interface {
    //监听是否已关闭，未关闭则阻塞，已关闭直接返回(false,false)
    C() <-chan bool
    //关闭监听channel，首次关闭channel返回nil，已经关闭返回util.CLOSED
    Close() error
    //是否已关闭(非阻塞，仅查询状态)
    IsClosed() bool
}
