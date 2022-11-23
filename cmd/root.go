package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var logger = log.New(os.Stdout, "[ root ]\t", log.Llongfile|log.Lmsgprefix|log.LstdFlags)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "og",
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
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	//logger.Println("init")
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//logger.Println("reading flags")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "c", "config file (default is $HOME/.og.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringArrayP("file", "f", nil, "files to be edited; might be provided multiple times; example: -f file.go")
	//_ = rootCmd.MarkFlagRequired("file")
	_ = viper.BindPFlag("files", rootCmd.Flag("file"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	//logger.Println("init config")

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("get working directory error:", err)
	}
	//logger.Println("Working directory:", wd)

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".og" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(wd)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".og")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// set defaults
	viper.SetDefault("workdir", wd)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Println("Using config file:", viper.ConfigFileUsed())
	}
}
