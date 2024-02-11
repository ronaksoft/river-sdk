package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"

    "github.com/ronaksoft/river-msg/go/msg"
    riversdk "github.com/ronaksoft/river-sdk/sdk/prime"
    "gopkg.in/abiosoft/ishell.v2"
)

var Auth = &ishell.Cmd{
    Name: "Auth",
}

var AuthCheckPhone = &ishell.Cmd{
    Name: "CheckPhone",
    Func: func(c *ishell.Context) {
        req := msg.AuthCheckPhone{}
        req.Phone = fnGetPhone(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_AuthCheckPhone, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var AuthSendCode = &ishell.Cmd{
    Name: "SendCode",
    Func: func(c *ishell.Context) {
        req := msg.AuthSendCode{}
        req.Phone = fnGetPhone(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_AuthSendCode, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var AuthRegister = &ishell.Cmd{
    Name: "Register",
    Func: func(c *ishell.Context) {
        req := msg.AuthRegister{}

        req.Phone = fnGetPhone(c)
        req.PhoneCode = fnGetPhoneCode(c)
        req.PhoneCodeHash = fnGetPhoneCodeHash(c)
        req.FirstName = fnGetFirstName(c)
        req.LastName = fnGetLastName(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_AuthRegister, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }

    },
}

var AuthLogin = &ishell.Cmd{
    Name: "Login",
    Func: func(c *ishell.Context) {
        req := msg.AuthLogin{}
        phoneFile, err := os.Open("./_phone")
        if err != nil {
            req.Phone = fnGetPhone(c)
            req.PhoneCode = fnGetPhoneCode(c)
            req.PhoneCodeHash = fnGetPhoneCodeHash(c)

        } else {
            b, _ := ioutil.ReadAll(phoneFile)
            req.Phone = string(b)
            if strings.HasPrefix(req.Phone, "2374") {
                File, err := os.Open("./_phoneCodeHash")
                if err != nil {
                    req.PhoneCodeHash = fnGetPhoneCode(c)
                } else {
                    req.PhoneCode = req.Phone[len(req.Phone)-5:]
                    b, _ := ioutil.ReadAll(File)
                    req.PhoneCodeHash = string(b)
                }
            }
        }
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        os.Remove("./_phone")
        os.Remove("./_phoneCodeHash")
        if reqID, err := _SDK.ExecuteCommand(msg.C_AuthLogin, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }

    },
}

var AuthCheckPassword = &ishell.Cmd{
    Name: "CheckPassword",
    Func: func(c *ishell.Context) {
        passwordFile, err := os.Open("./_password")
        if err != nil {
            c.Println(err)
            return
        }
        b, _ := ioutil.ReadAll(passwordFile)
        passwordFile.Close()
        os.Remove("./_password")

        inputPassword := &msg.InputPassword{}
        inputPassword.Unmarshal(riversdk.GenInputPassword(fnGetPassword(c), b))
        req := &msg.AuthCheckPassword{
            Password: inputPassword,
        }
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_AuthCheckPassword, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var AuthLogout = &ishell.Cmd{
    Name: "Logout",
    Func: func(c *ishell.Context) {
        if err := _SDK.Logout(true, 0); err != nil {
            c.Println("Command Failed:", err)
        }
        _ = os.Remove(filepath.Join(_DbPath, fmt.Sprintf("connInfo.%s", _DbID)))
    },
}

var AuthRecall = &ishell.Cmd{
    Name: "Recall",
    Func: func(c *ishell.Context) {
        req := msg.AuthRecall{}
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_AuthRecall, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var AuthLoginByToken = &ishell.Cmd{
    Name: "LoginByToken",
    Func: func(c *ishell.Context) {
        req := msg.AuthLoginByToken{}
        req.Provider = fnGetProvider(c)
        req.Token = fnGetToken(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_AuthLoginByToken, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

func init() {
    Auth.AddCmd(AuthSendCode)
    Auth.AddCmd(AuthCheckPhone)
    Auth.AddCmd(AuthRegister)
    Auth.AddCmd(AuthLogin)
    Auth.AddCmd(AuthCheckPassword)
    Auth.AddCmd(AuthRecall)
    Auth.AddCmd(AuthLogout)
    Auth.AddCmd(AuthLoginByToken)
}
