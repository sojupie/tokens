package usage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	_ "modernc.org/sqlite"
)

const RetentionDays = 90

type Event struct {
	Time       time.Time `json:"time"`
	Event      string    `json:"event"`
	Route      string    `json:"route"`
	Subject    string    `json:"subject,omitempty"`
	Format     string    `json:"format,omitempty"`
	Kind       string    `json:"kind,omitempty"`
	Scope      string    `json:"scope,omitempty"`
	Method     string    `json:"method,omitempty"`
	Tool       string    `json:"tool,omitempty"`
	Status     int       `json:"status"`
	Success    bool      `json:"success"`
	DurationMS float64   `json:"duration_ms"`
	Bytes      int64     `json:"bytes"`
}

type Store struct {
	db     *sql.DB
	queue  chan Event
	done   chan struct{}
	mu     sync.RWMutex
	closed bool
	lost   atomic.Int64
	output io.Writer
}

func Open(directory string, output io.Writer) (*Store, error) {
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(directory, "usage.sqlite"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`PRAGMA busy_timeout=5000; PRAGMA journal_mode=WAL; PRAGMA max_page_count=32768;
 CREATE TABLE IF NOT EXISTS totals (
 hour TEXT NOT NULL, event TEXT NOT NULL, route TEXT NOT NULL, subject TEXT NOT NULL,
 format TEXT NOT NULL, kind TEXT NOT NULL, scope TEXT NOT NULL, method TEXT NOT NULL,
 tool TEXT NOT NULL, status INTEGER NOT NULL, success INTEGER NOT NULL,
 count INTEGER NOT NULL, bytes INTEGER NOT NULL, duration REAL NOT NULL, max_duration REAL NOT NULL,
 PRIMARY KEY(hour,event,route,subject,format,kind,scope,method,tool,status,success));
 CREATE TABLE IF NOT EXISTS metadata (key TEXT PRIMARY KEY, value TEXT NOT NULL);`)
	if err != nil {
		db.Close()
		return nil, err
	}
	if _, err = db.Exec("INSERT OR IGNORE INTO metadata VALUES ('started',?)", time.Now().UTC().Format(time.RFC3339)); err != nil {
		db.Close()
		return nil, err
	}
	s := &Store{db: db, queue: make(chan Event, 4096), done: make(chan struct{}), output: output}
	if err = s.expire(time.Now()); err != nil {
		db.Close()
		return nil, err
	}
	go s.writeLoop()
	return s, nil
}

// Record never waits for disk or notification delivery.
func (s *Store) Record(event Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return
	}
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}
	select {
	case s.queue <- event:
	default:
		s.lost.Add(1)
	}
}

func (s *Store) Close() error {
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		close(s.queue)
	}
	s.mu.Unlock()
	<-s.done
	return s.db.Close()
}

func (s *Store) expire(now time.Time) error {
	_, err := s.db.Exec("DELETE FROM totals WHERE hour < ?", now.UTC().AddDate(0, 0, -RetentionDays).Truncate(time.Hour).Format(time.RFC3339))
	return err
}

func (s *Store) writeLoop() {
	defer close(s.done)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	encoder := json.NewEncoder(s.output)
	lastDay := ""
	batch := make([]Event, 0, 256)
	flush := func() {
		now := time.Now().UTC()
		if day := now.Format("2006-01-02"); day != lastDay {
			if err := s.expire(now); err != nil {
				fmt.Fprintln(os.Stderr, "usage retention:", err)
			} else {
				lastDay = day
			}
		}
		if len(batch) > 0 {
			tx, err := s.db.Begin()
			if err == nil {
				for _, e := range batch {
					_, err = tx.Exec(`INSERT INTO totals VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
      ON CONFLICT(hour,event,route,subject,format,kind,scope,method,tool,status,success)
      DO UPDATE SET count=count+1,bytes=bytes+excluded.bytes,duration=duration+excluded.duration,max_duration=max(max_duration,excluded.max_duration)`,
						e.Time.UTC().Truncate(time.Hour).Format(time.RFC3339), e.Event, e.Route, e.Subject, e.Format, e.Kind, e.Scope, e.Method, e.Tool, e.Status, e.Success, 1, e.Bytes, e.DurationMS, e.DurationMS)
					if err != nil {
						break
					}
				}
				if err == nil {
					err = tx.Commit()
				} else {
					_ = tx.Rollback()
				}
			}
			if err != nil {
				s.lost.Add(int64(len(batch)))
				fmt.Fprintln(os.Stderr, "usage storage:", err)
			}
			batch = batch[:0]
		}
		if lost := s.lost.Swap(0); lost > 0 {
			_, err := s.db.Exec(`INSERT INTO totals VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
    ON CONFLICT(hour,event,route,subject,format,kind,scope,method,tool,status,success)
    DO UPDATE SET count=count+excluded.count`, now.Truncate(time.Hour).Format(time.RFC3339), "usage_lost", "", "", "", "", "", "", "", 0, false, lost, 0, 0, 0)
			if err != nil {
				s.lost.Add(lost)
				fmt.Fprintln(os.Stderr, "usage lost events:", lost, err)
			}
		}
	}
	for {
		select {
		case e, ok := <-s.queue:
			if !ok {
				flush()
				return
			}
			if s.output != nil {
				if err := encoder.Encode(e); err != nil {
					fmt.Fprintln(os.Stderr, "usage log:", err)
				}
			}
			batch = append(batch, e)
			if len(batch) >= 256 {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
