package email

import (
	"fmt"
	"strings"

	"github.com/boxingoctopus/snackmates/api/internal/config"
)

// Send delivers an HTML email using the configured provider (Mailgun or SMTP).
func Send(cfg config.Config, to, subject, body string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("email recipient is required")
	}

	switch strings.ToLower(strings.TrimSpace(cfg.EmailProvider)) {
	case "mailgun":
		return sendMailgun(cfg, to, subject, body)
	default:
		return sendSMTP(cfg, to, subject, body)
	}
}
