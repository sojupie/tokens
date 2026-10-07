package update

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

// Notify delivers a configured notification and records delivery failures.
// Credentials stay in the process environment, not in saved reports.
func Notify(ctx context.Context, root, id, webhook string) error {
	if !NotificationsConfigured(webhook) {
		return nil
	}
	report, err := Read(root, id)
	if err != nil {
		return err
	}
	err = sendNotification(ctx, root, report, webhook)
	report.NotificationError = ""
	if err != nil {
		report.NotificationError = err.Error()
	}
	return errors.Join(err, save(root, report))
}

// NotifyRun serializes a manual resend with updates and approvals.
func NotifyRun(ctx context.Context, root, id, webhook string) error {
	root, err := stateRoot(root)
	if err != nil {
		return err
	}
	unlock, err := lock(root)
	if err != nil {
		return err
	}
	defer unlock()
	return Notify(ctx, root, id, webhook)
}

// Alert is the systemd OnFailure fallback, including timeouts and process kills.
func Alert(ctx context.Context, root, webhook string) error {
	notice := Report{Status: "service_failed", Phase: "service", Error: "The update service failed. Check journalctl -u tokens-update.service."}
	originalRoot := root
	root, err := stateRoot(root)
	if err != nil {
		notice.Error += " State directory unavailable: " + err.Error()
		return sendNotification(ctx, originalRoot, notice, webhook)
	}
	unlock, err := lock(root)
	if err != nil {
		notice.Error += " Could not acquire update lock: " + err.Error()
		return sendNotification(ctx, root, notice, webhook)
	}
	defer unlock()
	reports, err := List(root)
	if err != nil {
		notice.Error += " Saved reports unavailable: " + err.Error()
		return sendNotification(ctx, root, notice, webhook)
	}
	if len(reports) == 0 {
		return sendNotification(ctx, root, notice, webhook)
	}
	report := reports[0]
	if report.Status == "running" {
		report.Status = "failed"
		report.Error = "The update process stopped before finishing. Check journalctl -u tokens-update.service."
		report.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
		if err := save(root, report); err != nil {
			return errors.Join(err, sendNotification(ctx, root, report, webhook))
		}
	}
	if report.Status == "failed" || report.Status == "needs_review" {
		return Notify(ctx, root, report.ID, webhook)
	}
	notice.ID = report.ID
	err = sendNotification(ctx, root, notice, webhook)
	report.NotificationError = ""
	if err != nil {
		report.NotificationError = err.Error()
	}
	return errors.Join(err, save(root, report))
}

func sendNotification(ctx context.Context, root string, report Report, webhook string) error {
	var deliveryErrors []error
	if smtpConfigured() {
		config, err := smtpEnvironment()
		if err == nil {
			// Failure notifications must survive cancellation of the update.
			mailContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			err = sendEmail(mailContext, config, root, report)
			cancel()
		}
		deliveryErrors = append(deliveryErrors, err)
	}
	if webhook != "" {
		deliveryErrors = append(deliveryErrors, sendWebhook(ctx, root, report, webhook))
	}
	return errors.Join(deliveryErrors...)
}

// TestNotification sends a test message without changing saved runs.
func TestNotification(ctx context.Context, root, webhook string) error {
	return sendNotification(ctx, root, Report{Status: "test", Phase: "notifications"}, webhook)
}

func sendWebhook(ctx context.Context, root string, report Report, webhook string) error {
	target, err := url.Parse(webhook)
	if err != nil || target.Host == "" || (target.Scheme != "https" && target.Scheme != "http") {
		return errors.New("notification endpoint must be an http(s) URL")
	}
	live, err := active(root)
	if err != nil {
		live = "unavailable"
		report.Error += " Cannot resolve the live publication: " + err.Error()
	}
	reportPath := ""
	if report.ID != "" {
		reportPath = filepath.Join(root, "runs", report.ID, "report.json")
	}
	body, err := json.Marshal(struct {
		RunID          string  `json:"runId"`
		Status         string  `json:"status"`
		Phase          string  `json:"phase"`
		Error          string  `json:"error,omitempty"`
		Issues         []Issue `json:"issues,omitempty"`
		LiveGeneration string  `json:"liveGeneration"`
		ReportPath     string  `json:"reportPath"`
		Summary        Summary `json:"summary"`
	}{report.ID, report.Status, report.Phase, report.Error, report.Issues, live, reportPath, report.Summary})
	if err != nil {
		return err
	}
	// A canceled capture must still be able to send its failure notification.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(body))
	if err != nil {
		return errors.New("cannot create notification request")
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		return errors.New("notification request failed; check endpoint and network")
	}
	defer response.Body.Close()
	io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("notification endpoint returned HTTP %d", response.StatusCode)
	}
	return nil
}
