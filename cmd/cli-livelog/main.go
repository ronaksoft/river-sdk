package main

import (
	"context"
	"fmt"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	mtx   sync.RWMutex
	pipes = map[string]*io.PipeWriter{}
)

func getPipe(k string) *io.PipeWriter {
	mtx.RLock()
	p := pipes[k]
	mtx.RUnlock()
	return p
}

func setPipe(k string, p *io.PipeWriter) {
	mtx.Lock()
	pipes[k] = p
	mtx.Unlock()
}

func delPipe(k string) {
	mtx.Lock()
	delete(pipes, k)
	mtx.Unlock()
}

func main() {
	readConfig()

	_ = RootCmd.Execute()
}

var RootCmd = &cobra.Command{
	Use: "",
	Run: func(cmd *cobra.Command, args []string) {

		if viper.GetBool(ConfClient) {
			// Run Client Mode
			fmt.Println("Client is connecting to: ", viper.GetInt(ConfServerUrl))
			runClient()

		} else {
			fmt.Println("Server is running on: ", viper.GetInt(ConfListenPort))
			runServer()
		}

	},
}

func runClient() {
	ctx, cf := context.WithTimeout(context.Background(), time.Second*5)
	defer cf()
	conn, _, _, err := ws.Dial(ctx, fmt.Sprintf("%s/%s", strings.TrimRight(viper.GetString(ConfServerUrl), "/"), viper.GetString(ConfPid)))
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		data, err := wsutil.ReadServerText(conn)
		if err != nil {
			break
		}
		fmt.Print(string(data))
	}
}

func runServer() {
	// Run Server Mode
	// 1. Run Http Log Inputs
	go func() {
		err := fasthttp.ListenAndServe(
			fmt.Sprintf(":%d", viper.GetInt(ConfListenPort)),
			func(ctx *fasthttp.RequestCtx) {
				pid := strings.TrimLeft(domain.ByteToStr(ctx.Request.RequestURI()), "/")
				p := getPipe(pid)
				if p == nil {
					return
				}

				_, err := p.Write(ctx.Request.Body())
				if err != nil {
					delPipe(pid)
				}
			},
		)
		if err != nil {
			fmt.Println("Error on Http Server:", err)
		}
	}()

	// 2. Run Websocket Log Printer
	go func() {
		err := http.ListenAndServe(
			fmt.Sprintf(":%d", viper.GetInt(ConfMonitorPort)),
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, _, _, err := ws.UpgradeHTTP(r, w)
				if err != nil {
					// handle error
				}
				pid := strings.TrimLeft(r.RequestURI, "/")
				go func(pid string) {
					defer conn.Close()

					pr, pw := io.Pipe()
					buf := make([]byte, 4096)
					setPipe(pid, pw)
					for {
						n, err := pr.Read(buf)
						if err != nil {
							fmt.Println(err)
							break
						}
						err = wsutil.WriteServerMessage(conn, ws.OpText, buf[:n])
						if err != nil {
							fmt.Println(err)
							break
						}
					}

					_ = pr.Close()
					_ = pw.Close()
					delPipe(pid)
				}(pid)
			}),
		)
		if err != nil {
			fmt.Println("Error on Websocket Server:", err)
		}

	}()

	select {}
}
