package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SMTPClient interface {
	Send(ctx context.Context, to []string, subject, body string) error
}

type Client struct {
	host string
	port int
	user string
	pass string
	from string
}

// NewClient 创建 SMTP 客户端。
// 参数：host - SMTP 主机；port - SMTP 端口；user - 用户名；pass - 密码或授权码；from - 发件人邮箱。
// 返回：*Client - SMTP 客户端实例。
func NewClient(host string, port int, user, pass, from string) *Client {
	return &Client{host: host, port: port, user: user, pass: pass, from: from}
}

// Send 发送纯文本邮件。
// 参数：ctx - 上下文；to - 收件人列表；subject - 邮件主题；body - 邮件正文。
// 返回：error - 发送失败时返回错误。
func (c *Client) Send(ctx context.Context, to []string, subject, body string) error {
	if len(to) == 0 {
		return fmt.Errorf("empty receiver list")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	auth := smtp.PlainAuth("", c.user, c.pass, c.host)

	header := map[string]string{
		"From":         c.from,
		"To":           strings.Join(to, ";"),
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}
	msg := ""
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body

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
