package mailer


import (
	"context"
	"fmt"
	"net/smtp"
)

type SMTPSender struct {
	host string
	port string
	from string
	auth smtp.Auth
}

func NewSMTPSender(host, port, user, password string) *SMTPSender {
	return &SMTPSender{
		host: host,
		port: port,
		from: user,
		auth: smtp.PlainAuth("", user, password, host),
	}
}

func (s *SMTPSender) SendVerificationCode(ctx context.Context, toEmail, code string) error {
	// ТВОЯ ЛОГИКА:
	// 1. Собрать адрес сервера: host:port (например, "smtp.mail.ru:587")
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	// 2. Собрать тело письма (Заголовок Subject + текст с кодом)
	msg := fmt.Sprintf("Subject: Registration Code\r\n\r\nYour code is: %s", code)
	// 3. Вызвать smtp.SendMail(...) и вернуть ошибку, если она есть
	err := smtp.SendMail(addr, s.auth, s.from, []string{toEmail}, []byte(msg))

	if err != nil {
		return fmt.Errorf("smtp.SendMail: %w", err)
	}
	return nil
}