package wind

import (
	"log"
	"os"

	"github.com/regattebzh/dataLoader/database"
	"github.com/sethgrid/multibar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	flags := MainCmd.Flags()

	flags.String("wind-name", "wind", "Redis key name to store data")
	viper.BindPFlag("wind_name", flags.Lookup("wind-name"))

	flags.Bool("fake", false, "Do not write into redis. Just display")
	viper.BindPFlag("fake", flags.Lookup("fake"))

}

// MainCmd is the main command manager
var MainCmd = &cobra.Command{
	Use:   "wind <path>",
	Short: "Load wind forecast",
	Run: func(cmd *cobra.Command, args []string) {

		fake := viper.GetBool("fake")

		client := database.Open()
		defer client.Close()

		if len(args) == 0 {
			log.Fatal("Missing file")
		}

		progressBars, err := multibar.New()
		if err != nil {
			log.Fatal(err)
		}

		progressBars.Printf("Loading WindForecast %s\n", args[0])
		progressBar := progressBars.MakeBar(180, "Wind")

		file, err := os.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		if !fake {
			go progressBars.Listen()
		}

		err = Loader(file, client, viper.GetString("wind_name"), progressBar, fake)

	},
}
