/*
Copyright © 2024 P4K Ennead  <ennead.tbc@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"wartracker/pkg/db"
	"wartracker/pkg/scanner"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	EnableTraverseRunHooks bool

	cfgFile     string
	DBFile      string
	ScratchDir  string
	Debug       bool
	TessdataDir string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "wartracker-cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initDB, initScratch, initDebug, initTessdataDir)

	EnableTraverseRunHooks = true

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wartracker-cli.yaml)")
	RootCmd.PersistentFlags().StringVar(&DBFile, "dbfile", "", "database file")
	RootCmd.PersistentFlags().StringVar(&ScratchDir, "scratch", "", "Directory to store scratch files")
	RootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "Directory to store scratch files")
	RootCmd.PersistentFlags().StringVar(&TessdataDir, "tessdata", "", "Tesseract data directory")
	cobra.CheckErr(viper.BindPFlag("dbfile", RootCmd.PersistentFlags().Lookup("dbfile")))
	cobra.CheckErr(viper.BindPFlag("scratch", RootCmd.PersistentFlags().Lookup("scratch")))
	cobra.CheckErr(viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug")))
	cobra.CheckErr(viper.BindPFlag("tessdata", RootCmd.PersistentFlags().Lookup("tessdata")))
	viper.SetDefault("dbfile", "db/wartracker.db")
	viper.SetDefault("scratch", "_scratch")
	viper.SetDefault("debug", false)
	viper.SetDefault("tessdata", "/Users/erumer/src/github.com/tesseract-ocr/tessdata")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".wartracker-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".wartracker-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func initDB() {
	var err error
	db.Connection, err = db.Connect(viper.GetString("dbfile"))
	if err != nil {
		panic(err)
	}
}

func initScratch() {
	ScratchDir := viper.GetString("scratch")
	err := os.RemoveAll(ScratchDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to initialize scratch directory: ", ScratchDir)
	}
	err = os.MkdirAll(ScratchDir, 0755)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to initialize scratch directory: ", ScratchDir)
	}
}

func initDebug() {
	if viper.GetBool("debug") {
		scanner.Debug = true
		scanner.Process = os.Getpid()
		scanner.ScratchDir = viper.GetString("scratch")
	} else {
		scanner.Debug = false
	}
}

func initTessdataDir() {
	TessdataDir = viper.GetString("tessdata")
	scanner.TessdataDir = TessdataDir
}
