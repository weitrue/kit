package conf

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

const (
	jsonTagKey = "json"
	jsonTagSep = ','
)

var (
	loaders = map[string]func([]byte, any) error{
		".json": LoadFromJsonBytes,
		".toml": LoadFromTomlBytes,
		".yaml": LoadFromYamlBytes,
		".yml":  LoadFromYamlBytes,
	}
)

// MustLoad loads config into v from path, exits on error.
func MustLoad(path string, v any, opts ...Option) {
	if err := Load(path, v, opts...); err != nil {
		log.Fatalf("error: config file %s, %s", path, err.Error())
	}
}

// Load loads config into v from file, .json, .yaml and .yml are acceptable.
func Load(file string, v any, opts ...Option) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	loader, ok := loaders[strings.ToLower(path.Ext(file))]
	if !ok {
		return fmt.Errorf("unrecognized file type: %s", file)
	}

	var opt options
	for _, o := range opts {
		o(&opt)
	}

	if opt.env {
		return loader([]byte(os.ExpandEnv(string(content))), v)
	}

	return loader(content, v)
}
