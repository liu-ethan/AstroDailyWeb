package llm

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type PromptTemplate struct {
	UserPrompt string `yaml:"user_prompt"`
}

// LoadPromptTemplate 从 YAML 文件加载 Prompt 模板。
// 参数：path - 模板文件路径。
// 返回：PromptTemplate - 模板对象；error - 加载失败错误。
func LoadPromptTemplate(path string) (PromptTemplate, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return PromptTemplate{}, err
	}
	tpl := PromptTemplate{}
	if err = yaml.Unmarshal(content, &tpl); err != nil {
		return PromptTemplate{}, err
	}
	if strings.TrimSpace(tpl.UserPrompt) == "" {
		return PromptTemplate{}, fmt.Errorf("prompt user_prompt is empty")
	}
	return tpl, nil
}

// RenderFortunePrompt 渲染模板占位符为具体用户资料。
// 参数：tpl - 模板对象；profile - 运势输入资料。
// 返回：string - 渲染后的 Prompt 文本。
func RenderFortunePrompt(tpl PromptTemplate, profile FortuneProfile) string {
	r := tpl.UserPrompt
	r = strings.ReplaceAll(r, "{birthday}", profile.Birthday)
	r = strings.ReplaceAll(r, "{today}", profile.Today)
	r = strings.ReplaceAll(r, "{constellation}", profile.Constellation)
	r = strings.ReplaceAll(r, "{gender}", profile.Gender)
	r = strings.ReplaceAll(r, "{city}", profile.City)
	r = strings.ReplaceAll(r, "{occupation}", profile.Occupation)
	return r
}
