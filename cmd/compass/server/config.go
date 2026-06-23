package server

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	gemara "github.com/gemaraproj/go-gemara"
	"github.com/goccy/go-yaml"

	"github.com/complytime/gemara-content-service/mapper"
	"github.com/complytime/gemara-content-service/mapper/factory"
)

func NewScopeFromCatalogPath(catalogPath string) (mapper.Scope, error) {
	cleanedPath := filepath.Clean(catalogPath)
	slog.Debug("loading catalog", slog.String("path", cleanedPath))

	catalogData, err := os.ReadFile(cleanedPath)
	if err != nil {
		return nil, err
	}

	var catalog gemara.ControlCatalog
	err = yaml.Unmarshal(catalogData, &catalog)
	if err != nil {
		return nil, err
	}

	slog.Debug("catalog loaded",
		slog.String("catalog_id", catalog.Metadata.Id),
	)

	return mapper.Scope{
		catalog.Metadata.Id: catalog,
	}, nil
}

type Config struct {
	Plugins     []PluginConfig `json:"plugins" yaml:"plugins"`
	Certificate CertConfig     `json:"certConfig" yaml:"certConfig"`
	JWTAuth     JWTAuthConfig  `json:"jwtAuth" yaml:"jwtAuth"`
}

type JWTAuthConfig struct {
	Enabled             bool     `json:"enabled" yaml:"enabled"`
	IssuerURL           string   `json:"issuerUrl" yaml:"issuerUrl"`
	KubernetesServiceIP string   `json:"kubernetesServiceIp" yaml:"kubernetesServiceIp"`
	ExpectedAudience    string   `json:"expectedAudience" yaml:"expectedAudience"`
	AllowedSubjects     []string `json:"allowedSubjects" yaml:"allowedSubjects"`
}

type CertConfig struct {
	CertPath string `json:"cert" yaml:"cert"`
	KeyPath  string `json:"key" yaml:"key"`
}

type PluginConfig struct {
	Id             string `json:"id" yaml:"id"`
	EvaluationsDir string `json:"evaluations-dir" yaml:"evaluations-dir"`
}

func NewMapperSet(config *Config) (mapper.Set, error) {
	pluginSet := make(mapper.Set)
	slog.Debug("loading plugins", slog.Int("count", len(config.Plugins)))

	for _, pluginConf := range config.Plugins {
		transformerId := mapper.ID(pluginConf.Id)
		if pluginConf.EvaluationsDir == "" {
			slog.Info("plugin has no evaluations; skipping",
				slog.String("plugin_id", string(transformerId)),
			)
			continue
		}

		info, err := os.Stat(pluginConf.EvaluationsDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return pluginSet, fmt.Errorf("evaluations directory %s for plugin %s: %w", pluginConf.EvaluationsDir, pluginConf.Id, err)
			}
			return pluginSet, err
		}

		if !info.IsDir() {
			return pluginSet, fmt.Errorf("evaluations directory %s for plugin %s is not a directory", pluginConf.EvaluationsDir, pluginConf.Id)
		}

		tfmr, err := NewMapperFromDir(transformerId, pluginConf.EvaluationsDir)
		if err != nil {
			return pluginSet, fmt.Errorf("unable to load configuration for %s: %w", pluginConf.Id, err)
		}
		pluginSet[transformerId] = tfmr
	}
	slog.Debug("plugins loaded", slog.Int("count", len(pluginSet)))
	return pluginSet, nil
}

func NewMapperFromDir(pluginID mapper.ID, evaluationsPath string) (mapper.Mapper, error) {
	mpr := factory.MapperByID(pluginID)

	root, err := os.OpenRoot(evaluationsPath)
	if err != nil {
		return mpr, fmt.Errorf("opening root directory %s: %w", evaluationsPath, err)
	}
	defer root.Close()

	err = fs.WalkDir(root.FS(), ".", func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		content, err := root.ReadFile(relPath)
		if err != nil {
			return err
		}

		var evaluation mapper.EvaluationPlan
		err = yaml.Unmarshal(content, &evaluation)
		if err != nil {
			return err
		}

		// Extract reference-ids from Assessment Plans to determine the
		// control source.
		for _, plan := range evaluation.Plans {
			if plan.Control.ReferenceId == "" {
				continue
			}
			mpr.AddEvaluationPlan(plan.Control.ReferenceId, plan)
		}
		return nil
	})
	if err != nil {
		return mpr, err
	}
	slog.Info("plugin evaluations loaded",
		slog.String("plugin_id", string(pluginID)),
		slog.String("dir", evaluationsPath),
	)
	return mpr, nil
}
