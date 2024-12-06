/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package peer

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

// peerCmd represents the peer command
var PeerCmd = &cobra.Command{
	Use:   "peer",
	Short: "Get and validate peers from a given node URL",
	Long:  `This command fetches the peers from a given node URL and recursively checks for additional peer nodes.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取用户输入的 URL 参数
		url, _ := cmd.Flags().GetString("url")
		if url == "" {
			log.Fatal("URL is required")
		}

		// 初始化一个 URL 列表
		var urlList []string
		urlList = append(urlList, url)

		// 用来存储所有发现的 URL
		var allURLs []string
		// 存储已访问的 URL，避免重复访问
		visited := make(map[string]bool)

		for len(urlList) > 0 {
			// 获取当前 URL
			currentURL := urlList[0]
			urlList = urlList[1:]

			// 检查当前 URL 是否已访问
			if visited[currentURL] {
				continue
			}
			visited[currentURL] = true

			// 获取 net_info 中的 endpoints
			endpoints, err := getPeer(currentURL)
			if err != nil {
				log.Printf("Failed to fetch /net_info from %s: %v\n", currentURL, err)
				continue
			}

			// 输出当前节点的 peers 并将 URLs 添加到 allURLs 列表中
			fmt.Printf("Peers found at %s peers num:%d\n", currentURL, len(endpoints))
			for _, peer := range endpoints {
				allURLs = append(allURLs, peer)
			}
		}

		// 打印所有发现的 URLs
		fmt.Println("\nAll discovered URLs:")
		for _, u := range allURLs {
			fmt.Println(u)
		}

		fmt.Println("Discovery complete.")
	},
}

func init() {

	// 定义 URL 参数
	PeerCmd.PersistentFlags().String("url", "", "Initial node URL (required)")

	// 设置 URL 参数为必填
	PeerCmd.MarkPersistentFlagRequired("url")
}
