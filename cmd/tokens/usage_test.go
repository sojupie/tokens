package main

import (
	"strings"
	"testing"
	"time"

	"github.com/sojupie/tokens/internal/usage"
)

func TestUsageCommandValidationAndWeeklyDeduplication(t *testing.T) {
	for _, args := range [][]string{{"--since", "0d"}, {"--since", "91d"}, {"--since", "7h"}, {"unexpected"}, {"--svg", "--send"}, {"--svg", "--weekly"}} {
		if err := runUsage(args); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	if err := runUsage([]string{"--usage-dir", t.TempDir()}); err == nil {
		t.Fatal("missing database silently created")
	}
	dir := t.TempDir()
	store, err := usage.Open(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveDigest(time.Now().UTC().Truncate(time.Hour)); err != nil {
		t.Fatal(err)
	}
	store.Close()
	if err := runUsage([]string{"--usage-dir", dir, "--weekly"}); err != nil {
		t.Fatal(err)
	}
	if err := runUsage([]string{"--usage-dir", dir, "--since", "7d"}); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"HOST", "PORT", "USERNAME", "PASSWORD", "FROM", "TO"} {
		t.Setenv("TOKENS_SMTP_"+key, "")
	}
	if err := runUsage([]string{"--usage-dir", dir, "--send"}); err == nil || !strings.Contains(err.Error(), "SMTP") {
		t.Fatalf("missing SMTP not detected: %v", err)
	}
}
