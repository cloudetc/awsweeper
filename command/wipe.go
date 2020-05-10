package command

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/cloudetc/awsweeper/internal"
	"github.com/cloudetc/awsweeper/resource"
	"github.com/jckuester/terradozer/pkg/provider"
	terradozerRes "github.com/jckuester/terradozer/pkg/resource"
	"github.com/mitchellh/cli"
	"gopkg.in/yaml.v2"
)

// Wipe is currently the only command.
//
// It deletes selected AWS resources by
// a given filter (yaml configuration file).
type Wipe struct {
	UI          cli.Ui
	dryRun      bool
	forceDelete bool
	client      *resource.AWS
	provider    *provider.TerraformProvider
	filter      *resource.Filter
	outputType  string
}

func list(c *Wipe) []terradozerRes.DestroyableResource {
	var destroyableRes []terradozerRes.DestroyableResource

	for _, resType := range c.filter.Types() {
		rawResources, err := c.client.RawResources(resType)
		if err != nil {
			log.WithError(err).Fatal("failed to get raw resources")
		}

		deletableResources, err := resource.DeletableResources(resType, rawResources)
		if err != nil {
			log.WithError(err).Fatal("failed to convert raw resources into deletable resources")
		}

		filteredRes := c.filter.Apply(resType, deletableResources, rawResources, c.client)
		for _, res := range filteredRes {
			print(res, c.outputType)
		}

		for _, resFiltered := range filteredRes {
			for _, r := range resFiltered {
				destroyableRes = append(destroyableRes, terradozerRes.New(string(r.Type), r.ID, c.provider))
			}
		}
	}

	return destroyableRes
}

// Run executes the wipe command.
func (c *Wipe) Run(args []string) int {
	if len(args) == 1 {
		filter, err := resource.NewFilter(args[0])
		if err != nil {
			log.WithError(err).Fatal("failed to create resource filter")
		}

		c.filter = filter

		err = c.filter.Validate()
		if err != nil {
			log.WithError(err).Fatal("invalid filter config")
		}
	} else {
		fmt.Println(help())
		return 1
	}

	internal.LogTitle("showing resources that would be deleted (dry run)")
	resources := list(c)

	if len(resources) == 0 {
		internal.LogTitle("no resources found to delete")
		return 0
	}

	internal.LogTitle(fmt.Sprintf("total number of resources that would be deleted: %d", len(resources)))

	if !c.dryRun {
		if !internal.UserConfirmedDeletion(os.Stdin, c.forceDelete) {
			return 0
		}

		internal.LogTitle("Starting to delete resources")

		numDeletedResources := terradozerRes.DestroyResources(resources, false, 10)

		internal.LogTitle(fmt.Sprintf("total number of deleted resources: %d", numDeletedResources))
	}

	return 0
}

func print(res resource.Resources, outputType string) {
	if len(res) == 0 {
		return
	}

	switch strings.ToLower(outputType) {
	case "string":
		printString(res)
	case "json":
		printJson(res)
	case "yaml":
		printYaml(res)
	default:
		log.WithField("output", outputType).Fatal("Unsupported output type")
	}
}

func printString(res resource.Resources) {
	fmt.Printf("\n\t---\n\tType: %s\n\tFound: %d\n\n", res[0].Type, len(res))

	for _, r := range res {
		printStat := fmt.Sprintf("\t\tId:\t\t%s", r.ID)
		if r.Tags != nil {
			if len(r.Tags) > 0 {
				var keys []string
				for k := range r.Tags {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				printStat += "\n\t\tTags:\t\t"
				for _, k := range keys {
					printStat += fmt.Sprintf("[%s: %v] ", k, r.Tags[k])
				}
			}
		}
		printStat += "\n"
		if r.Created != nil {
			printStat += fmt.Sprintf("\tCreated:\t%s", r.Created)
			printStat += "\n"
		}
		fmt.Println(printStat)
	}
	fmt.Print("\t---\n\n")
}

func printJson(res resource.Resources) {
	b, err := json.Marshal(res)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal resources into JSON")
	}

	fmt.Print(string(b))
}

func printYaml(res resource.Resources) {
	b, err := yaml.Marshal(res)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal resources into YAML")
	}

	fmt.Print(string(b))
}

// Help returns help information of this command
func (c *Wipe) Help() string {
	return help()
}

// Synopsis returns a short version of the help information of this command
func (c *Wipe) Synopsis() string {
	return "Delete AWS resources via a yaml configuration"
}
