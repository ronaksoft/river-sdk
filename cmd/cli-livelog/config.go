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
	ConfReportPort = "REPORT_PORT"
	ConfClient     = "CLIENT"
	ConfServerUrl  = "SERVER_URL"
	ConfPid        = "PID"
)

func readConfig() {
	RootCmd.Flags().Int("port", 2374, "listening port")
	RootCmd.Flags().Int("monitor_port", 2375, "monitoring port")
	RootCmd.Flags().Bool("client", false, "run in client mode")
	RootCmd.Flags().String("server_url", "https://livemon.ronaksoftware.com", "server address")
	RootCmd.Flags().String("pid", "", "")
	_ = viper.BindPFlag(ConfListenPort, RootCmd.Flags().Lookup("port"))
	_ = viper.BindPFlag(ConfReportPort, RootCmd.Flags().Lookup("report_port"))
	_ = viper.BindPFlag(ConfClient, RootCmd.Flags().Lookup("client"))
	_ = viper.BindPFlag(ConfServerUrl, RootCmd.Flags().Lookup("server_url"))
	_ = viper.BindPFlag(ConfPid, RootCmd.Flags().Lookup("pid"))
}
