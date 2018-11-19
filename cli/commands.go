package main

import (
	"strconv"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

func fnGetPhone(c *ishell.Context) string {
	c.Print("Phone: ")
	phone := c.ReadLine()
	return phone
}

func fnGetPhoneCode(c *ishell.Context) string {
	c.Print("Phone Code: ")
	code := c.ReadLine()
	return code
}

func fnGetPhoneCodeHash(c *ishell.Context) string {
	c.Print("Phone Code Hash: ")
	hash := c.ReadLine()
	return hash
}

func fnGetFirstName(c *ishell.Context) string {
	c.Print("First Name: ")
	fName := c.ReadLine()
	return fName
}

func fnGetLastName(c *ishell.Context) string {
	c.Print("Last Name: ")
	lName := c.ReadLine()
	return lName
}

func fnGetSearchPhrase(c *ishell.Context) string {
	c.Print("Search Phrase:")
	phrase := c.ReadLine()
	return phrase
}

func fnGetPeerID(c *ishell.Context) int64 {
	var peerID int64
	for {
		c.Print("Peer ID: ")
		id, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			peerID = id
			break
		} else {
			c.Println(err.Error())
		}
	}
	return peerID
}

func fnGetAccessHash(c *ishell.Context) uint64 {
	var accessHash uint64
	for {
		c.Print("Access Hash: ")
		hash, err := strconv.ParseUint(c.ReadLine(), 10, 64)
		if err == nil {
			accessHash = uint64(hash)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return accessHash
}

func fnGetTries(c *ishell.Context) int {
	var count int
	for {
		c.Print("Tries : ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			count = int(tmp)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return count
}

func fnGetInterval(c *ishell.Context) time.Duration {
	var interval time.Duration
	for {
		c.Print("Interval: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			interval = time.Duration(tmp) * time.Millisecond
			break
		} else {
			c.Println(err.Error())
		}
	}
	return interval
}

func fnGetRequestID(c *ishell.Context) int64 {
	var requestID int64
	for {
		c.Print("RequestID : ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			requestID = tmp
			break
		} else {
			c.Println(err.Error())
		}
	}
	return requestID
}

func fnGetBody(c *ishell.Context) string {
	c.Print("Body: ")
	body := c.ReadLine()
	return body
}

func fnGetMaxID(c *ishell.Context) int64 {
	var maxID int64
	for {
		c.Print("Max ID: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			maxID = tmp
			break
		} else {
			c.Println(err.Error())
		}
	}
	return maxID
}

func fnGetMinID(c *ishell.Context) int64 {
	var minID int64
	for {
		c.Print("Min ID: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			minID = tmp
			break
		} else {
			c.Println(err.Error())
		}
	}
	return minID
}

func fnGetLimit(c *ishell.Context) int32 {
	var limit int32
	for {
		c.Print("Limit: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			limit = int32(tmp)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return limit
}

func fnGetTypingAction(c *ishell.Context) msg.TypingAction {
	var action msg.TypingAction
	for {
		c.Print("Action (0:Typing, 4:Cancel): ")
		actionID, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			action = msg.TypingAction(actionID)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return action
}

func fnGetMessageIDs(c *ishell.Context) []int64 {
	messagesIDs := make([]int64, 0)
	for {

		c.Print(len(messagesIDs), "Enter none numeric character to break\r\n")
		c.Print(len(messagesIDs), "MessageID: ")
		msgID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err != nil {
			break
		} else {
			messagesIDs = append(messagesIDs, msgID)
		}
	}
	return messagesIDs
}

func fnGetFromUpdateID(c *ishell.Context) int64 {
	var updateID int64
	for {
		c.Print("From UpdateID: ")
		fromUpdateID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			updateID = fromUpdateID
			break
		} else {
			c.Println(err.Error())
		}
	}
	return updateID
}

func fnGetInputUser(c *ishell.Context) []*msg.InputUser {
	users := make([]*msg.InputUser, 0)
	for {
		c.Print("Enter none numeric character to break\r\n")

		c.Print(len(users), "User ID: ")
		userID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err != nil {
			break
		}

		c.Print(len(users), "Access Hash: ")
		accessHash, err := strconv.ParseUint(c.ReadLine(), 10, 64)
		if err != nil {
			break
		}

		u := new(msg.InputUser)
		u.UserID = userID
		u.AccessHash = accessHash
		users = append(users, u)
	}
	return users
}

func fnGetUsername(c *ishell.Context) string {
	c.Print("Username: ")
	uname := c.ReadLine()
	return uname
}

func fnGetTitle(c *ishell.Context) string {
	c.Print("Title: ")
	title := c.ReadLine()
	return title
}
