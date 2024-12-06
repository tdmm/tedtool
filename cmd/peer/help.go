package peer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

type BlockResponse struct {
	Result struct {
		BlockID struct {
			Hash string `json:"hash"`
		} `json:"block_id"`
	} `json:"result"`
}

// 定义结构体来解析 net_info 的 JSON 响应

// 获取 /status 返回的结果
func getStatus(endpoint string) (*StatusResponse, error) {
	url := fmt.Sprintf("%s/status", endpoint)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch /status from %s\n: %v", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %v", endpoint, err)
	}

	var status StatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON from %s: %v\n", endpoint, err)
	}

	return &status, nil
}

// 获取特定高度的区块信息
func getBlockByHeight(endpoint string, height string) (*BlockResponse, error) {
	url := fmt.Sprintf("%s/block?height=%s", endpoint, height)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch /block from %s: %v", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %v", endpoint, err)
	}

	var block BlockResponse
	if err := json.Unmarshal(body, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON from %s: %v", endpoint, err)
	}

	return &block, nil
}

// 获取 net_info 并提取 peers 的 endpoints
func getNetInfo(endpoint string) ([]Peers, error) {
	url := fmt.Sprintf("%s/net_info", endpoint)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch /net_info from %s: %v", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %v", endpoint, err)
	}

	var netInfo NetInfo
	if err := json.Unmarshal(body, &netInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON from %s: %v", endpoint, err)
	}

	//var endpoints []string
	//for _, peer := range netInfo.Result.Peers {
	//	// 构建 http 格式的 endpoint
	//	address := strings.Replace(peer.NodeInfo.ListenAddr, "tcp://", "http://", 1)
	//	endpoints = append(endpoints, address)
	//}

	return netInfo.Result.Peers, nil
}

// GetEndpoints 封装了整个逻辑，并发处理校验节点
func GetEndpoints(initEndpoint string) ([]string, error) {
	// 获取初始节点的 /status
	initialStatus, err := getStatus(initEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error fetching initial node status: %v", err)
	}

	initialChainID := initialStatus.Result.NodeInfo.Network
	initialBlockHeight := initialStatus.Result.SyncInfo.LatestBlockHeight
	initialBlockHash := initialStatus.Result.SyncInfo.LatestBlockHash

	fmt.Printf("Initial Node ChainID: %s, BlockHeight: %s, BlockHash: %s\n",
		initialChainID, initialBlockHeight, initialBlockHash)

	// 获取其他节点的 net_info
	endpoints, err := getNetInfo(initEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error fetching net_info: %v", err)
	}

	// 保存通过校验的节点
	var validEndpoints []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 并发校验每个节点
	for _, endpoint := range endpoints[:4] {
		wg.Add(1)
		go func(peer Peers) {
			defer wg.Done()

			endpoint := peer.NodeInfo.Other.RPCAddress

			if peer.NodeInfo.Network != initialChainID || peer.NodeInfo.Other.TxIndex != "on" {
				log.Printf("%s wrong", endpoint)
				return
			}
			if strings.Contains(endpoint, "127.0.0.1") {
				log.Printf("%s is 127", endpoint)
				return
			}

			endpoint = strings.ReplaceAll(endpoint, "tcp", "http")
			endpoint = strings.ReplaceAll(endpoint, "0.0.0.0", peer.RemoteIP)

			// 获取节点的 /status 信息
			status, err := getStatus(endpoint)
			if err != nil {
				log.Printf("Failed to fetch /status from %s: %v\n", endpoint, err)
				return
			}

			// 校验 chain_id
			if status.Result.NodeInfo.Network != initialChainID {
				fmt.Printf("ChainID mismatch for %s: %s (expected %s)\n", endpoint, status.Result.NodeInfo.Network, initialChainID)
				return
			}

			// 获取节点在指定高度的 block
			block, err := getBlockByHeight(endpoint, initialBlockHeight)
			if err != nil {
				log.Printf("Failed to fetch /block from %s: %v\n", endpoint, err)
				return
			}

			// 比较区块哈希
			if block.Result.BlockID.Hash == initialBlockHash {
				fmt.Printf("Block hash match for %s at height %s\n", endpoint, initialBlockHeight)
				// 加锁保存通过校验的 endpoint
				mu.Lock()
				validEndpoints = append(validEndpoints, endpoint)
				mu.Unlock()
			} else {
				fmt.Printf("Block hash mismatch for %s at height %s\n", endpoint, initialBlockHeight)
			}
		}(endpoint)
	}

	// 等待所有并发任务完成
	wg.Wait()

	// 返回通过校验的 endpoints
	return validEndpoints, nil
}
func getPeer(endpoint string) ([]string, error) {
	url := fmt.Sprintf("%s/net_info", endpoint)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch /net_info from %s: %v", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %v", endpoint, err)
	}

	var netInfo NetInfo
	if err := json.Unmarshal(body, &netInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON from %s: %v", endpoint, err)
	}

	var peers []string
	for _, peer := range netInfo.Result.Peers {
		// 防止崩溃的检查
		if peer.RemoteIP == "" {
			// 跳过没有 remote_ip 的 peer
			continue
		}

		// 检查 RPC 地址是否有效
		if peer.NodeInfo.Other.RPCAddress == "" {
			// 跳过没有 rpc_address 的 peer
			continue
		}

		// 确保 RPC 地址以 "tcp://" 开头，避免出错
		if !strings.HasPrefix(peer.NodeInfo.Other.RPCAddress, "tcp://") {
			continue
		}

		// 提取端口并构建 HTTP URL
		portIndex := strings.LastIndex(peer.NodeInfo.Other.RPCAddress, ":")
		if portIndex == -1 {
			continue
		}
		port := peer.NodeInfo.Other.RPCAddress[portIndex+1:]
		peerURL := fmt.Sprintf("http://%s:%s", peer.RemoteIP, port)

		peers = append(peers, peerURL)
	}

	return peers, nil
}
