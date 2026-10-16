package tokensnapshot

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeGraphMatchesSchemaValidation(t *testing.T) {
	const valid = `{"schemaVersion":3,"snapshotId":"fixture","crawlId":"capture","pages":["https://example.test/specs"],"observations":[{"pageUrl":"https://example.test/specs","status":200}],"sourcePayloads":[{"pageUrls":[],"sourcePayloadUrl":"https://example.test/tokens.json","records":1}],"records":[{"id":"a","kind":"token","sourcePayloadId":"payload","ordinal":0,"raw":{"number":9007199254740993,"array":[false,null,{"value":"a b"}]}}]}`
	cases := map[string]string{
		"valid":              valid,
		"missing field":      strings.Replace(valid, `"crawlId":"capture",`, "", 1),
		"extra field":        strings.Replace(valid, `"schemaVersion":3`, `"extra":true,"schemaVersion":3`, 1),
		"wrong version":      strings.Replace(valid, `"schemaVersion":3`, `"schemaVersion":2`, 1),
		"null array":         strings.Replace(valid, `"pages":["https://example.test/specs"]`, `"pages":null`, 1),
		"bad page URI":       strings.Replace(valid, `"pages":["https://example.test/specs"]`, `"pages":["not a URI"]`, 1),
		"wrong page type":    strings.Replace(valid, `"pages":["https://example.test/specs"]`, `"pages":[17]`, 1),
		"bad status":         strings.Replace(valid, `"status":200`, `"status":1000`, 1),
		"bad payload count":  strings.Replace(valid, `"records":1`, `"records":-1`, 1),
		"bad ordinal":        strings.Replace(valid, `"ordinal":0`, `"ordinal":-1`, 1),
		"extra record field": strings.Replace(valid, `"ordinal":0`, `"ordinal":0,"extra":true`, 1),
		"missing raw":        strings.Replace(valid, `"raw":{"number":9007199254740993,"array":[false,null,{"value":"a b"}]}`, `"name":"token"`, 1),
		"null raw":           strings.Replace(valid, `"raw":{"number":9007199254740993,"array":[false,null,{"value":"a b"}]}`, `"raw":null`, 1),
		"scalar raw":         strings.Replace(valid, `"raw":{"number":9007199254740993,"array":[false,null,{"value":"a b"}]}`, `"raw":42`, 1),
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			wantErr := validateJSONSchema("generated-token-graph.schema.json", []byte(body))
			_, err := decodeGraph(strings.NewReader(body), true, true)
			if (err == nil) != (wantErr == nil) {
				t.Fatalf("stream validation %v; whole-document validation %v", err, wantErr)
			}
			_, err = decodeGraph(strings.NewReader(body), true, false)
			if (err == nil) != (wantErr == nil) {
				t.Fatalf("validation without retention %v; want %v", err, wantErr)
			}
		})
	}
	var original Graph
	if err := json.Unmarshal([]byte(valid), &original); err != nil {
		t.Fatal(err)
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, []byte(valid), "", "  "); err != nil {
		t.Fatal(err)
	}
	loaded, err := decodeGraph(&pretty, true, true)
	if err != nil {
		t.Fatal(err)
	}
	// Compact only raw formatting before comparing the complete typed graph.
	var compact bytes.Buffer
	if err := json.Compact(&compact, original.Records[0].Raw); err != nil {
		t.Fatal(err)
	}
	original.Records[0].Raw = compact.Bytes()
	if !reflect.DeepEqual(original, loaded) {
		t.Fatal("streaming changed graph values or provenance")
	}
	for _, body := range []string{valid + ` {}`, valid + ` garbage`, strings.TrimSuffix(valid, "}"), strings.Replace(valid, `"schemaVersion":3`, `"schemaVersion":3,"schemaVersion":3`, 1)} {
		if _, err := decodeGraph(strings.NewReader(body), true, true); err == nil {
			t.Fatalf("accepted malformed graph: %s", body)
		}
	}
}

func TestDecodeGraphLegacyAndEmptyArrays(t *testing.T) {
	for _, version := range []string{"2", "3"} {
		body := `{"schemaVersion":` + version + `,"snapshotId":"fixture","crawlId":"capture","pages":[],"observations":[],"sourcePayloads":[],"records":[]}`
		var want Graph
		if err := json.Unmarshal([]byte(body), &want); err != nil {
			t.Fatal(err)
		}
		got, err := decodeGraph(strings.NewReader(body), version == "3", true)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("version %s graph differs: %#v", version, got)
		}
	}
}

func TestGzipDigestChecksUncompressedBytesAndCorruption(t *testing.T) {
	root := t.TempDir()
	const body = `{"upstream":"data","number":9007199254740993}`
	writeGzipArtifact(t, root, "payload.gz", body)
	path := filepath.Join(root, "payload.gz")
	hash, size, err := gzipDigest(path)
	if err != nil {
		t.Fatal(err)
	}
	if hash != sha256Hex([]byte(body)) || size != int64(len(body)) {
		t.Fatalf("wrong gzip digest: %s, %d", hash, size)
	}
	compressed, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	compressed[len(compressed)-8] ^= 1 // Corrupt the gzip checksum, not JSON syntax.
	if err := os.WriteFile(path, compressed, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := gzipDigest(path); err == nil {
		t.Fatal("accepted corrupt gzip checksum")
	}
	if err := os.WriteFile(path, compressed[:len(compressed)-8], 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := gzipDigest(path); err == nil {
		t.Fatal("accepted truncated gzip")
	}
}
