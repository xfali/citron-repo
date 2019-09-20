// All right reserved.
// Copyright (C) 2019, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description: 

package token

import (
    "github.com/xfali/goutils/container/recycleMap"
    "github.com/xfali/goutils/idUtil"
    "time"
)

type TokenMgr struct {
    rmap *recycleMap.RecycleMap
}

func New() *TokenMgr {
    ret := TokenMgr{
        rmap: recycleMap.New(),
    }
    ret.rmap.PurgeInterval = 10 * time.Millisecond
    ret.rmap.Run()
    return &ret
}

func (tm *TokenMgr) CreateToken(key string, duration time.Duration) string {
    token := idUtil.RandomId(32)
    tm.rmap.Set(token, key, duration)
    return token
}

func (tm *TokenMgr) Get(token string) string {
    return tm.rmap.Get(token).(string)
}

func (tm *TokenMgr) Close() {
    tm.rmap.Close()
}
