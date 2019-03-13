package main

import (
	"strconv"
	"time"

	"gopkg.in/abiosoft/ishell.v2"
)

func fnGetDuration(c *ishell.Context) time.Duration {
	var tmpNo time.Duration
	for {
		c.Print("Duration (second): ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			tmpNo = time.Duration(tmp)
			break
		} else {
			c.Println(err.Error())
		}
	}
	return tmpNo * time.Second
}

func fnGetTickerAction(c *ishell.Context) int {
	var action int
	for {
		c.Print("Actions (SendMsg = 1 , SendFile = 2) : ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			if tmp > 0 && tmp < 3 {
				action = int(tmp)
				break
			}
		}
	}
	return action
}

func fnGetServerURL(c *ishell.Context) string {
	c.Print("Server URL: ")
	tmp := c.ReadLine()
	return tmp
}

func fnGetFileServerURL(c *ishell.Context) string {
	c.Print("File Server URL: ")
	tmp := c.ReadLine()
	return tmp
}

func fnGetTimeout(c *ishell.Context) time.Duration {
	var duration time.Duration
	for {
		c.Print("Timeout (second): ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			duration = time.Duration(tmp) * time.Second
			break
		} else {
			c.Println(err.Error())
		}
	}
	return duration
}
