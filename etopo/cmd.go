package etopo

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

	flags.Int64("etopo-threshold", 0, "Altitude for ground limit in meter")
	viper.BindPFlag("etopo_threshold", flags.Lookup("etopo-threshold"))

	flags.String("etopo-name", "etopo", "Redis key name to store data")
	viper.BindPFlag("etopo_name", flags.Lookup("etopo-name"))

	flags.Bool("fake", false, "Do not write into redis. Just display")
	viper.BindPFlag("fake", flags.Lookup("fake"))
}

// MainCmd is the main command manager
var MainCmd = &cobra.Command{
	Use:   "etopo <path>",
	Short: "Load etopo 1 minute grid path",
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

		progressBars.Printf("Loading ETOPO1 %s\n", args[0])
		progressBar := progressBars.MakeBar(ETOTO1HEIGHT, "Etopo1")
		file, err := os.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		if !fake {
			go progressBars.Listen()
		}

		err = Loader1Minute(file, client, viper.GetString("etopo_name"), int16(viper.GetInt("etopo_threshold")), progressBar, fake)
	},
}
