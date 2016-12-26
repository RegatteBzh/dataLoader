package main

import (
	"fmt"

	"github.com/regattebzh/dataLoader/etopo"
	"github.com/regattebzh/dataLoader/wind"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/* Main Command to parse
   command line */
var mainCommand = &cobra.Command{
	Use:   "dataLoader",
	Short: "Data loader",
	Long:  "Load data in Redis",
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetEnvPrefix("regat")
		viper.AutomaticEnv()
		// Application statup here
		err := mainApp()
		if err != nil {
			fmt.Println(err)
		}
	},
}

/**
 * The Main application really starts here
 */
func mainApp() (err error) {

	return nil
}

func main() {
	mainCommand.Execute()
}

func init() {
	mainCommand.AddCommand(etopo.MainCmd)
	mainCommand.AddCommand(wind.MainCmd)
}
