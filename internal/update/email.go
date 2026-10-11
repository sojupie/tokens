package update

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type smtpConfig struct {
	host, port, username, password string
	from, to                       *mail.Address
	implicitTLS                    bool
	tlsConfig                      *tls.Config
}

func smtpConfigured() bool {
	for _, key := range []string{"HOST", "PORT", "USERNAME", "PASSWORD", "FROM", "TO"} {
		if os.Getenv("TOKENS_SMTP_"+key) != "" {
			return true
		}
	}
	return false
}

// NotificationsConfigured detects partial settings too, so mistakes are reported.
func NotificationsConfigured(webhook string) bool { return webhook != "" || smtpConfigured() }

func smtpEnvironment() (smtpConfig, error) {
	config := smtpConfig{host: os.Getenv("TOKENS_SMTP_HOST"), port: os.Getenv("TOKENS_SMTP_PORT"), username: os.Getenv("TOKENS_SMTP_USERNAME"), password: os.Getenv("TOKENS_SMTP_PASSWORD")}
	if config.host == "" || strings.ContainsAny(config.host, "\r\n /\\") {
		return config, errors.New("TOKENS_SMTP_HOST must be an SMTP hostname, not a URL")
	}
	if config.port == "" {
		config.port = "587"
	}
	port, err := strconv.Atoi(config.port)
	if err != nil || port < 1 || port > 65535 {
		return config, errors.New("TOKENS_SMTP_PORT must be between 1 and 65535")
	}
	if (config.username == "") != (config.password == "") {
		return config, errors.New("set both TOKENS_SMTP_USERNAME and TOKENS_SMTP_PASSWORD")
	}
	for _, field := range []struct {
		key     string
		address **mail.Address
	}{{"FROM", &config.from}, {"TO", &config.to}} {
		value := os.Getenv("TOKENS_SMTP_" + field.key)
		if strings.ContainsAny(value, "\r\n") {
			return config, fmt.Errorf("TOKENS_SMTP_%s must be one email address", field.key)
		}
		address, err := mail.ParseAddress(value)
		if err != nil {
			return config, fmt.Errorf("TOKENS_SMTP_%s must be one email address", field.key)
		}
		*field.address = address
	}
	config.implicitTLS = port == 465
	config.tlsConfig = &tls.Config{ServerName: config.host, MinVersion: tls.VersionTLS12}
	return config, nil
}

func emailMessage(config smtpConfig, root string, report Report) []byte {
	subject := "Token Catalog: update failed"
	intro := "The token update did not finish successfully. Check the report and service logs before retrying."
	if report.Status == "needs_review" {
		subject = "Token Catalog: update needs review"
		intro = "A token update needs your review before it can be published."
	} else if report.Status == "test" {
		subject = "Token Catalog: test notification"
		intro = "Email notifications are working. This test did not run a crawl or change the publication."
	} else if report.Status == "service_failed" {
		intro = "The token update service failed. Check its logs to see what happened."
	}
	var body strings.Builder
	fmt.Fprintf(&body, "%s\n\nStatus: %s\nPhase: %s\n", intro, report.Status, report.Phase)
	if report.ID != "" {
		fmt.Fprintf(&body, "Run: %s\nReport: %s\n", report.ID, filepath.Join(root, "runs", report.ID, "report.json"))
		fmt.Fprintf(&body, "\nChanges: %d added, %d updated, %d removed.\n", report.Summary.Added, report.Summary.Updated, report.Summary.Removed)
		fmt.Fprintf(&body, "\nInspect on the host:\nsudo docker compose -f deploy/compose.yaml run --rm --no-deps updater update inspect %s --state-dir %s\n", report.ID, root)
	}
	if report.Error != "" {
		fmt.Fprintf(&body, "\nError: %s\n", report.Error)
	}
	for _, issue := range report.Issues {
		fmt.Fprintf(&body, "\n%s: %s\n", issue.Code, issue.Message)
	}
	if report.Status != "test" {
		body.WriteString("\nService logs:\nsudo journalctl -u tokens-update.service -n 100 --no-pager\n")
	}
	var message strings.Builder
	fmt.Fprintf(&message, "From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", config.from.String(), config.to.String(), subject, time.Now().UTC().Format(time.RFC1123Z))
	writer := quotedprintable.NewWriter(&message)
	_, _ = writer.Write([]byte(body.String()))
	_ = writer.Close()
	return []byte(message.String())
}

func sendEmail(ctx context.Context, config smtpConfig, root string, report Report) error {
	connection, err := (&net.Dialer{}).DialContext(ctx, "tcp", net.JoinHostPort(config.host, config.port))
	if err != nil {
		return errors.New("SMTP connection failed; check host, port and network")
	}
	defer connection.Close()
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(10 * time.Second)
	}
	if err := connection.SetDeadline(deadline); err != nil {
		return errors.New("cannot set SMTP deadline")
	}
	if config.implicitTLS {
		secure := tls.Client(connection, config.tlsConfig)
		if err := secure.HandshakeContext(ctx); err != nil {
			return errors.New("SMTP TLS handshake failed; check hostname and certificate")
		}
		connection = secure
	}
	client, err := smtp.NewClient(connection, config.host)
	if err != nil {
		return errors.New("SMTP greeting failed")
	}
	defer client.Close()
	if !config.implicitTLS {
		if supported, _ := client.Extension("STARTTLS"); !supported {
			return errors.New("SMTP server does not support required STARTTLS")
		}
		if err := client.StartTLS(config.tlsConfig); err != nil {
			return errors.New("SMTP STARTTLS failed; check hostname and certificate")
		}
	}
	if config.username != "" {
		if err := client.Auth(smtp.PlainAuth("", config.username, config.password, config.host)); err != nil {
			return errors.New("SMTP authentication failed; check username and password")
		}
	}
	if err := client.Mail(config.from.Address); err != nil {
		return errors.New("SMTP server rejected the sender")
	}
	if err := client.Rcpt(config.to.Address); err != nil {
		return errors.New("SMTP server rejected the recipient")
	}
	writer, err := client.Data()
	if err != nil {
		return errors.New("SMTP server refused the message")
	}
	if _, err := writer.Write(emailMessage(config, root, report)); err != nil {
		return errors.New("SMTP message transfer failed")
	}
	if err := writer.Close(); err != nil {
		return errors.New("SMTP server did not accept the message")
	}
	if err := client.Quit(); err != nil {
		return errors.New("SMTP message accepted, but closing the session failed; check your inbox before resending")
	}
	return nil
}
