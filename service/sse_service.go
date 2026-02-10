package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func SendSSE(w http.ResponseWriter, f http.Flusher, event string, data map[string]any) {
	// 1. 将数据转换为 JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		// 处理错误，可以记录日志或发送错误消息
		fmt.Fprintf(w, "event: error\ndata: {\"error\":\"json encode failed\"}\n\n")
		f.Flush()
		return
	}

	// 2. 发送 SSE 格式数据
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonBytes) // jsonBytes 是 []byte，自动转换为 string

	// 3. 立即刷新
	f.Flush()
}
