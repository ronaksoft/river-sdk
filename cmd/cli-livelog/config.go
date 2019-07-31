package main

import "github.com/spf13/viper"

/*
   Creation Time: 2019 - Jun - 04
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

const (
	ConfListenPort = "LISTEN_PORT"
)

func readConfig() {
	RootCmd.Flags().Int("port", 2374, "listen port")
	_ = viper.BindPFlag(ConfListenPort, RootCmd.Flags().Lookup("port"))
}
