package wind

import (
	"log"
	"os"

	"github.com/regattebzh/dataLoader/database"
	"github.com/sethgrid/multibar"
	"github.com/spf13/cobra"
)

func init() {
	//flags := MainCmd.Flags()

}

// MainCmd is the main command manager
var MainCmd = &cobra.Command{
	Use:   "wind <path>",
	Short: "Load wind forecast",
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

		progressBars.Printf("Loading WindForecast %s\n", args[0])
		progressBar := progressBars.MakeBar(180, "Wind")

		file, err := os.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		go progressBars.Listen()

		err = Loader(file, client, "wind", progressBar)

	},
}
