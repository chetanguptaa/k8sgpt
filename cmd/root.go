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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/adrg/xdg"
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/cmd/serve"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/k8sgpt-ai/k8sgpt/cmd/analyze"
	"github.com/k8sgpt-ai/k8sgpt/cmd/auth"
	"github.com/k8sgpt-ai/k8sgpt/cmd/filters"
	"github.com/k8sgpt-ai/k8sgpt/cmd/generate"
	"github.com/k8sgpt-ai/k8sgpt/cmd/integration"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

var (
	cfgFile     string
	kubecontext string
	kubeconfig  string
	version     string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8sgpt",
	Short: "Kubernetes debugging powered by AI",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v string) {
	version = v
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	performConfigMigrationIfNeeded()

	cobra.OnInitialize(initConfig)

	var kubeconfigPath string
	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(analyze.AnalyzeCmd)
	rootCmd.AddCommand(filters.FiltersCmd)
	rootCmd.AddCommand(generate.GenerateCmd)
	rootCmd.AddCommand(integration.IntegrationCmd)
	rootCmd.AddCommand(serve.ServeCmd)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.k8sgpt.yaml)")
	rootCmd.PersistentFlags().StringVar(&kubecontext, "kubecontext", "", "Kubernetes context to use. Only required if out-of-cluster.")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", kubeconfigPath, "Path to a kubeconfig. Only required if out-of-cluster.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// the config will belocated under `~/.config/k8sgpt/k8sgpt.yaml` on linux
		configDir := filepath.Join(xdg.ConfigHome, "k8sgpt")

		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("k8sgpt")

		viper.SafeWriteConfig()
	}

	viper.Set("kubecontext", kubecontext)
	viper.Set("kubeconfig", kubeconfig)

	viper.SetEnvPrefix("K8SGPT")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//	fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func performConfigMigrationIfNeeded() {
	oldConfig, err := getLegacyConfigFilePath()
	cobra.CheckErr(err)
	oldConfigExists, err := util.FileExists(oldConfig)
	cobra.CheckErr(err)

	newConfig := getConfigFilePath()
	newConfigExists, err := util.FileExists(newConfig)
	cobra.CheckErr(err)

	configDir := filepath.Dir(newConfig)
	err = util.EnsureDirExists(configDir)
	cobra.CheckErr(err)

	if oldConfigExists && newConfigExists {
		fmt.Fprintln(os.Stderr, color.RedString("Warning: Legacy config file at `%s` detected! This file will be ignored!", oldConfig))
		return
	}

	if oldConfigExists && !newConfigExists {
		fmt.Fprintln(os.Stderr, color.RedString("Performing config file migration from `%s` to `%s`", oldConfig, newConfig))

		err = os.Rename(oldConfig, newConfig)
		cobra.CheckErr(err)
	}
}

func getConfigFilePath() string {
	return filepath.Join(xdg.ConfigHome, "k8sgpt", "k8sgpt.yaml")
}

func getLegacyConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".k8sgpt.yaml"), nil
}
