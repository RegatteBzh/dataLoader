package gshhg

import (
	"fmt"
	"log"
	"os"

	redis "gopkg.in/redis.v4"

	"github.com/regattebzh/dataLoader/database"
	"github.com/sethgrid/multibar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	flags := MainCmd.Flags()

	flags.Bool("fake", false, "Do not write into redis. Just display")
	viper.BindPFlag("fake", flags.Lookup("fake"))

}

// MainCmd is the main command manager
var MainCmd = &cobra.Command{
	Use:   "gshhg <path>",
	Short: "Load coast data",
	Run: func(cmd *cobra.Command, args []string) {
		var client *redis.Client

		fake := viper.GetBool("fake")

		if !fake {
			client = database.Open()
			defer client.Close()
		} else {
			fmt.Printf("Fake mode.\n")
		}

		if len(args) == 0 {
			log.Fatal("Missing file")
		}

		progressBars, err := multibar.New()
		if err != nil {
			log.Fatal(err)
		}

		progressBars.Printf("Loading Coast %s\n", args[0])
		progressBar := progressBars.MakeBar(100, "Coast")

		file, err := os.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		if !fake {
			go progressBars.Listen()
		}

		err = Loader(file, client, progressBar, fake)

	},
}
