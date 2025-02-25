/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package filters

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available filters",
	Long:  `The list command displays a list of available filters that can be used to analyze Kubernetes resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		activeFilters := viper.GetStringSlice("active_filters")
		coreFilters, additionalFilters, integrationFilters := analyzer.ListFilters()

		availableFilters := append(append(coreFilters, additionalFilters...), integrationFilters...)

		if len(activeFilters) == 0 {
			activeFilters = coreFilters
		}
		inactiveFilters := util.SliceDiff(availableFilters, activeFilters)
		fmt.Printf(color.YellowString("Active: \n"))
		for _, filter := range activeFilters {

			// if the filter is an integration, mark this differently
			if util.SliceContainsString(integrationFilters, filter) {
				fmt.Printf("> %s\n", color.BlueString("%s (integration)", filter))
			} else {
				fmt.Printf("> %s\n", color.GreenString(filter))
			}
		}

		// display inactive filters
		if len(inactiveFilters) != 0 {
			fmt.Printf(color.YellowString("Unused: \n"))
			for _, filter := range inactiveFilters {
				// if the filter is an integration, mark this differently
				if util.SliceContainsString(integrationFilters, filter) {
					fmt.Printf("> %s\n", color.BlueString("%s (integration)", filter))
				} else {
					fmt.Printf("> %s\n", color.RedString(filter))
				}
			}
		}

	},
}
