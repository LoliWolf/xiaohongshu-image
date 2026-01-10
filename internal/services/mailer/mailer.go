package mailer

import (
	"fmt"
	"time"

	"github.com/xiaohongshu-image/internal/config"
	"gopkg.in/gomail.v2"
)

type Email struct {
	To      string
	Subject string
	Body    string
	IsHTML  bool
}

type Service struct {
	dialer *gomail.Dialer
	from   string
}

func NewService(cfg *config.SMTPConfig) *Service {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Password)
	dialer.Timeout = 30 * time.Second

	return &Service{
		dialer: dialer,
		from:   cfg.From,
	}
}

func (s *Service) Send(email Email) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", email.To)
	m.SetHeader("Subject", email.Subject)

	if email.IsHTML {
		m.SetBody("text/html", email.Body)
	} else {
		m.SetBody("text/plain", email.Body)
	}

	return s.dialer.DialAndSend(m)
}

func (s *Service) SendResultEmail(to, requestType, prompt, resultURL string) error {
	subject := fmt.Sprintf("您的%s生成结果已就绪", s.getRequestTypeText(requestType))

	body := fmt.Sprintf(`您好！

您请求的%s已经生成完成。

请求描述：%s

下载链接：%s

链接有效期为1小时，请及时下载。

此邮件由系统自动发送，请勿回复。`,
		s.getRequestTypeText(requestType),
		prompt,
		resultURL,
	)

	return s.Send(Email{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	})
}

func (s *Service) SendErrorEmail(to, requestType, prompt, errorMsg string) error {
	subject := fmt.Sprintf("您的%s生成失败", s.getRequestTypeText(requestType))

	body := fmt.Sprintf(`您好！

很抱歉，您请求的%s生成失败。

请求描述：%s

错误信息：%s

请稍后重试或联系管理员。

此邮件由系统自动发送，请勿回复。`,
		s.getRequestTypeText(requestType),
		prompt,
		errorMsg,
	)

	return s.Send(Email{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	})
}

func (s *Service) getRequestTypeText(requestType string) string {
	switch requestType {
	case "image":
		return "图片"
	case "video":
		return "视频"
	default:
		return "内容"
	}
}
