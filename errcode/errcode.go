// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package errcode

import "github.com/xfali/go-web-starter/web/model"

var (
    OK         = model.OK

    ConfigError  = model.Result{Code: "801", Msg: "config failed"}

    LoginError = model.Result{Code: "1001", Msg: "login failed"}
    AuthError  = model.Result{Code: "1002", Msg: "login auth failed"}

    FilenamNotFound  = model.Result{Code: "2001", Msg: "file name not found, add it to header: CITRON-FILENAME"}
    FileUploadFailed  = model.Result{Code: "3001", Msg: "file upload failed"}
    FileTokenMissing  = model.Result{Code: "3002", Msg: "file token missing, add it to header: CITRON-FILE-TOKEN"}
    FileTokenError  = model.Result{Code: "3003", Msg: "file token error"}

    PackageNotReady  = model.Result{Code: "5001", Msg: "package not ready"}
)

func Ok(data interface{}) model.Result {
    return model.Ok(data)
}
