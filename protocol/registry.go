package protocol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hotsbuff/stormparser-go"
)

const schemaDirectory = "schemas/protocol/builds"

func Latest() (*Schema, error) {
	entries, err := stormparser.SchemaFiles.ReadDir(schemaDirectory)
	if err != nil {
		return nil, err
	}
	var latest uint32
	found := false
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		build, err := strconv.ParseUint(strings.TrimSuffix(entry.Name(), ".yaml"), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid protocol schema name %q: %w", entry.Name(), err)
		}
		if !found || uint32(build) > latest {
			latest = uint32(build)
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("no embedded protocol schemas")
	}
	return Load(latest)
}

func Load(build uint32) (*Schema, error) {
	data, err := stormparser.SchemaFiles.ReadFile(fmt.Sprintf("%s/%d.yaml", schemaDirectory, build))
	if err != nil {
		return nil, fmt.Errorf("protocol build %d is not embedded: %w", build, err)
	}
	schema, err := ParseSchema(data)
	if err != nil {
		return nil, err
	}
	if schema.Extends == 0 {
		return schema, nil
	}
	base, err := Load(schema.Extends)
	if err != nil {
		return nil, fmt.Errorf("protocol build %d extends %d: %w", build, schema.Extends, err)
	}
	result := *base
	result.Build = schema.Build
	result.Extends = schema.Extends
	result.Exact = schema.Exact
	return &result, nil
}

func LoadCompatible(build uint32) (*Schema, bool, error) {
	schema, err := Load(build)
	if err == nil {
		return schema, schema.Exact, nil
	}
	latest, latestErr := Latest()
	if latestErr != nil {
		return nil, false, latestErr
	}
	if build > latest.Build {
		return latest, false, nil
	}
	return nil, false, err
}
