// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package handler

import (
    "citron-repo/errcode"
    "citron-repo/model"
    "citron-repo/token"
    "github.com/gin-gonic/gin"
    "github.com/xfali/goutils/log"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

const (
    LOGIN_EXPIRE_TIME = 3 * time.Hour
    FILE_EXPIRE_TIME  = 15 * time.Second

    CITRON_TOKEN      = "CITRON-TOKEN"
    CITRON_FILE_TOKEN = "CITRON-FILE-TOKEN"
    CITRON_REL        = "CITRON-REL"
    CITRON_FILENAME   = "CITRON-FILENAME"
)

type restfulApi struct {
    conf     model.Config
    tokenMgr *token.TokenMgr
}

func NewRestful(conf model.Config) *restfulApi {
    ret := &restfulApi{
        conf:     conf,
        tokenMgr: token.New(),
    }
    return ret
}

func (rest *restfulApi) Close() {
    rest.tokenMgr.Close()
}

func (rest *restfulApi) Api(engine *gin.Engine) {
    engine.Handle(http.MethodPost, "/meta", rest.CreateMeta)
    engine.Handle(http.MethodPut, "/config", rest.Config)
    engine.Handle(http.MethodPost, "/login", rest.Login)
    engine.Handle(http.MethodPost, "/file", rest.upload)
}

func (rest *restfulApi) Login(ctx *gin.Context) {
    login := model.LoginInfo{}
    err := ctx.Bind(&login)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, errcode.LoginError)
        return
    }

    if rest.conf.Username == login.Username && rest.conf.Password == login.Password {
        ctx.JSON(http.StatusOK, errcode.Ok(rest.tokenMgr.CreateToken(login.Username, LOGIN_EXPIRE_TIME)))
    }

    ctx.JSON(http.StatusUnauthorized, errcode.AuthError)
    return
}

func checkToken(ctx *gin.Context) bool {
    token := ctx.GetHeader(CITRON_TOKEN)
    if token == "" {
        ctx.JSON(http.StatusUnauthorized, errcode.AuthError)
        return false
    }
    return true
}

//header 包含CITRON-TOKEN（登录token）
//header 包含CITRON-REL（相对目录)
//header 包含CITRON-FILENAME（文件名称)
func (rest *restfulApi) CreateMeta(ctx *gin.Context) {
    if !checkToken(ctx) {
        return
    }

    rel := ctx.GetHeader(CITRON_REL)
    filename := ctx.GetHeader(CITRON_FILENAME)
    if filename == "" {
        ctx.JSON(http.StatusBadRequest, errcode.FileTokenMissing)
        return
    }

    path := filepath.Join(rest.conf.BackupDir, rel, filename)
    fileToken := rest.tokenMgr.CreateToken(path, FILE_EXPIRE_TIME)

    ctx.JSON(http.StatusOK, errcode.Ok(fileToken))
}

func (rest *restfulApi) Config(ctx *gin.Context) {
    if !checkToken(ctx) {
        return
    }

    conf := model.Config{}
    err := ctx.Bind(&conf)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, errcode.ConfigError)
        return
    }

    rest.conf = conf
    ctx.JSON(http.StatusOK, errcode.OK)
}

//header 包含CITRON-TOKEN（登录token）
//header 包含CITRON-FILE-TOKEN（文件上传token)
func (rest *restfulApi) upload(ctx *gin.Context) {
    if !checkToken(ctx) {
        return
    }

    fileToken := ctx.GetHeader(CITRON_FILE_TOKEN)
    if fileToken == "" {
        ctx.JSON(http.StatusUnauthorized, errcode.FileTokenMissing)
        return
    }

    path := rest.tokenMgr.Get(fileToken)
    if path == "" {
        ctx.JSON(http.StatusUnauthorized, errcode.FileTokenError)
        return
    }

    file, _, err := ctx.Request.FormFile("file")
    if err != nil {
        ctx.JSON(http.StatusBadRequest, errcode.FileUploadFailed)
        return
    }
    //写入文件
    out, err := os.Create(path)
    if err != nil {
        log.Error("create file failed")
        ctx.JSON(http.StatusBadRequest, errcode.FileUploadFailed)
        return
    }

    defer out.Close()
    _, err = io.Copy(out, file)
    if err != nil {
        log.Error("copy file failed")
        ctx.JSON(http.StatusBadRequest, errcode.FileUploadFailed)
        return
    }

    ctx.JSON(http.StatusOK, errcode.OK)
}
