package output

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"unicode"
)

type EmailOutput struct {
	SMTPHost         string
	SMTPPort         int
	FromEmail        string
	FromPassword     string
	ToEmails         []string
	IgnoredProducers map[string]struct{}
}

func NewEmailOutput() *EmailOutput {
	password := os.Getenv("CONSENSUS_EMAIL_PASSWORD")
	recipients := os.Getenv("CONSENSUS_EMAIL_RECIPIENTS")

	var toEmails []string
	if recipients != "" {
		toEmails = strings.Split(recipients, ",")
		// Trim whitespace from each email
		for i := range toEmails {
			toEmails[i] = strings.TrimSpace(toEmails[i])
		}
		// Filter out empty strings
		filteredEmails := make([]string, 0, len(toEmails))
		for _, email := range toEmails {
			if email != "" {
				filteredEmails = append(filteredEmails, email)
			}
		}
		toEmails = filteredEmails
	}

	return &EmailOutput{
		SMTPHost:         "smtp.gmail.com",
		SMTPPort:         587,
		FromEmail:        "consensus.ai.25@gmail.com",
		FromPassword:     password,
		ToEmails:         toEmails,
		IgnoredProducers: make(map[string]struct{}),
	}
}

func (e *EmailOutput) Name() string {
	return "email"
}

func (e *EmailOutput) WithIgnored(producers ...string) Writer {
	if e.IgnoredProducers == nil {
		e.IgnoredProducers = make(map[string]struct{})
	}
	for _, producer := range producers {
		e.IgnoredProducers[producer] = struct{}{}
	}
	return e
}

func (e *EmailOutput) Send(content, sessionID, producer string) error {
	if _, ignored := e.IgnoredProducers[producer]; ignored {
		return nil
	}

	// Skip if no password or recipients configured
	if e.FromPassword == "" || len(e.ToEmails) == 0 {
		return nil
	}

	subject := e.formatSubject(sessionID, producer)
	body := e.formatBody(content, sessionID, producer)

	return e.sendEmail(subject, body)
}

func (e *EmailOutput) formatSubject(sessionID, producer string) string {
	switch producer {
	case "request":
		return fmt.Sprintf("[Consensus AI] Session %s - User Request", sessionID)
	case "prompt":
		return fmt.Sprintf("[Consensus AI] Session %s - Optimized Prompt", sessionID)
	case "openai":
		return fmt.Sprintf("[Consensus AI] Session %s - OpenAI Response", sessionID)
	case "anthropic":
		return fmt.Sprintf("[Consensus AI] Session %s - Anthropic Response", sessionID)
	case "gemini":
		return fmt.Sprintf("[Consensus AI] Session %s - Gemini Response", sessionID)
	default:
		return fmt.Sprintf("[Consensus AI] Session %s - %s", sessionID, toTitleCase(producer))
	}
}

func toTitleCase(input string) string {
	words := strings.Fields(input)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}
func (e *EmailOutput) formatBody(content, sessionID, producer string) string {
	contentType := e.getContentType(producer)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; margin: 20px; }
        .header { background-color: #f4f4f4; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        .session-id { color: #666; font-size: 0.9em; }
        .content-type { color: #333; font-weight: bold; margin: 10px 0; }
        .content { background-color: #fff; padding: 15px; border: 1px solid #ddd; border-radius: 5px; white-space: pre-wrap; }
        .provider-badge { display: inline-block; padding: 4px 8px; border-radius: 3px; font-size: 0.8em; font-weight: bold; margin-bottom: 10px; }
        .openai { background-color: #10a37f; color: white; }
        .anthropic { background-color: #d4a574; color: white; }
        .gemini { background-color: #4285f4; color: white; }
        .system { background-color: #6c757d; color: white; }
    </style>
</head>
<body>
    <div class="header">
        <div class="session-id">Session ID: %s</div>
        <div class="content-type">%s</div>
        %s
    </div>
    <div class="content">%s</div>
</body>
</html>`, sessionID, contentType, e.getProviderBadge(producer), content)

	return html
}

func (e *EmailOutput) getContentType(producer string) string {
	switch producer {
	case "request":
		return "User Request"
	case "prompt":
		return "Optimized Prompt"
	case "openai":
		return "OpenAI Response"
	case "anthropic":
		return "Anthropic Response"
	case "gemini":
		return "Gemini Response"
	default:
		return strings.Title(producer) + " Response"
	}
}

func (e *EmailOutput) getProviderBadge(producer string) string {
	switch producer {
	case "openai":
		return `<div class="provider-badge openai">OpenAI</div>`
	case "anthropic":
		return `<div class="provider-badge anthropic">Anthropic</div>`
	case "gemini":
		return `<div class="provider-badge gemini">Gemini</div>`
	case "request", "prompt":
		return `<div class="provider-badge system">System</div>`
	default:
		return ""
	}
}

func (e *EmailOutput) sendEmail(subject, body string) error {
	// Gmail SMTP configuration
	auth := smtp.PlainAuth("", e.FromEmail, e.FromPassword, e.SMTPHost)

	// Email headers and body
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		e.FromEmail, strings.Join(e.ToEmails, ","), subject, body)

	// Send email
	addr := fmt.Sprintf("%s:%d", e.SMTPHost, e.SMTPPort)
	return smtp.SendMail(addr, auth, e.FromEmail, e.ToEmails, []byte(message))
}

