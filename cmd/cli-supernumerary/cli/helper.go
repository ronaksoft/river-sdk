package main

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/config"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/abiosoft/ishell.v2"
)

type Params struct {
	ServerUrl     string
	FileServerUrl string
}

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
	defaultServerUrl, _ := ioutil.ReadFile(".river-server")
	if defaultServerUrl == nil {
		defaultServerUrl = ronak.StrToByte("test.river.im")
	}
	c.Print(fmt.Sprintf("Server URL (ws://%s): ", defaultServerUrl))
	tmp := c.ReadLine()
	if tmp == "" {
		tmp = ronak.ByteToStr(defaultServerUrl)
	} else {
		tmp = strings.TrimPrefix(tmp, "ws://")
		_ = ioutil.WriteFile(".river-server", ronak.StrToByte(tmp), os.ModePerm)
	}
	return fmt.Sprintf("ws://%s", tmp)
}

func fnGetFileServerURL(c *ishell.Context) string {
	defaultServerUrl, _ := ioutil.ReadFile(".river-server")
	if defaultServerUrl == nil {
		defaultServerUrl = ronak.StrToByte("test.river.im")
	}
	c.Print(fmt.Sprintf("File Server URL (http://%s/file): ", defaultServerUrl))
	tmp := c.ReadLine()
	if tmp == "" {
		tmp = fmt.Sprintf("http://%s/file", ronak.ByteToStr(defaultServerUrl))
	}
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

func fnGetMaxInterval(c *ishell.Context) time.Duration {
	var duration time.Duration
	for {
		c.Print("MaxInterval (second): ")
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

func healthCheck(c *ishell.Context) {
	instanceIDs := make([]string, 0)
	_NodesLock.RLock()
	c.Println("Checking all node are alive")
	for instanceID := range _Nodes {
		instanceIDs = append(instanceIDs, instanceID)
	}
	_NodesLock.RUnlock()

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(instanceIDs))
	for _, instanceID := range instanceIDs {
		go func(instanceID string) {
			defer waitGroup.Done()
			_, err := _NATS.Request(fmt.Sprintf("%s.%s", instanceID, config.SubjectHealthCheck), []byte("HEALTH_CHECK"), 5*time.Second)
			if err != nil {
				_NodesLock.Lock()
				delete(_Nodes, instanceID)
				_NodesLock.Unlock()
			}

		}(instanceID)
	}
	waitGroup.Wait()
}
