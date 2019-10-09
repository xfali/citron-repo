// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package test

import (
    "bytes"
    "citron-repo/ioutil"
    "testing"
)

func TestIOUtil(t *testing.T) {
    t.Run("Copy", func(t *testing.T) {
        r := bytes.NewReader([]byte("12345678"))
        w := bytes.NewBuffer(nil)

        n, e := ioutil.Copy(w, r)

        t.Logf("%v %v", n, e)
        t.Logf("ret : %s", string(w.Bytes()))
    })

    t.Run("CopyN", func(t *testing.T) {
        r := bytes.NewReader([]byte("12345678"))
        w := bytes.NewBuffer(nil)

        n, e := ioutil.CopyN(w, r, 3)

        t.Logf("%v %v", n, e)
        t.Logf("ret : %s", string(w.Bytes()))
    })

    t.Run("CopyWithBuffer", func(t *testing.T) {
        r := bytes.NewReader([]byte("12345678"))
        w := bytes.NewBuffer(nil)

        buf := make([]byte, 3)
        n, e := ioutil.CopyWithBuffer(w, r, buf)

        t.Logf("%v %v", n, e)
        t.Logf("ret : %s", string(w.Bytes()))
    })

    t.Run("CopyNWithBuffer", func(t *testing.T) {
        r := bytes.NewReader([]byte("12345678"))
        w := bytes.NewBuffer(nil)

        buf := make([]byte, 3)
        n, e := ioutil.CopyNWithBuffer(w, r, 3, buf)

        t.Logf("%v %v", n, e)
        t.Logf("ret : %s", string(w.Bytes()))
    })

    t.Run("CopyNWithBuffer buf:100", func(t *testing.T) {
        r := bytes.NewReader([]byte("12345678"))
        w := bytes.NewBuffer(nil)

        buf := make([]byte, 100)
        n, e := ioutil.CopyNWithBuffer(w, r, 3, buf)

        t.Logf("%v %v", n, e)
        t.Logf("ret : %s", string(w.Bytes()))
    })
}
