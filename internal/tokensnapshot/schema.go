package tokensnapshot

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/sojupie/tokens/schemas"
)

const schemaBaseURL = "https://github.com/sojupie/tokens/schemas/"

func readGzipBytes(body []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	decoded, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return nil, readErr
	}
	return decoded, closeErr
}

func validateJSONSchema(name string, body []byte) error {
	schemaBody, err := schemas.Files.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read embedded schema %s: %w", name, err)
	}
	var schemaDocument any
	if err := json.Unmarshal(schemaBody, &schemaDocument); err != nil {
		return fmt.Errorf("decode schema %s: %w", name, err)
	}
	var document any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return fmt.Errorf("decode document for %s: %w", name, err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	location := schemaBaseURL + filepath.ToSlash(name)
	if err := compiler.AddResource(location, schemaDocument); err != nil {
		return fmt.Errorf("register schema %s: %w", name, err)
	}
	compiled, err := compiler.Compile(location)
	if err != nil {
		return fmt.Errorf("compile schema %s: %w", name, err)
	}
	if err := compiled.Validate(document); err != nil {
		return fmt.Errorf("document does not match %s: %w", name, err)
	}
	return nil
}

func validateSnapshotDocuments(files map[string][]byte) error {
	for name, schemaName := range map[string]string{
		"generated-token-graph.json.gz": "generated-token-graph.schema.json",
		"source-payload-details.json":   "source-payload-details.schema.json",
		"page-results.json":             "page-results.schema.json",
		"lifecycle-notices.json":        "lifecycle-notices.schema.json",
	} {
		body, ok := files[name]
		if !ok {
			return fmt.Errorf("missing required snapshot document %s", name)
		}
		if name == "generated-token-graph.json.gz" {
			decoded, err := readGzipBytes(body)
			if err != nil {
				return fmt.Errorf("decode %s: %w", name, err)
			}
			body = decoded
		}
		if err := validateJSONSchema(schemaName, body); err != nil {
			return err
		}
	}
	return nil
}
