package gossip

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"coordinator/config"

	"github.com/hashicorp/memberlist"
)

type MessageType uint8

const (
	MsgWorkerUpdate  MessageType = 1
	MsgRebalanceReq  MessageType = 2
	MsgStateSync     MessageType = 3
)

type Message struct {
	Type    MessageType     `json:"type"`
	From    string          `json:"from"`
	Payload json.RawMessage `json:"payload"`
}

type WorkerUpdate struct {
	WorkerID string `json:"worker_id"`
	Action   string `json:"action"`
	Status   string `json:"status,omitempty"`
}

type RebalanceRequest struct {
	Reason string `json:"reason"`
}

type DelegateCallbacks struct {
	OnWorkerUpdate   func(update WorkerUpdate)
	OnRebalanceReq   func(reason string)
}

type delegate struct {
	mu        sync.RWMutex
	localMeta []byte
	msgCh     chan []byte
	callbacks *DelegateCallbacks
	logger    *slog.Logger
}

func (d *delegate) NodeMeta(limit int) []byte {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if len(d.localMeta) > limit {
		return nil
	}
	return d.localMeta
}

func (d *delegate) NotifyMsg(msg []byte) {
	if len(msg) == 0 {
		return
	}
	select {
	case d.msgCh <- msg:
	default:
		d.logger.Warn("gossip 消息通道已满, 丢弃消息")
	}
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}

func (d *delegate) LocalState(join bool) []byte {
	return nil
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
}

type GossipCoordinator struct {
	cfg      *config.Config
	list     *memberlist.Memberlist
	delegate *delegate
	msgCh    chan []byte
	logger   *slog.Logger

	mu    sync.RWMutex
	nodes map[string]*memberlist.Node
}

func New(cfg *config.Config, callbacks *DelegateCallbacks, logger *slog.Logger) *GossipCoordinator {
	msgCh := make(chan []byte, 256)
	d := &delegate{
		localMeta: []byte(cfg.ZoneID),
		msgCh:     msgCh,
		callbacks: callbacks,
		logger:    logger,
	}

	return &GossipCoordinator{
		cfg:      cfg,
		delegate: d,
		msgCh:    msgCh,
		logger:   logger.With("component", "gossip"),
		nodes:    make(map[string]*memberlist.Node),
	}
}

func (g *GossipCoordinator) Start(ctx context.Context) error {
	mlConfig := memberlist.DefaultLANConfig()
	mlConfig.BindPort = g.cfg.Gossip.BindPort
	mlConfig.AdvertisePort = g.cfg.Gossip.BindPort
	mlConfig.Name = g.cfg.Gossip.NodeName
	mlConfig.Delegate = g.delegate

	list, err := memberlist.Create(mlConfig)
	if err != nil {
		return fmt.Errorf("创建 memberlist 失败: %w", err)
	}
	g.list = list

	if len(g.cfg.Gossip.Seeds) > 0 {
		n, err := list.Join(g.cfg.Gossip.Seeds)
		if err != nil {
			g.logger.Warn("gossip 加入种子节点失败, 将以独立模式运行", "seeds", g.cfg.Gossip.Seeds, "error", err)
		} else {
			g.logger.Info("gossip 已加入集群", "joined", n)
		}
	}

	go g.messageLoop(ctx)
	go g.memberSync(ctx)

	g.logger.Info("gossip 已启动", "port", g.cfg.Gossip.BindPort, "node", g.cfg.Gossip.NodeName)
	return nil
}

func (g *GossipCoordinator) Stop() error {
	if g.list != nil {
		return g.list.Shutdown()
	}
	return nil
}

func (g *GossipCoordinator) messageLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-g.msgCh:
			g.handleMessage(msg)
		}
	}
}

func (g *GossipCoordinator) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		g.logger.Warn("gossip 消息解析失败", "error", err)
		return
	}

	switch msg.Type {
	case MsgWorkerUpdate:
		var update WorkerUpdate
		if err := json.Unmarshal(msg.Payload, &update); err != nil {
			return
		}
		if g.delegate.callbacks != nil && g.delegate.callbacks.OnWorkerUpdate != nil {
			g.delegate.callbacks.OnWorkerUpdate(update)
		}

	case MsgRebalanceReq:
		var req RebalanceRequest
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			return
		}
		if g.delegate.callbacks != nil && g.delegate.callbacks.OnRebalanceReq != nil {
			g.delegate.callbacks.OnRebalanceReq(req.Reason)
		}
	}
}

func (g *GossipCoordinator) memberSync(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.syncMembers()
		}
	}
}

func (g *GossipCoordinator) syncMembers() {
	if g.list == nil {
		return
	}
	members := g.list.Members()
	g.mu.Lock()
	g.nodes = make(map[string]*memberlist.Node, len(members))
	for _, m := range members {
		g.nodes[m.Name] = m
	}
	g.mu.Unlock()
}

func (g *GossipCoordinator) BroadcastWorkerUpdate(update WorkerUpdate) {
	payload, _ := json.Marshal(update)
	msg := Message{
		Type:    MsgWorkerUpdate,
		From:    g.cfg.Gossip.NodeName,
		Payload: payload,
	}
	data, _ := json.Marshal(msg)
	g.list.SendReliable(nil, data)
}

func (g *GossipCoordinator) BroadcastRebalance(reason string) {
	payload, _ := json.Marshal(RebalanceRequest{Reason: reason})
	msg := Message{
		Type:    MsgRebalanceReq,
		From:    g.cfg.Gossip.NodeName,
		Payload: payload,
	}
	data, _ := json.Marshal(msg)
	g.list.SendReliable(nil, data)
}

func (g *GossipCoordinator) AliveMembers() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var names []string
	for name := range g.nodes {
		names = append(names, name)
	}
	return names
}

func (g *GossipCoordinator) MemberCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}
