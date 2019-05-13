package main

import (
	"encoding/json"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/shared"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/supernumerary"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/config"
	"github.com/nats-io/go-nats"
	"go.uber.org/zap"
)

// Node supernumerary client
type Node struct {
	Config     *config.NodeConfig
	su         *supernumerary.Supernumerary
	natsClient *nats.Conn
	subs       map[string]*nats.Subscription
	StartPhone int64
	EndPhone   int64
}

// NewNode create supernumerary new client
func NewNode(cfg *config.NodeConfig) (*Node, error) {
	n := &Node{
		Config: cfg,
		su:     nil,
		subs:   make(map[string]*nats.Subscription),
	}
	natsClient, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return nil, err
	}
	n.natsClient = natsClient

	err = n.RegisterSubscription()
	if err != nil {
		return nil, err
	}

	go func() {
		cmd := config.NodeRegisterCmd{
			InstanceID: cfg.InstanceID,
		}
		cmdBytes, _ := json.Marshal(cmd)
		for {
			_ = natsClient.Publish(config.SubjectCommander, cmdBytes)
			time.Sleep(10 * time.Second)
		}
	}()


	return n, nil
}

func (n *Node) cbStart(msg *nats.Msg) {
	_Log.Info("cbStart()")
	cfg := config.StartCfg{}
	err := json.Unmarshal(msg.Data, &cfg)
	if err != nil {
		_Log.Error("Failed to unmarshal StartCfg", zap.Error(err))
		return
	}

	// check start state
	if n.su != nil {
		if len(n.su.Actors) > 0 {
			_Log.Error("cbStart() supernumerary already started")
			return
		}
	}

	shared.DefaultFileServerURL = cfg.FileServerURL
	shared.DefaultServerURL = cfg.ServerURL
	shared.DefaultTimeout = cfg.Timeout
	shared.DefaultSendTimeout = cfg.Timeout

	su, err := supernumerary.NewSupernumerary(n.StartPhone, n.EndPhone)
	if err != nil {
		_Log.Error("cbStart()", zap.Error(err))
	}
	n.su = su
}

func (n *Node) cbStop(msg *nats.Msg) {
	_Log.Info("cbStop()")

	if n.su == nil {
		_Log.Error("cbStop() supernumerary not initialized")
		return
	}

	n.su.Stop()
}

func (n *Node) cbCreateAuthKey(msg *nats.Msg) {
	_Log.Info("cbCreateAuthKey()")

	if n.su == nil {
		_Log.Error("cbCreateAuthKey() supernumerary not initialized")
		return
	}
	n.su.CreateAuthKey()
}

func (n *Node) cbLogin(msg *nats.Msg) {
	_Log.Info("cbLogin()")

	if n.su == nil {
		_Log.Error("cbLogin() supernumerary not initialized")
		return
	}
	n.su.Login()
}

func (n *Node) cbRegister(msg *nats.Msg) {
	_Log.Info("cbRegister()")
	if n.su == nil {
		_Log.Error("cbRegister() supernumerary not initialized")
		return
	}
	n.su.Register()
}

func (n *Node) cbTicker(msg *nats.Msg) {
	if n.su == nil {
		_Log.Error("cbTicker() supernumerary not initialized")
		return
	}

	_Log.Info("cbTicker()", zap.String("Data", string(msg.Data)))

	cfg := config.TickerCfg{}
	err := json.Unmarshal(msg.Data, &cfg)
	if err != nil {
		_Log.Error("cbTicker() failed to unmarshal", zap.Error(err))
		return
	}

	n.su.SetTickerApplier(cfg.Duration, cfg.Action)
}

func (n *Node) cbPhoneRange(msg *nats.Msg) {
	_Log.Info("cbPhoneRange()")
	cfg := config.PhoneRangeCfg{}
	err := json.Unmarshal(msg.Data, &cfg)
	if err != nil {
		_Log.Error("cbPhoneRange() failed to unmarshal", zap.Error(err))
		return
	}
	n.StartPhone = cfg.StartPhone
	n.EndPhone = cfg.EndPhone
	_ = n.natsClient.Publish(msg.Reply, []byte("OK"))
}

func (n *Node) cbHealthCheck(msg *nats.Msg) {
	_Log.Info("cbHealthCheck()")
	_ = n.natsClient.Publish(msg.Reply, []byte("OK"))
}

// RegisterSubscription subscribe subjects
func (n *Node) RegisterSubscription() error {
	subStart, err := n.natsClient.Subscribe(config.SubjectStart, n.cbStart)
	if err != nil {
		return err
	}
	n.subs[config.SubjectStart] = subStart

	subStop, err := n.natsClient.Subscribe(config.SubjectStop, n.cbStop)
	if err != nil {
		return err
	}
	n.subs[config.SubjectStop] = subStop

	subCreateAuthKey, err := n.natsClient.Subscribe(config.SubjectCreateAuthKey, n.cbCreateAuthKey)
	if err != nil {
		return err
	}
	n.subs[config.SubjectCreateAuthKey] = subCreateAuthKey

	subLogin, err := n.natsClient.Subscribe(config.SubjectLogin, n.cbLogin)
	if err != nil {
		return err
	}
	n.subs[config.SubjectLogin] = subLogin

	subRegister, err := n.natsClient.Subscribe(config.SubjectRegister, n.cbRegister)
	if err != nil {
		return err
	}
	n.subs[config.SubjectRegister] = subRegister

	subTicker, err := n.natsClient.Subscribe(config.SubjectTicker, n.cbTicker)
	if err != nil {
		return err
	}
	n.subs[config.SubjectTicker] = subTicker

	subPhoneRange, err := n.natsClient.Subscribe(fmt.Sprintf("%s.%s", n.Config.InstanceID, config.SubjectPhoneRange), n.cbPhoneRange)
	if err != nil {
		return err
	}
	n.subs[fmt.Sprintf("%s.%s", n.Config.InstanceID, config.SubjectPhoneRange)] = subPhoneRange

	subHealthCheck, err := n.natsClient.Subscribe(fmt.Sprintf("%s.%s", n.Config.InstanceID, config.SubjectHealthCheck), n.cbHealthCheck)
	if err != nil {
		return err
	}
	n.subs[fmt.Sprintf("%s.%s", n.Config.InstanceID, config.SubjectHealthCheck)] = subHealthCheck
	return nil
}

// Unsubscribe unsubscribe subjects
func (n *Node) Unsubscribe() {
	for _, s := range n.subs {
		s.Unsubscribe()
	}
}
