package main

import (
	"fmt"
	"strconv"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

func fnStartPhone(c *ishell.Context) int64 {
	var tmpNo int64
	for {

		c.Print("Start Phone: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			if tmp < 10000000 {
				tmpNo = tmp
				break
			}
			c.Println("max 7 digit allowed")
		} else {
			c.Println(err.Error())
		}
	}
	return tmpNo
}

func fnEndPhone(c *ishell.Context) int64 {
	var tmpNo int64
	for {

		c.Print("End Phone: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 64)
		if err == nil {
			if tmp < 10000000 {
				tmpNo = tmp
				break
			}
			c.Println("max 7 digit allowed")
		} else {
			c.Println(err.Error())
		}
	}
	return tmpNo
}

func fnGetPhone(c *ishell.Context) string {
	var tmpNo string
	for {

		c.Print("Phone: ")
		tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		if err == nil {
			tmpNo = strconv.Itoa(int(tmp))
			break
		} else {
			c.Println(err.Error())
		}
	}
	return tmpNo
}

func fnClearScreeen() {
	fmt.Println("\033[H\033[2J") // clear screen
}
