package update

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"time"

	"github.com/sojupie/tokens/source/crawl"
	"github.com/sojupie/tokens/source/publish"
)

type Issue struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Blocking bool   `json:"blocking"`
}

type Summary struct {
	Subjects int `json:"subjects"`
	Tokens   int `json:"tokens"`
	Values   int `json:"values"`
	Added    int `json:"added"`
	Updated  int `json:"updated"`
	Removed  int `json:"removed"`
}

type Report struct {
	SchemaVersion     int                 `json:"schemaVersion"`
	ID                string              `json:"id"`
	ParentRunID       string              `json:"parentRunId,omitempty"`
	Status            string              `json:"status"`
	Phase             string              `json:"phase"`
	StartedAt         string              `json:"startedAt"`
	FinishedAt        string              `json:"finishedAt,omitempty"`
	ApprovedAt        string              `json:"approvedAt,omitempty"`
	BaseGeneration    string              `json:"baseGeneration,omitempty"`
	HistoryDir        string              `json:"historyDir,omitempty"`
	ReceiptPath       string              `json:"receiptPath,omitempty"`
	SnapshotPath      string              `json:"snapshotPath,omitempty"`
	CandidatePath     string              `json:"candidatePath,omitempty"`
	CandidateDigest   string              `json:"candidateDigest,omitempty"`
	SnapshotDigest    string              `json:"snapshotDigest,omitempty"`
	Summary           Summary             `json:"summary"`
	Capture           crawl.CaptureReport `json:"capture"`
	Issues            []Issue             `json:"issues,omitempty"`
	Error             string              `json:"error,omitempty"`
	NotificationError string              `json:"notificationError,omitempty"`
	Crawl             crawl.Config        `json:"crawl"`
	PageTimeout       string              `json:"pageTimeout"`
	MaxChangeFraction float64             `json:"maxChangeFraction"`
}

var runIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,100}$`)

func stateRoot(path string) (string, error) {
	if path == "" {
		return "", errors.New("state directory is required")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		return "", err
	}
	// Linux deployments may place state on a symlinked data volume. Windows
	// development can lack permission to inspect symlink ancestors.
	if runtime.GOOS != "linux" {
		return path, nil
	}
	return filepath.EvalSymlinks(path)
}

func runDirectory(root, id string) (string, error) {
	if !runIDPattern.MatchString(id) {
		return "", errors.New("invalid run ID")
	}
	return filepath.Join(root, "runs", id), nil
}

func Read(root, id string) (Report, error) {
	directory, err := runDirectory(root, id)
	if err != nil {
		return Report{}, err
	}
	var report Report
	if err := readJSON(filepath.Join(directory, "report.json"), &report); err != nil {
		return Report{}, err
	}
	if report.SchemaVersion != 1 || report.ID != id {
		return Report{}, errors.New("invalid run report identity or version")
	}
	return report, nil
}

func List(root string) ([]Report, error) {
	entries, err := os.ReadDir(filepath.Join(root, "runs"))
	if os.IsNotExist(err) {
		return []Report{}, nil
	}
	if err != nil {
		return nil, err
	}
	reports := make([]Report, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		report, err := Read(root, entry.Name())
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].StartedAt > reports[j].StartedAt })
	return reports, nil
}

func save(root string, report Report) error {
	directory, err := runDirectory(root, report.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return err
	}
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(directory, ".report-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, err = file.Write(append(body, '\n'))
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := os.Rename(file.Name(), filepath.Join(directory, "report.json")); err != nil {
		return err
	}
	return syncReportDirectories(directory)
}

func readJSON(path string, value any) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, value)
}

func active(root string) (string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("state path is not a directory")
	}
	path := filepath.Join(root, "current")
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return publish.ResolveGeneration(path)
}

func newReport(config Config) Report {
	return Report{SchemaVersion: 1, ID: time.Now().UTC().Format("20060102T150405") + "-" + rand.Text()[:8], Status: "running", Phase: "capture", StartedAt: time.Now().UTC().Format(time.RFC3339Nano), Crawl: config.Crawl, PageTimeout: config.Crawl.PageTimeout.String(), MaxChangeFraction: config.MaxChangeFraction}
}

func lock(root string) (func(), error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(root, "update.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := lockFile(file); err != nil {
		file.Close()
		return nil, fmt.Errorf("another update may be running: %w", err)
	}
	return func() { file.Close() }, nil
}
