package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type SMTPClient interface {
	Send(ctx context.Context, to []string, subject, body string) error
	SendVerifyCode(ctx context.Context, to []string, code string) error
}

type Client struct {
	host string
	port int
	user string
	pass string
	from string
}

type mailTemplate struct {
	Subject string `yaml:"subject"`
	Content string `yaml:"content"`
}

const defaultTemplatePath = "internal/notify/mail.yaml"
const verifyCodeTemplatePath = "internal/notify/verify_code.yaml"

// NewClient 创建 SMTP 客户端。
// 参数：host - SMTP 主机；port - SMTP 端口；user - 用户名；pass - 密码或授权码；from - 发件人邮箱。
// 返回：*Client - SMTP 客户端实例。
func NewClient(host string, port int, user, pass, from string) *Client {
	return &Client{host: host, port: port, user: user, pass: pass, from: from}
}

// Send 发送模板邮件。
// 参数：ctx - 上下文；to - 收件人列表；subject - 兼容参数（当前不使用模板外主题）；body - 作为模板中的 {content}。
// 返回：error - 发送失败时返回错误。
func (c *Client) Send(ctx context.Context, to []string, subject, body string) error {
	if len(to) == 0 {
		return fmt.Errorf("empty receiver list")
	}
	_ = subject
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tpl, err := loadMailTemplate(defaultTemplatePath)
	if err != nil {
		return err
	}
	date := time.Now().Format("2006-01-02")
	name := receiverName(to[0])
	// 将 LLM 输出的换行符转为 HTML <br>，邮件以 HTML 格式发送
	htmlBody := strings.ReplaceAll(body, "\n", "<br>")
	renderedSubject := renderTemplate(tpl.Subject, name, date, htmlBody)
	renderedBody := renderTemplate(tpl.Content, name, date, htmlBody)

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	auth := smtp.PlainAuth("", c.user, c.pass, c.host)
	msg := buildHTMLMessage(c.from, to, renderedSubject, renderedBody)

	// 163 邮箱常用 465 SSL 端口，这里使用 TLS 直连发送。
	dialer := &tls.Dialer{NetDialer: &net.Dialer{Timeout: 5 * time.Second}}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, c.host)
	if err != nil {
		return err
	}
	defer func() { _ = client.Quit() }()

	if err = client.Auth(auth); err != nil {
		return err
	}
	if err = client.Mail(c.from); err != nil {
		return err
	}
	for _, receiver := range to {
		if err = client.Rcpt(receiver); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write([]byte(msg)); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return nil
}

func loadMailTemplate(path string) (mailTemplate, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return mailTemplate{}, err
	}
	tpl := mailTemplate{}
	if err = yaml.Unmarshal(b, &tpl); err != nil {
		return mailTemplate{}, err
	}
	if strings.TrimSpace(tpl.Subject) == "" || strings.TrimSpace(tpl.Content) == "" {
		return mailTemplate{}, fmt.Errorf("mail template is invalid")
	}
	return tpl, nil
}

func renderTemplate(templateText, name, date, content string) string {
	r := templateText
	r = strings.ReplaceAll(r, "{name}", name)
	r = strings.ReplaceAll(r, "{date}", date)
	r = strings.ReplaceAll(r, "{content}", content)
	return r
}

func receiverName(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "用户"
	}
	return parts[0]
}

func buildMessage(from string, to []string, subject, body string) string {
	headers := []string{
		"From: " + from,
		"To: " + strings.Join(to, ";"),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}

func buildHTMLMessage(from string, to []string, subject, body string) string {
	headers := []string{
		"From: " + from,
		"To: " + strings.Join(to, ";"),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}

// SendVerifyCode 使用验证码专用模板发送邮件。
// 参数：ctx - 上下文；to - 收件人列表；code - 验证码字符串。
// 返回：error - 发送失败时返回错误。
func (c *Client) SendVerifyCode(ctx context.Context, to []string, code string) error {
	if len(to) == 0 {
		return fmt.Errorf("empty receiver list")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tpl, err := loadMailTemplate(verifyCodeTemplatePath)
	if err != nil {
		return err
	}
	date := time.Now().Format("2006-01-02")
	name := receiverName(to[0])
	renderedSubject := renderTemplate(tpl.Subject, name, date, code)
	renderedBody := renderTemplate(tpl.Content, name, date, code)

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	auth := smtp.PlainAuth("", c.user, c.pass, c.host)
	msg := buildMessage(c.from, to, renderedSubject, renderedBody)

	dialer := &tls.Dialer{NetDialer: &net.Dialer{Timeout: 5 * time.Second}}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, c.host)
	if err != nil {
		return err
	}
	defer func() { _ = client.Quit() }()

	if err = client.Auth(auth); err != nil {
		return err
	}
	if err = client.Mail(c.from); err != nil {
		return err
	}
	for _, receiver := range to {
		if err = client.Rcpt(receiver); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write([]byte(msg)); err != nil {
		return err
	}
	return w.Close()
}
