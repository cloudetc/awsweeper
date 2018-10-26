package resource

import (
	"regexp"
	"time"

	"github.com/sirupsen/logrus"

	"log"

	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// AppFs is an abstraction of the file system to allow mocking in tests.
var AppFs = afero.NewOsFs()

// Config represents the content of a yaml file that is used as a contract to filter resources for deletion.
type Config map[TerraformResourceType]resTypeConfig

// configEntry represents an entry in Config and selects the resources of a particular resource type.
type resTypeConfig struct {
	ID   *string           `yaml:",omitempty"`
	Tags map[string]string `yaml:",omitempty"`
	// select resources by creation time
	Created *Created `yaml:",omitempty"`
}

type Created struct {
	Before *time.Time `yaml:",omitempty"`
	After  *time.Time `yaml:",omitempty"`
}

// Filter selects resources based on a given yaml config.
type Filter struct {
	Cfg Config
}

// NewFilter creates a new filter based on a config given via a yaml file.
func NewFilter(yamlFile string) *Filter {
	return &Filter{
		Cfg: read(yamlFile),
	}
}

// read reads a filter from a yaml file.
func read(filename string) Config {
	var cfg Config

	data, err := afero.ReadFile(AppFs, filename)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to read config file: %s", filename)
	}

	err = yaml.UnmarshalStrict([]byte(data), &cfg)
	if err != nil {
		logrus.WithError(err).Fatalf("Cannot unmarshal config: %s", filename)
	}

	return cfg
}

// Validate checks if all resource types appearing in the config are currently supported.
func (f Filter) Validate() error {
	for _, resType := range f.Types() {
		if !SupportedResourceType(resType) {
			return fmt.Errorf("unsupported resource type found in yaml config: %s", resType)
		}
	}
	return nil
}

// Types returns all the resource types in the config.
func (f Filter) Types() []TerraformResourceType {
	resTypes := make([]TerraformResourceType, 0, len(f.Cfg))

	for k := range f.Cfg {
		resTypes = append(resTypes, k)
	}

	return resTypes
}

// MatchID checks whether a resource (given by its type and id) matches the filter.
func (f Filter) matchID(resType TerraformResourceType, id string) bool {
	cfgEntry, found := f.Cfg[resType]
	if !found {
		return false
	}

	if cfgEntry.ID == nil {
		return true
	}

	if ok, err := regexp.MatchString(*cfgEntry.ID, id); ok {
		if err != nil {
			log.Fatal(err)
		}
		return true
	}

	return false
}

// MatchesTags checks whether a resource (given by its type and findTags)
// matches the filter. The keys must match exactly, whereas the tag value is checked against a regex.
func (f Filter) matchTags(resType TerraformResourceType, tags map[string]string) bool {
	cfgEntry, found := f.Cfg[resType]
	if !found {
		return false
	}

	if cfgEntry.Tags == nil {
		return true
	}

	for cfgTagKey, regex := range cfgEntry.Tags {
		if tagVal, ok := tags[cfgTagKey]; ok {
			if matched, err := regexp.MatchString(regex, tagVal); !matched {
				if err != nil {
					log.Fatal(err)
				}
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func (f Filter) matchCreated(resType TerraformResourceType, creationTime *time.Time) bool {
	cfgEntry, found := f.Cfg[resType]
	if !found {
		return false
	}

	if cfgEntry.Created == nil {
		return true
	}

	if creationTime == nil {
		return false
	}

	createdAfter := true
	if cfgEntry.Created.After != nil {
		createdAfter = creationTime.Unix() > cfgEntry.Created.After.Unix()
	}

	createdBefore := true
	if cfgEntry.Created.Before != nil {
		createdBefore = creationTime.Unix() < cfgEntry.Created.Before.Unix()
	}

	return createdAfter && createdBefore
}

// matches checks whether a resource matches the filter criteria.
func (f Filter) matches(r *Resource) bool {
	return f.matchTags(r.Type, r.Tags) && f.matchID(r.Type, r.ID) && f.matchCreated(r.Type, r.Created)
}
