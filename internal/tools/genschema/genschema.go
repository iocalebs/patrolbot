// Genschema generates a JSON schema for PatrolBot config.
// This schema provides in-editor autocompletion and validation for the config.yaml
// that is unmarshalled into that struct.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/iocalebs/patrolbot/internal/config"
)

func main() {
	err := generateSchema()
	if err != nil {
		log.Fatal(err)
	}
}

func generateSchema() error {
	reflector := new(jsonschema.Reflector)

	err := reflector.AddGoComments("github.com/iocalebs/patrolbot", "./internal/config")
	if err != nil {
		return fmt.Errorf("failed to create reflector: %w", err)
	}

	schema := reflector.Reflect(&config.Config{})

	file, err := os.Create("config.schema.json")
	if err != nil {
		return fmt.Errorf("failed to create schema file: %w", err)
	}
	defer file.Close() //nolint:errcheck

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	err = enc.Encode(schema)
	if err != nil {
		return fmt.Errorf("failed to JSON-encode schema: %w", err)
	}

	return nil
}
