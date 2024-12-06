/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"time"
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start a proxy server",
	Long: `Start a proxy server that forwards requests to the specified Tendermint node
and limits the number of requests per second.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取用户传入的参数
		url, _ := cmd.Flags().GetString("url")
		port, _ := cmd.Flags().GetInt("port")
		limit, _ := cmd.Flags().GetInt("limit")

		// 校验参数
		if url == "" {
			log.Fatalf("Target URL is required. Use the --url flag to specify it.")
		}

		// 启动代理服务器
		proxyHandlerFunc(url, port, limit)
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	// 定义命令行标志
	proxyCmd.PersistentFlags().String("url", "", "proxy to tendermint url")
	proxyCmd.PersistentFlags().Int("port", 26657, "listen port")
	proxyCmd.PersistentFlags().Int("limit", 10, "maximum requests per second")
}

func proxyHandlerFunc(targetURL string, port int, limit int) {
	// 创建令牌桶
	tokens := make(chan struct{}, limit)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			// 每秒重新填充令牌桶
			for i := 0; i < limit-len(tokens); i++ {
				tokens <- struct{}{}
			}
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 从令牌桶中获取令牌（如果没有令牌，会阻塞直到有令牌可用）
		<-tokens

		// 构建请求
		req, err := http.NewRequest(r.Method, targetURL+r.URL.Path, r.Body)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			log.Printf("Error creating request: %v", err)
			return
		}

		// 设置请求头
		req.Header = r.Header
		req.Header.Set("Content-Type", "application/json")

		// 创建 HTTP 客户端
		client := &http.Client{}
		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to make proxy request", http.StatusInternalServerError)
			log.Printf("Error making proxy request: %v", err)
			return
		}
		defer resp.Body.Close()

		// 记录目标服务器的响应信息
		log.Printf("Response from target: %s, status: %d, duration: %v", targetURL, resp.StatusCode, time.Since(start))

		// 将响应写回客户端
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("Error copying response body: %v", err)
		}
	})

	serverAddr := fmt.Sprintf(":%d", port)
	log.Printf("Proxy server listening on %s, forwarding to %s, limit: %d requests/sec...", serverAddr, targetURL, limit)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("ListenAndServe failed: %v", err)
	}
}
