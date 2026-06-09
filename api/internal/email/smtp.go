package email

import (
	"fmt"
	"net/smtp"

	"github.com/boxingoctopus/snackmates/api/internal/config"
)

func sendSMTP(cfg config.Config, to, subject, body string) error {
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		cfg.EmailFrom, to, subject, body,
	)
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	return smtp.SendMail(addr, nil, cfg.EmailFrom, []string{to}, []byte(msg))
}
