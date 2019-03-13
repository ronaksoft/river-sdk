package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/supernumerary"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/supernumerary/config"
	"github.com/nats-io/go-nats"
	"go.uber.org/zap"
)

// Node supernumerary client
type Node struct {
	Config *config.NodeConfig
	su     *supernumerary.Supernumerary
	nats   *nats.Conn
	subs   map[string]*nats.Subscription
}

// NewNode create supernumerary new client
func NewNode(cfg *config.NodeConfig) (*Node, error) {
	n := &Node{
		Config: cfg,
		su:     nil,
		subs:   make(map[string]*nats.Subscription),
	}
	nats, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return nil, err
	}

	n.nats = nats
	return n, nil
}

func (n *Node) cbStart(msg *nats.Msg) {
	cfg := config.StartCfg{}
	err := json.Unmarshal(msg.Data, &cfg)
	if err != nil {
		logs.Error("Failed to unmarshal SatrtCfg", zap.Error(err))
		return
	}

	shared.DefaultFileServerURL = cfg.FileServerURL
	shared.DefaultServerURL = cfg.ServerURL
	shared.DefaultTimeout = cfg.Timeout
	shared.DefaultSendTimeout = cfg.Timeout

	n.RegisterSubscribtion()

	su, err := supernumerary.NewSupernumerary(n.Config.StartPhone, n.Config.StartPhone)
	if err != nil {
		logs.Error("cbStart()", zap.Error(err))
	}
	n.su = su
}

func (n *Node) cbStop(msg *nats.Msg) {
	if n.su == nil {
		logs.Error("cbStop() supernumerary not initialized")
		return
	}

	n.su.Stop()
	n.Unsubscribe()
}

func (n *Node) cbCreateAuthKey(msg *nats.Msg) {
	if n.su == nil {
		logs.Error("cbCreateAuthKey() supernumerary not initialized")
		return
	}
	n.su.CreateAuthKey()
}

func (n *Node) cbLogin(msg *nats.Msg) {
	if n.su == nil {
		logs.Error("cbLogin() supernumerary not initialized")
		return
	}
	n.su.Login()
}

func (n *Node) cbRegister(msg *nats.Msg) {
	if n.su == nil {
		logs.Error("cbRegister() supernumerary not initialized")
		return
	}
	n.su.Register()
}

func (n *Node) cbTicker(msg *nats.Msg) {
	data := strings.Split(string(msg.Data), ":")
	if len(data) == 2 {
		duration, err := strconv.Atoi(data[0])
		if err == nil {
			logs.Error("cbTicker() failed to parse", zap.Error(err))
			return
		}
		action, err := strconv.Atoi(data[1])
		if err == nil {
			logs.Error("cbTicker() failed to parse", zap.Error(err))
			return
		}
		n.su.SetTickerApplier(time.Duration(duration), supernumerary.TickerAction(action))

	} else {
		logs.Error("cbTicker() invalid parameters", zap.String("data", string(msg.Data)))
	}
}

// RegisterSubscribtion subscribe subjects
func (n *Node) RegisterSubscribtion() error {
	subStart, err := n.nats.Subscribe(config.SUBJECT_START, n.cbStart)
	if err != nil {
		return err
	}
	n.subs[config.SUBJECT_START] = subStart

	subStop, err := n.nats.Subscribe(config.SUBJECT_STOP, n.cbStop)
	if err != nil {
		return err
	}
	n.subs[config.SUBJECT_STOP] = subStop

	subCreateAuthKey, err := n.nats.Subscribe(config.SUBJECT_CREATEAUTHKEY, n.cbCreateAuthKey)
	if err != nil {
		return err
	}
	n.subs[config.SUBJECT_CREATEAUTHKEY] = subCreateAuthKey

	subLogin, err := n.nats.Subscribe(config.SUBJECT_LOGIN, n.cbLogin)
	if err != nil {
		return err
	}
	n.subs[config.SUBJECT_LOGIN] = subLogin

	subRegister, err := n.nats.Subscribe(config.SUBJECT_RIGISTER, n.cbRegister)
	if err != nil {
		return err
	}
	n.subs[config.SUBJECT_RIGISTER] = subRegister

	subTicker, err := n.nats.Subscribe(config.SUBJECT_TICKER, n.cbTicker)
	if err != nil {
		return err
	}
	n.subs[config.SUBJECT_TICKER] = subTicker

	return nil
}

// Unsubscribe unsubscribe subjects
func (n *Node) Unsubscribe() {
	for _, s := range n.subs {
		s.Unsubscribe()
	}
}
