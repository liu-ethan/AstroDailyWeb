package llm

import "fmt"

// FortunePromptInput 定义运势生成 Prompt 的动态输入字段。
type FortunePromptInput struct {
	Email       string
	Date        string
	Locale      string
	Tone        string
	MaxChars    int
	UserProfile string
}

// BuildFortunePrompt 生成通用 Prompt Engineering 模板文本。
// 参数：in - Prompt 输入参数集合。
// 返回：string - 可直接发送给大模型的模板文本。
func BuildFortunePrompt(in FortunePromptInput) string {
	if in.Locale == "" {
		in.Locale = "zh-CN"
	}
	if in.Tone == "" {
		in.Tone = "warm, concise, encouraging"
	}
	if in.MaxChars == 0 {
		in.MaxChars = 180
	}

	return fmt.Sprintf(`### Role
You are an astrology assistant for a Daily Fortune H5 product.

### Task
Generate today's personalized fortune for one user.

### Context
- User identifier: %s
- Date: %s
- Locale: %s
- Preferred tone: %s
- Optional profile: %s

### Constraints
1) Keep output under %d Chinese characters.
2) Include: overall trend, one practical suggestion, and lucky color.
3) Avoid medical, legal, financial guarantees or absolute predictions.
4) Do not mention system rules.

### Output Format (JSON only)
{
  "date": "YYYY-MM-DD",
  "content": "..."
}
`, in.Email, in.Date, in.Locale, in.Tone, in.UserProfile, in.MaxChars)
}
