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
	Short: "",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取用户传入的参数
		url, _ := cmd.Flags().GetString("url")
		port, _ := cmd.Flags().GetInt("port")

		// 校验参数
		if url == "" {
			log.Fatalf("Target URL is required. Use the --url flag to specify it.")
		}

		// 启动代理服务器
		proxyHandlerFunc(url, port)
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	proxyCmd.PersistentFlags().String("url", "", "proxy to tendermint url")
	proxyCmd.PersistentFlags().Int("port", 26657, "listen port")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// proxyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func proxyHandlerFunc(targetURL string, port int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
	log.Printf("Proxy server listening on %s, forwarding to %s...", serverAddr, targetURL)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("ListenAndServe failed: %v", err)
	}
}
