package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var proxysCmd = &cobra.Command{
	Use:   "proxys",
	Short: "Start a proxy server with URL fallback support",
	Long: `Start a proxy server that forwards requests to a list of URLs specified in a file.
If the current URL fails (returns non-200 status), it switches to the next URL.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取用户传入的参数
		file, _ := cmd.Flags().GetString("file")
		port, _ := cmd.Flags().GetInt("port")
		limit, _ := cmd.Flags().GetInt("limit")

		// 从文件中读取 URL 列表
		urls, err := readURLsFromFile(file)
		if err != nil {
			log.Fatalf("Failed to read URLs from file: %v", err)
		}
		if len(urls) == 0 {
			log.Fatalf("No URLs found in file: %s", file)
		}

		// 启动代理服务器
		proxyWithFallbackHandlerFunc(urls, port, limit)
	},
}

func init() {
	rootCmd.AddCommand(proxysCmd)

	// 定义命令行标志
	proxysCmd.PersistentFlags().String("file", "urls", "file containing list of URLs")
	proxysCmd.PersistentFlags().Int("port", 26657, "listen port")
	proxysCmd.PersistentFlags().Int("limit", 10, "maximum requests per second")
}

func proxyWithFallbackHandlerFunc(urls []string, port int, limit int) {
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

	// 使用一个变量记录当前的目标 URL 索引
	currentURLIndex := 0
	errorRecords := make([]time.Time, 0) // 存储错误记录的时间

	// 定期检查过去 20 秒内是否有错误，并切换 URL
	go func() {
		for range time.Tick(30 * time.Second) {
			// 记录过去 20 秒内的错误
			threshold := time.Now().Add(-20 * time.Second)
			var recentErrors []time.Time
			for _, record := range errorRecords {
				if record.After(threshold) {
					recentErrors = append(recentErrors, record)
				}
			}

			// 如果有错误记录，则切换 URL
			if len(recentErrors) > 0 {
				log.Printf("Detected errors in the past 20 seconds, switching to the next URL.")
				currentURLIndex = (currentURLIndex + 1) % len(urls)
				errorRecords = []time.Time{} // 清空错误记录
			}
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 从令牌桶中获取令牌（如果没有令牌，会阻塞直到有令牌可用）
		<-tokens

		targetURL := urls[currentURLIndex]

		// 构建请求
		req, err := http.NewRequest(r.Method, targetURL+r.URL.Path, r.Body)
		if err != nil {
			log.Printf("Error creating request for URL %s: %v", targetURL, err)
			// 切换到下一个 URL
			currentURLIndex = (currentURLIndex + 1) % len(urls)
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
			log.Printf("Error making proxy request to URL %s: %v", targetURL, err)
			// 切换到下一个 URL
			errorRecords = append(errorRecords, time.Now())
			return
		}
		defer resp.Body.Close()

		// 检查响应状态码
		if resp.StatusCode != http.StatusOK {
			log.Printf("Non-200 response from URL %s: %d", targetURL, resp.StatusCode)
			// 记录错误时间
			errorRecords = append(errorRecords, time.Now())
			return
		}

		// 记录目标服务器的响应信息
		log.Printf("Response from target: %s, status: %d, duration: %v", targetURL, resp.StatusCode, time.Since(start))

		// 将响应写回客户端
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("Error copying response body: %v", err)
		}
		return
	})

	serverAddr := fmt.Sprintf(":%d", port)
	log.Printf("Proxy server with fallback listening on %s, limit: %d requests/sec...", serverAddr, limit)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("ListenAndServe failed: %v", err)
	}
}

func readURLsFromFile(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var urls []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urls = append(urls, url)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}
