// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package main

import (
    "citron-repo/handler"
    "citron-repo/model"
    "citron-repo/transport"
    "flag"
    "github.com/xfali/go-web-starter/config"
)

func main() {
    username := flag.String("u", "", "username")
    password := flag.String("a", "", "password")
    port := flag.Int("p", 8080, "port")
    backupDir := flag.String("b", "./backup", "dir to backup")

    conf := config.Default()
    conf.ServerPort = *port

    myconf := model.Config{}
    myconf.Username = *username
    myconf.Password = *password
    myconf.BackupDir = *backupDir

    handler := handler.NewRestful(myconf)
    defer handler.Close()

    //web.StartupWithConf(conf, handler.Api)
    s := transport.NewServer()
    s.ListenAndServe()
}
