// +build ignore

package main

import (
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

/*
   Creation Time: 2019 - Nov - 30
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func main() {
	f, err := os.Create("version.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var riverVer, gitVer string
	b, err := exec.Command("git", "describe", "--abbrev=0").Output()
	riverVer = strings.TrimSpace(string(b))
	b, err = exec.Command("git", "log", "--format=\"%H\"", "-n", "1").Output()
	gitVer = strings.TrimSpace(string(b))

	packageTemplate.Execute(f, struct {
		Timestamp    time.Time
		GitVersion   string
		RiverVersion string
	}{
		Timestamp:    time.Now(),
		GitVersion:   gitVer,
		RiverVersion: riverVer,
	})
}

var packageTemplate = template.Must(template.New("").Parse(`
// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at
// {{ .Timestamp }}
package domain

var (
	GitVersion = {{ .GitVersion }}
	SDKVersion = "{{ .RiverVersion }}"
)
`))