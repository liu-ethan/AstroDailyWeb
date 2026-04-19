package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client interface {
	GenerateTodayFortune(ctx context.Context, email string) (string, error)
}

type OpenAICompatibleClient struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
}

// NewOpenAICompatibleClient 创建兼容 OpenAI 协议的 LLM 客户端。
// 参数：apiKey - API 密钥；baseURL - 接口地址；model - 模型名称；timeout - 请求超时。
// 返回：*OpenAICompatibleClient - LLM 客户端实例。
func NewOpenAICompatibleClient(apiKey, baseURL, model string, timeout time.Duration) *OpenAICompatibleClient {
	return &OpenAICompatibleClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		http:    &http.Client{Timeout: timeout},
	}
}

// GenerateTodayFortune 调用大模型接口生成今日运势。
// 参数：ctx - 上下文；email - 用户标识（当前使用邮箱）。
// 返回：string - 运势文本；error - 调用或解析失败错误。
func (c *OpenAICompatibleClient) GenerateTodayFortune(ctx context.Context, email string) (string, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return "", fmt.Errorf("LLM API config missing")
	}

	payload := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": "你是占星助手，请输出简洁的今日运势。"},
			{"role": "user", "content": "请为用户 " + email + " 生成今日运势。"},
		},
	}
	buf, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm status=%d", resp.StatusCode)
	}

	// 当前阶段仅提供通用适配，后续再补充具体响应体解析。
	return "", fmt.Errorf("LLM response parser not implemented")
}

type StubClient struct{}

// GenerateTodayFortune 作为未接入模型时的兜底实现。
// 参数：ctx - 上下文；email - 用户标识。
// 返回：string - 为空；error - 未实现错误。
func (s *StubClient) GenerateTodayFortune(ctx context.Context, email string) (string, error) {
	_ = ctx
	_ = email
	return "", fmt.Errorf("LLM client not implemented")
}
