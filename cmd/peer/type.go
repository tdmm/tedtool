package peer

import "time"

type NetInfo struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  Result `json:"result"`
}
type ProtocolVersion struct {
	P2P   string `json:"p2p"`
	Block string `json:"block"`
	App   string `json:"app"`
}
type Other struct {
	TxIndex    string `json:"tx_index"`
	RPCAddress string `json:"rpc_address"`
}
type NodeInfo struct {
	ProtocolVersion ProtocolVersion `json:"protocol_version"`
	ID              string          `json:"id"`
	ListenAddr      string          `json:"listen_addr"`
	Network         string          `json:"network"`
	Version         string          `json:"version"`
	Channels        string          `json:"channels"`
	Moniker         string          `json:"moniker"`
	Other           Other           `json:"other"`
}
type SendMonitor struct {
	Start    time.Time `json:"Start"`
	Bytes    string    `json:"Bytes"`
	Samples  string    `json:"Samples"`
	InstRate string    `json:"InstRate"`
	CurRate  string    `json:"CurRate"`
	AvgRate  string    `json:"AvgRate"`
	PeakRate string    `json:"PeakRate"`
	BytesRem string    `json:"BytesRem"`
	Duration string    `json:"Duration"`
	Idle     string    `json:"Idle"`
	TimeRem  string    `json:"TimeRem"`
	Progress int       `json:"Progress"`
	Active   bool      `json:"Active"`
}
type RecvMonitor struct {
	Start    time.Time `json:"Start"`
	Bytes    string    `json:"Bytes"`
	Samples  string    `json:"Samples"`
	InstRate string    `json:"InstRate"`
	CurRate  string    `json:"CurRate"`
	AvgRate  string    `json:"AvgRate"`
	PeakRate string    `json:"PeakRate"`
	BytesRem string    `json:"BytesRem"`
	Duration string    `json:"Duration"`
	Idle     string    `json:"Idle"`
	TimeRem  string    `json:"TimeRem"`
	Progress int       `json:"Progress"`
	Active   bool      `json:"Active"`
}
type Channels struct {
	ID                int    `json:"ID"`
	SendQueueCapacity string `json:"SendQueueCapacity"`
	SendQueueSize     string `json:"SendQueueSize"`
	Priority          string `json:"Priority"`
	RecentlySent      string `json:"RecentlySent"`
}
type ConnectionStatus struct {
	Duration    string      `json:"Duration"`
	SendMonitor SendMonitor `json:"SendMonitor"`
	RecvMonitor RecvMonitor `json:"RecvMonitor"`
	Channels    []Channels  `json:"Channels"`
}
type Peers struct {
	NodeInfo         NodeInfo         `json:"node_info"`
	IsOutbound       bool             `json:"is_outbound"`
	ConnectionStatus ConnectionStatus `json:"connection_status"`
	RemoteIP         string           `json:"remote_ip"`
}
type Result struct {
	Listening bool     `json:"listening"`
	Listeners []string `json:"listeners"`
	NPeers    string   `json:"n_peers"`
	Peers     []Peers  `json:"peers"`
}

type StatusResponse struct {
	Jsonrpc string  `json:"jsonrpc"`
	ID      int     `json:"id"`
	Result  Result2 `json:"result"`
}
type ProtocolVersion2 struct {
	P2P   string `json:"p2p"`
	Block string `json:"block"`
	App   string `json:"app"`
}
type Other2 struct {
	TxIndex    string `json:"tx_index"`
	RPCAddress string `json:"rpc_address"`
}
type NodeInfo2 struct {
	ProtocolVersion ProtocolVersion2 `json:"protocol_version"`
	ID              string           `json:"id"`
	ListenAddr      string           `json:"listen_addr"`
	Network         string           `json:"network"`
	Version         string           `json:"version"`
	Channels        string           `json:"channels"`
	Moniker         string           `json:"moniker"`
	Other           Other2           `json:"other"`
}
type SyncInfo struct {
	LatestBlockHash     string    `json:"latest_block_hash"`
	LatestAppHash       string    `json:"latest_app_hash"`
	LatestBlockHeight   string    `json:"latest_block_height"`
	LatestBlockTime     time.Time `json:"latest_block_time"`
	EarliestBlockHash   string    `json:"earliest_block_hash"`
	EarliestAppHash     string    `json:"earliest_app_hash"`
	EarliestBlockHeight string    `json:"earliest_block_height"`
	EarliestBlockTime   time.Time `json:"earliest_block_time"`
	CatchingUp          bool      `json:"catching_up"`
}
type PubKey struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type ValidatorInfo struct {
	Address     string `json:"address"`
	PubKey      PubKey `json:"pub_key"`
	VotingPower string `json:"voting_power"`
}
type Result2 struct {
	NodeInfo      NodeInfo2     `json:"node_info"`
	SyncInfo      SyncInfo      `json:"sync_info"`
	ValidatorInfo ValidatorInfo `json:"validator_info"`
}
