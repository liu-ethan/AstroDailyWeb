package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type FortuneProfile struct {
	Birthday      string
	Today         string
	Constellation string
	Gender        string
	City          string
	Occupation    string
}

type Client interface {
	GenerateTodayFortune(ctx context.Context, profile FortuneProfile) (string, error)
}

type OpenAICompatibleClient struct {
	apiKey       string
	baseURL      string
	model        string
	templatePath string
	http         *http.Client
}

// NewOpenAICompatibleClient 创建兼容 OpenAI 协议的 LLM 客户端。
// 参数：apiKey - API 密钥；baseURL - 接口地址；model - 模型名称；timeout - 请求超时。
// 返回：*OpenAICompatibleClient - LLM 客户端实例。
func NewOpenAICompatibleClient(apiKey, baseURL, model string, timeout time.Duration) *OpenAICompatibleClient {
	return &OpenAICompatibleClient{
		apiKey:       apiKey,
		baseURL:      baseURL,
		model:        model,
		templatePath: "internal/llm/prompt_template.yaml",
		http:         &http.Client{Timeout: timeout},
	}
}

func (c *OpenAICompatibleClient) resolveTemplatePath() (string, error) {
	candidates := []string{c.templatePath, "internal/llm/prompt_template.yaml", "../llm/prompt_template.yaml", "backend/internal/llm/prompt_template.yaml"}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		tpl, err := LoadPromptTemplate(candidate)
		if err == nil {
			_ = tpl
			return filepath.Clean(candidate), nil
		}
	}
	return "", fmt.Errorf("failed to locate prompt template")
}

// GenerateTodayFortune 调用大模型接口生成今日运势。
// 参数：ctx - 上下文；profile - 用户资料。
// 返回：string - 运势文本；error - 调用或解析失败错误。
func (c *OpenAICompatibleClient) GenerateTodayFortune(ctx context.Context, profile FortuneProfile) (string, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return "", fmt.Errorf("LLM API config missing")
	}
	path, err := c.resolveTemplatePath()
	if err != nil {
		return "", err
	}
	tpl, err := LoadPromptTemplate(path)
	if err != nil {
		return "", err
	}
	userPrompt := RenderFortunePrompt(tpl, profile)

	payload := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": tpl.SystemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.5,
	}
	buf, _ := json.Marshal(payload)

	endpoint := strings.TrimRight(c.baseURL, "/")
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint = endpoint + "/chat/completions"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(buf))
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

	type chatResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	var parsed chatResponse
	if err = json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("llm empty response")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

type StubClient struct{}

// GenerateTodayFortune 作为未接入模型时的兜底实现。
// 参数：ctx - 上下文；profile - 用户资料。
// 返回：string - 为空；error - 未实现错误。
func (s *StubClient) GenerateTodayFortune(ctx context.Context, profile FortuneProfile) (string, error) {
	_ = ctx
	_ = profile
	return "", fmt.Errorf("LLM client not implemented")
}
