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
}

// MainCmd is the main command manager
var MainCmd = &cobra.Command{
	Use:   "etopo <path>",
	Short: "Load etopo 1 minute grid path",
	Run: func(cmd *cobra.Command, args []string) {

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

		go progressBars.Listen()

		err = Loader1Minute(file, client, "etopo", int16(viper.GetInt64("etopo_threshold")), progressBar)

	},
}
