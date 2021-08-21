package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	gormlogger "gorm.io/gorm/logger"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var logger = gormlogger.Default.LogMode(gormlogger.Info)

var rootFlags struct {
	cfgFile  *string
	logLevel *string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "sqedule-server",
	Short:         "Sqedule server",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error(context.Background(), err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initGlobalLogger)

	rootFlags.cfgFile = rootCmd.PersistentFlags().String("config", "", "config file (default $HOME/.sqedule-server.yaml)")
	rootFlags.logLevel = rootCmd.PersistentFlags().String("log-level", "info", "log level, one of: error,warn,info,silent")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if *rootFlags.cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(*rootFlags.cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".sqedule" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".sqedule-server")
	}

	viper.SetEnvPrefix("SQEDULE")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func createLoggerWithLevel(logLevel string) (gormlogger.Interface, error) {
	switch logLevel {
	case "error":
		return gormlogger.Default.LogMode(gormlogger.Error), nil
	case "warn":
		return gormlogger.Default.LogMode(gormlogger.Warn), nil
	case "info":
		return gormlogger.Default.LogMode(gormlogger.Info), nil
	case "silent":
		return gormlogger.Default.LogMode(gormlogger.Silent), nil
	default:
		return nil, fmt.Errorf("invalid log level %s", logLevel)
	}
}

func initGlobalLogger() {
	newLogger, err := createLoggerWithLevel(*rootFlags.logLevel)
	if err != nil {
		logger.Error(context.Background(), "Error initializing logger: %s", err.Error())
		os.Exit(1)
	} else {
		logger = newLogger
	}
}

func main() {
	Execute()
}
