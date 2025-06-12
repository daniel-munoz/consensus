package output

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strings"
	"unicode"
)

// emailHTMLTemplate defines the HTML email template with placeholders:
// - {{.SessionID}}: unique identifier for the session
// - {{.ContentType}}: type of content (User Request, Optimized Prompt, etc.)
// - {{.ProviderBadge}}: HTML badge for the provider/system
// - {{.Content}}: the actual email content (HTML-escaped for security)
const emailHTMLTemplate = `<!DOCTYPE html>
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
        <div class="session-id">Session ID: {{.SessionID}}</div>
        <div class="content-type">{{.ContentType}}</div>
        {{.ProviderBadge}}
    </div>
    <div class="content">{{.Content}}</div>
</body>
</html>`

// subjectTemplates maps producer types to their email subject templates
var subjectTemplates = map[string]string{
	"request":   "%s Session %s - User Request",
	"prompt":    "%s Session %s - Optimized Prompt",
	"openai":    "%s Session %s - OpenAI Response",
	"anthropic": "%s Session %s - Anthropic Response",
	"gemini":    "%s Session %s - Gemini Response",
}

// contentTypeLabels maps producer types to their content type labels
var contentTypeLabels = map[string]string{
	"request":   "User Request",
	"prompt":    "Optimized Prompt",
	"openai":    "OpenAI Response",
	"anthropic": "Anthropic Response",
	"gemini":    "Gemini Response",
}

// providerBadges maps producer types to their HTML badge elements
var providerBadges = map[string]template.HTML{
	"openai":    template.HTML(`<div class="provider-badge openai">OpenAI</div>`),
	"anthropic": template.HTML(`<div class="provider-badge anthropic">Anthropic</div>`),
	"gemini":    template.HTML(`<div class="provider-badge gemini">Gemini</div>`),
	"request":   template.HTML(`<div class="provider-badge system">System</div>`),
	"prompt":    template.HTML(`<div class="provider-badge system">System</div>`),
}

// emailTemplateData holds the data for the email HTML template
type emailTemplateData struct {
	SessionID     string
	ContentType   string
	ProviderBadge template.HTML
	Content       string
}

type EmailConfig struct {
	SMTPHost       string
	SMTPPort       int
	FromEmail      string
	FromName       string
	PasswordEnvVar string
	SubjectPrefix  string
}

type EmailOutput struct {
	Config           EmailConfig
	FromPassword     string
	ToEmails         []string
	IgnoredProducers map[string]struct{}
}

func NewEmailOutputWithRecipients(config EmailConfig, recipients string) *EmailOutput {
	password := os.Getenv(config.PasswordEnvVar)

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
		Config:           config,
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

func (e *EmailOutput) Send(content, sessionID, producerName, producerType string) error {
	if _, ignored := e.IgnoredProducers[producerName]; ignored {
		return nil
	}

	// Skip if no password or recipients configured
	if e.FromPassword == "" || len(e.ToEmails) == 0 {
		return nil
	}

	subject := e.formatSubject(sessionID, producerName)
	body, err := e.formatBody(content, sessionID, producerType)
	if err != nil {
		return fmt.Errorf("failed to format email body: %w", err)
	}

	// SMTP configuration
	auth := smtp.PlainAuth("", e.Config.FromEmail, e.FromPassword, e.Config.SMTPHost)

	// Email headers and body
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		e.Config.FromEmail, strings.Join(e.ToEmails, ","), subject, body)

	// Send email
	addr := fmt.Sprintf("%s:%d", e.Config.SMTPHost, e.Config.SMTPPort)
	return smtp.SendMail(addr, auth, e.Config.FromEmail, e.ToEmails, []byte(message))
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

func (e *EmailOutput) formatSubject(sessionID, producerName string) string {
	if template, exists := subjectTemplates[producerName]; exists {
		return fmt.Sprintf(template, e.Config.SubjectPrefix, sessionID)
	}
	// Default case for unknown producers
	return fmt.Sprintf("%s Session %s - %s Response", e.Config.SubjectPrefix, sessionID, toTitleCase(producerName))
}

func (e *EmailOutput) formatBody(content, sessionID, producerType string) (string, error) {
	tmpl, err := template.New("email").Parse(emailHTMLTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	data := emailTemplateData{
		SessionID:     sessionID,
		ContentType:   e.getContentType(producerType),
		ProviderBadge: e.getProviderBadge(producerType),
		Content:       content, // Content will be HTML-escaped by the template
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}

	return buf.String(), nil
}

func (e *EmailOutput) getContentType(producer string) string {
	if label, exists := contentTypeLabels[producer]; exists {
		return label
	}
	// Default case for unknown producers
	return fmt.Sprintf("%s Response", toTitleCase(producer))
}

func (e *EmailOutput) getProviderBadge(producer string) template.HTML {
	if badge, exists := providerBadges[producer]; exists {
		return badge
	}
	// Default case for unknown producers - return empty HTML
	return template.HTML("")
}
