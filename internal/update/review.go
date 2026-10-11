package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/sojupie/tokens/internal/publish"
	"github.com/sojupie/tokens/internal/tokensnapshot"
)

func counts(subjects []publish.SubjectView) (Summary, int) {
	summary := Summary{Subjects: len(subjects)}
	unresolved := 0
	for _, subject := range subjects {
		summary.Tokens += len(subject.Tokens)
		for _, token := range subject.Tokens {
			summary.Values += len(token.Values)
			for _, value := range token.Values {
				var decoded any
				if json.Unmarshal(value.Value, &decoded) != nil || decoded == nil || undefinedValue(decoded) {
					unresolved++
				}
			}
		}
	}
	return summary, unresolved
}

func undefinedValue(value any) bool {
	switch value := value.(type) {
	case map[string]any:
		if value["undefined"] == true {
			return true
		}
		for _, child := range value {
			if undefinedValue(child) {
				return true
			}
		}
	case []any:
		for _, child := range value {
			if undefinedValue(child) {
				return true
			}
		}
	}
	return false
}

func review(candidate publish.Publication, base string, fraction float64) (Summary, []Issue, error) {
	summary, unresolved := counts(candidate.Subjects)
	var issues []Issue
	if summary.Tokens == 0 || summary.Values == 0 {
		issues = append(issues, Issue{"empty_publication", "No usable token values were extracted.", true})
	}
	for _, change := range candidate.Changes {
		if change.SnapshotID != candidate.CurrentSnapshot.ID {
			continue
		}
		switch change.Type {
		case "added":
			summary.Added++
		case "changed":
			summary.Updated++
		case "removed":
			summary.Removed++
		}
	}
	if base == "" {
		issues = append(issues, Issue{"initial_baseline", "Review and approve the first complete capture before enabling weekly updates.", false})
		return summary, issues, nil
	}
	previous, err := publish.Load(filepath.Join(base, "tokens.sqlite"))
	if err != nil {
		return summary, nil, err
	}
	beforeSubjects := make([]publish.SubjectView, 0, len(previous.Subjects))
	nextSubjects := make(map[publish.SubjectKey]bool)
	pages := make(map[string]bool)
	for _, subject := range candidate.Subjects {
		nextSubjects[subject.Key] = true
		for _, page := range subject.SourcePageURLs {
			pages[page] = true
		}
	}
	for key, subject := range previous.Subjects {
		beforeSubjects = append(beforeSubjects, subject)
		if !nextSubjects[key] {
			issues = append(issues, Issue{"subject_removed", "Missing component or style: " + subject.Name, false})
		}
		for _, page := range subject.SourcePageURLs {
			if !pages[page] {
				issues = append(issues, Issue{"page_removed", "Previously published source page is missing: " + page, false})
			}
		}
	}
	before, previouslyUnresolved := counts(beforeSubjects)
	if summary.Tokens < before.Tokens {
		issues = append(issues, Issue{"token_count_drop", fmt.Sprintf("Token count fell from %d to %d.", before.Tokens, summary.Tokens), false})
	}
	if summary.Removed > 0 {
		issues = append(issues, Issue{"values_removed", fmt.Sprintf("%d contextual values or records were removed.", summary.Removed), false})
	}
	if unresolved > previouslyUnresolved {
		issues = append(issues, Issue{"unresolved_values", fmt.Sprintf("Unresolved values increased from %d to %d.", previouslyUnresolved, unresolved), false})
	}
	if before.Values > 0 && float64(summary.Added+summary.Updated+summary.Removed)/float64(before.Values) > fraction {
		issues = append(issues, Issue{"large_change", fmt.Sprintf("Changes exceed %.0f%% of the previous value count.", fraction*100), false})
	}
	return summary, issues, nil
}

func Approve(root, id string) (Report, error) {
	root, err := stateRoot(root)
	if err != nil {
		return Report{}, err
	}
	unlock, err := lock(root)
	if err != nil {
		return Report{}, err
	}
	defer unlock()
	report, err := Read(root, id)
	if err != nil {
		return report, err
	}
	if report.Status != "needs_review" && report.Status != "failed" && report.Status != "running" && report.Status != "published" && report.Status != "unchanged" {
		return report, errors.New("this run has no candidate to approve")
	}
	for _, issue := range report.Issues {
		if issue.Blocking {
			return report, fmt.Errorf("cannot approve: %s", issue.Message)
		}
	}
	if err := activate(root, &report); err != nil {
		return report, err
	}
	report.ApprovedAt = time.Now().UTC().Format(time.RFC3339Nano)
	report.FinishedAt = report.ApprovedAt
	if err := save(root, report); err != nil {
		return report, err
	}
	return report, nil
}

func activate(root string, report *Report) error {
	directory, err := runDirectory(root, report.ID)
	if err != nil {
		return err
	}
	if report.CandidatePath != filepath.Join(directory, "publication") || report.SnapshotPath != filepath.Join(directory, "snapshot") || report.CandidateDigest == "" || report.SnapshotDigest == "" {
		return errors.New("candidate is missing or does not belong to this run")
	}
	for _, item := range []struct{ path, digest string }{{report.CandidatePath, report.CandidateDigest}, {report.SnapshotPath, report.SnapshotDigest}} {
		digest, err := publish.GenerationDigest(item.path)
		if err != nil {
			return err
		}
		if digest != item.digest {
			return errors.New("candidate or snapshot changed after review; create a new run")
		}
	}
	manifest, err := tokensnapshot.Verify(report.SnapshotPath)
	if err != nil {
		return err
	}
	if !manifest.Capture.Complete || manifest.SnapshotID != report.ID {
		return errors.New("candidate snapshot is incomplete or has the wrong identity")
	}
	state, err := publish.Load(filepath.Join(report.CandidatePath, "tokens.sqlite"))
	if err != nil {
		return err
	}
	if state.CurrentSnapshot.ID != report.ID {
		return errors.New("candidate publication has the wrong snapshot")
	}
	current, err := active(root)
	if err != nil {
		return err
	}
	if current != report.CandidatePath {
		if current != report.BaseGeneration {
			return errors.New("live publication changed since this candidate was built; create a new run")
		}
		if err := publish.Activate(filepath.Join(root, "current"), report.CandidatePath); err != nil {
			return err
		}
	}
	// Also recovers interruption between activation and the final report write.
	report.Status = "published"
	report.Phase = "complete"
	report.Error = ""
	return nil
}
