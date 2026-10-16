package update

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"mime/quotedprintable"
	"net"
	"net/http/httptest"
	"net/mail"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

func TestEmailSMTP(t *testing.T) {
	certificateServer := httptest.NewTLSServer(nil)
	certificate := certificateServer.TLS.Certificates[0]
	certificateServer.Close()
	parsed, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(parsed)
	for _, mode := range []string{"tls", "starttls", "untrusted", "plaintext"} {
		t.Run(mode, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			result := make(chan string, 1)
			go func() {
				connection, err := listener.Accept()
				if err != nil {
					result <- ""
					return
				}
				defer connection.Close()
				connection.SetDeadline(time.Now().Add(5 * time.Second))
				secureConfig := &tls.Config{Certificates: []tls.Certificate{certificate}, MinVersion: tls.VersionTLS12}
				if mode == "tls" || mode == "untrusted" {
					connection = tls.Server(connection, secureConfig)
				}
				reader := bufio.NewReader(connection)
				fmt.Fprint(connection, "220 localhost SMTP\r\n")
				body := ""
				authenticated := false
				defer func() { result <- body }()
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					switch {
					case strings.HasPrefix(line, "EHLO"):
						if mode == "plaintext" {
							fmt.Fprint(connection, "250 localhost\r\n")
						} else {
							fmt.Fprint(connection, "250-localhost\r\n250-STARTTLS\r\n250 AUTH PLAIN\r\n")
						}
					case strings.HasPrefix(line, "STARTTLS"):
						fmt.Fprint(connection, "220 Begin TLS\r\n")
						connection = tls.Server(connection, secureConfig)
						reader = bufio.NewReader(connection)
					case strings.HasPrefix(line, "AUTH PLAIN "):
						credentials, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(strings.TrimPrefix(line, "AUTH PLAIN ")))
						_, encrypted := connection.(*tls.Conn)
						authenticated = encrypted && string(credentials) == "\x00user\x00password"
						if authenticated {
							fmt.Fprint(connection, "235 Accepted\r\n")
						} else {
							fmt.Fprint(connection, "535 Rejected\r\n")
						}
					case strings.HasPrefix(line, "MAIL FROM:"), strings.HasPrefix(line, "RCPT TO:"):
						if !authenticated {
							fmt.Fprint(connection, "530 Authenticate first\r\n")
						} else {
							fmt.Fprint(connection, "250 Accepted\r\n")
						}
					case strings.HasPrefix(line, "DATA"):
						fmt.Fprint(connection, "354 Send message\r\n")
						message, err := textproto.NewReader(reader).ReadDotBytes()
						if err != nil {
							return
						}
						body = string(message)
						fmt.Fprint(connection, "250 Accepted\r\n")
					case strings.HasPrefix(line, "QUIT"):
						fmt.Fprint(connection, "221 Bye\r\n")
						return
					default:
						fmt.Fprint(connection, "500 Unknown command\r\n")
					}
				}
			}()
			host, port, _ := net.SplitHostPort(listener.Addr().String())
			config := smtpConfig{host: host, port: port, username: "user", password: "password", from: &mail.Address{Address: "tokens@example.test"}, to: &mail.Address{Address: "admin@example.test"}, implicitTLS: mode == "tls" || mode == "untrusted", tlsConfig: &tls.Config{ServerName: host, RootCAs: pool, MinVersion: tls.VersionTLS12}}
			if mode == "untrusted" {
				config.tlsConfig.RootCAs = x509.NewCertPool()
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err = sendEmail(ctx, config, "/var/lib/tokens", Report{ID: "run-123", Status: "needs_review", Phase: "review", Issues: []Issue{{Code: "initial_baseline", Message: "Review the first capture."}}})
			if mode == "plaintext" || mode == "untrusted" {
				if err == nil || <-result != "" {
					t.Fatal("must refuse plaintext or an untrusted certificate")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			message, err := mail.ReadMessage(strings.NewReader(<-result))
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(quotedprintable.NewReader(message.Body))
			if message.Header.Get("Subject") != "Token Catalog: update needs review" || !strings.Contains(string(body), "update inspect run-123") || !strings.Contains(string(body), "Review the first capture.") || strings.Contains(string(body), "password") {
				t.Fatalf("unexpected notification: %s", body)
			}
		})
	}
}

func TestEmailEnvironmentAndSavedDeliveryError(t *testing.T) {
	for _, key := range []string{"HOST", "PORT", "USERNAME", "PASSWORD", "FROM", "TO"} {
		t.Setenv("TOKENS_SMTP_"+key, "")
	}
	if NotificationsConfigured("") {
		t.Fatal("email must be opt-in")
	}
	t.Setenv("TOKENS_SMTP_HOST", "smtp.purelymail.com")
	t.Setenv("TOKENS_SMTP_PORT", "465")
	t.Setenv("TOKENS_SMTP_FROM", "tokens@example.test")
	t.Setenv("TOKENS_SMTP_TO", "admin@example.test")
	config, err := smtpEnvironment()
	if err != nil || !config.implicitTLS {
		t.Fatalf("valid TLS configuration: %+v %v", config, err)
	}
	t.Setenv("TOKENS_SMTP_TO", "admin@example.test\r\nBcc: other@example.test")
	if _, err := smtpEnvironment(); err == nil {
		t.Fatal("must reject header injection")
	}
	root := t.TempDir()
	report := newReport(Config{})
	report.Status = "needs_review"
	if err := save(root, report); err != nil {
		t.Fatal(err)
	}
	if err := Notify(context.Background(), root, report.ID, ""); err == nil {
		t.Fatal("invalid email configuration must not fail silently")
	}
	saved, err := Read(root, report.ID)
	if err != nil || saved.NotificationError == "" {
		t.Fatalf("delivery error not recorded: %+v %v", saved, err)
	}
}
