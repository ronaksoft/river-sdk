package main

import (
	"fmt"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
)

func main() {
	readConfig()

	_ = RootCmd.Execute()
}

var RootCmd = &cobra.Command{
	Use: "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Server is running on: ", viper.GetInt(ConfListenPort))
		_ = fasthttp.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt(ConfListenPort)), func(ctx *fasthttp.RequestCtx) {
			fmt.Print(ronak.ByteToStr(ctx.Request.Body()))
		})
	},
}
