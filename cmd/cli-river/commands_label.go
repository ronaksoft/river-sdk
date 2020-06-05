package main

import (
	"git.ronaksoftware.com/river/msg/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

var Label = &ishell.Cmd{
	Name: "Label",
}

var LabelGet = &ishell.Cmd{
	Name: "Get",
	Func: func(c *ishell.Context) {
		req := msg.LabelsGet{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_LabelsListItems, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

func init() {
	Label.AddCmd(LabelGet)
	Label.AddCmd(LabelListItems)
	Label.AddCmd(LabelCreate)
}
