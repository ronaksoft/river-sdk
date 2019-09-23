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
			// if strings.Contains(ronak.ByteToStr(ctx.Request.Body()), "Pending Message") ||
			// 	strings.Contains(ronak.ByteToStr(ctx.Request.Body()), "updateMessageID") ||
			// 	strings.Contains(ronak.ByteToStr(ctx.Request.Body()), "UpdateHandler() -> UpdateAppliers") ||
			// 	strings.Contains(ronak.ByteToStr(ctx.Request.Body()), "SyncController::updateNewMessage") {
			// 	// fmt.Print(ronak.ByteToStr(ctx.Request.Body()))
			// }
		})
	},
}
