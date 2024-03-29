package main

import (
    "sync"

    "github.com/abiosoft/ishell/v2"
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/rony"
)

var Label = &ishell.Cmd{
    Name: "Label",
}

func init() {
    Label.AddCmd(LabelGet)
    Label.AddCmd(LabelListItems)
    Label.AddCmd(LabelCreate)
    Label.AddCmd(LabelTest)
    Label.AddCmd(LabelAddToMessage)
    Label.AddCmd(LabelRemoveFromMessage)
}

var LabelGet = &ishell.Cmd{
    Name: "Get",
    Func: func(c *ishell.Context) {
        req := msg.LabelsGet{}
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        // reqDelegate.FlagsVal = riversdk.RequestServerForced
        if reqID, err := _SDK.ExecuteCommand(msg.C_LabelsGet, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }

    },
}

var LabelCreate = &ishell.Cmd{
    Name: "Create",
    Func: func(c *ishell.Context) {
        req := msg.LabelsCreate{}
        req.Name = fnGetLabelName(c)
        req.Colour = fnGetLabelColour(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_LabelsCreate, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var LabelListItems = &ishell.Cmd{
    Name: "ListItems",
    Func: func(c *ishell.Context) {
        req := msg.LabelsListItems{}
        req.LabelID = fnGetLabelID(c)
        req.Limit = fnGetLimit(c)
        req.MinID = fnGetMinID(c)
        req.MaxID = fnGetMaxID(c)

        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        c.Println("LabelFill", repo.Labels.GetFilled(domain.GetCurrTeamID(), req.LabelID))
        if reqID, err := _SDK.ExecuteCommand(msg.C_LabelsListItems, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }

    },
}

var LabelAddToMessage = &ishell.Cmd{
    Name: "AddToMessage",
    Func: func(c *ishell.Context) {
        req := msg.LabelsAddToMessage{}
        req.Peer = fnGetPeer(c)
        getHistory(c, req.Peer)
        req.LabelIDs = []int32{fnGetLabelID(c)}
        req.MessageIDs = []int64{fnGetMessageID(c)}
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_LabelsAddToMessage, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

func getHistory(c *ishell.Context, p *msg.InputPeer) {
    req := msg.MessagesGetHistory{}
    req.Peer = p
    req.Limit = 10
    reqBytes, _ := req.Marshal()
    reqDelegate := new(RequestDelegate)
    if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGetHistory, reqBytes, reqDelegate); err != nil {
        c.Println("Command Failed:", err)
    } else {
        reqDelegate.RequestID = reqID
    }
}

var LabelRemoveFromMessage = &ishell.Cmd{
    Name: "RemoveFromMessage",
    Func: func(c *ishell.Context) {
        req := msg.LabelsRemoveFromMessage{}
        req.Peer = fnGetPeer(c)
        getHistory(c, req.Peer)
        req.LabelIDs = []int32{fnGetLabelID(c)}
        req.MessageIDs = []int64{fnGetMessageID(c)}
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_LabelsRemoveFromMessage, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var LabelTest = &ishell.Cmd{
    Name: "Test",
    Func: func(c *ishell.Context) {
        labels := getLabels(c)
        var labelIDs []int32
        for _, l := range labels {
            labelIDs = append(labelIDs, l.ID)
        }
        msgIDs := fnGetMessageIDs(c)
        addLabelToMessage(c, msgIDs, labelIDs)
        getMessage(c, msgIDs)
        removeLabelFromMessage(c, msgIDs, labelIDs)
        getMessage(c, msgIDs)
    },
}

func getLabels(c *ishell.Context) (labels []*msg.Label) {
    req := msg.LabelsGet{}
    reqBytes, _ := req.Marshal()
    reqD := NewCustomDelegate()
    wg := sync.WaitGroup{}
    wg.Add(1)
    reqD.OnCompleteFunc = func(b []byte) {
        defer wg.Done()
        x := &rony.MessageEnvelope{}
        _ = x.Unmarshal(b)
        switch x.Constructor {
        case msg.C_LabelsMany:
            xx := &msg.LabelsMany{}
            _ = xx.Unmarshal(x.Message)
            labels = xx.Labels
        default:
            c.Println(x)
        }
    }
    reqD.OnTimeoutFunc = func(err error) {
        wg.Done()
    }
    if reqID, err := _SDK.ExecuteCommand(msg.C_LabelsGet, reqBytes, reqD); err != nil {
        c.Println("Command Failed:", err)
    } else {
        reqD.RequestID = reqID
    }
    wg.Wait()
    return

}
func addLabelToMessage(c *ishell.Context, msgIDs []int64, labelIDs []int32) {
    req := msg.LabelsAddToMessage{
        Peer: &msg.InputPeer{
            ID:         _SDK.ConnInfo.UserID,
            Type:       msg.PeerType_PeerUser,
            AccessHash: 0,
        },
        LabelIDs:   labelIDs,
        MessageIDs: msgIDs,
    }
    reqBytes, _ := req.Marshal()
    reqD := NewCustomDelegate()
    wg := sync.WaitGroup{}
    wg.Add(1)
    reqD.OnCompleteFunc = func(b []byte) {
        defer wg.Done()
        x := &rony.MessageEnvelope{}
        _ = x.Unmarshal(b)
        switch x.Constructor {
        case msg.C_Bool:
            c.Println(x)
        default:
            c.Println(x)
        }
    }
    reqD.OnTimeoutFunc = func(err error) {
        wg.Done()
    }
    if reqID, err := _SDK.ExecuteCommand(msg.C_LabelsAddToMessage, reqBytes, reqD); err != nil {
        c.Println("Command Failed:", err)
    } else {
        reqD.RequestID = reqID
    }
    wg.Wait()
}
func removeLabelFromMessage(c *ishell.Context, msgIDs []int64, labelIDs []int32) {
    req := msg.LabelsRemoveFromMessage{
        Peer: &msg.InputPeer{
            ID:         _SDK.ConnInfo.UserID,
            Type:       msg.PeerType_PeerUser,
            AccessHash: 0,
        },
        LabelIDs:   labelIDs,
        MessageIDs: msgIDs,
    }
    reqBytes, _ := req.Marshal()
    reqD := NewCustomDelegate()
    wg := sync.WaitGroup{}
    wg.Add(1)
    reqD.OnCompleteFunc = func(b []byte) {
        defer wg.Done()
        x := &rony.MessageEnvelope{}
        _ = x.Unmarshal(b)
        switch x.Constructor {
        case msg.C_Bool:
            c.Println(x)
        default:
            c.Println(x)
        }
    }
    reqD.OnTimeoutFunc = func(err error) {
        wg.Done()
    }
    if reqID, err := _SDK.ExecuteCommand(msg.C_LabelsRemoveFromMessage, reqBytes, reqD); err != nil {
        c.Println("Command Failed:", err)
    } else {
        reqD.RequestID = reqID
    }
    wg.Wait()
}
func getMessage(c *ishell.Context, msgIDs []int64) {
    req := msg.MessagesGet{
        Peer: &msg.InputPeer{
            ID:         _SDK.ConnInfo.UserID,
            Type:       msg.PeerType_PeerUser,
            AccessHash: 0,
        },
        MessagesIDs: msgIDs,
    }
    reqBytes, _ := req.Marshal()
    reqD := NewCustomDelegate()
    wg := sync.WaitGroup{}
    wg.Add(1)
    reqD.OnCompleteFunc = func(b []byte) {
        defer wg.Done()
        x := &rony.MessageEnvelope{}
        _ = x.Unmarshal(b)
        switch x.Constructor {
        case msg.C_MessagesMany:
            xx := &msg.MessagesMany{}
            _ = xx.Unmarshal(x.Message)
            for _, m := range xx.Messages {
                c.Println(m.ID, m.Body, m.LabelIDs)
            }
        default:
            c.Println(x)
        }
    }
    reqD.OnTimeoutFunc = func(err error) {
        wg.Done()
    }
    if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGet, reqBytes, reqD); err != nil {
        c.Println("Command Failed:", err)
    } else {
        reqD.RequestID = reqID
    }
    wg.Wait()
}
