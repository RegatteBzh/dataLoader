package polar

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"

	redis "gopkg.in/redis.v4"

	"github.com/regattebzh/dataLoader/database"
	"github.com/sethgrid/multibar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dic map[string]string

func init() {
	flags := MainCmd.Flags()

	flags.String("polar-name", "polar", "Redis key name to store data")
	viper.BindPFlag("polar_name", flags.Lookup("polar-name"))

	flags.Bool("fake", false, "Do not write into redis. Just display")
	viper.BindPFlag("fake", flags.Lookup("fake"))

	dic = make(map[string]string)
	dic["1"] = "foc"
	dic["2"] = "spi"
	dic["4"] = "foc2"
	dic["8"] = "genois"
	dic["16"] = "zero-code"
	dic["32"] = "light-spi"
	dic["64"] = "gennaker"
}

func loadAllPolars(pathName string, redisName string, shipName string, fake bool) (err error) {
	var client *redis.Client

	if !fake {
		client = database.Open()
		defer client.Close()
	} else {
		fmt.Printf("Fake mode.\n")
	}

	progressBars, err := multibar.New()
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(pathName)
	if err != nil {
		log.Fatal(err)
	}

	filter, err := regexp.Compile(`(\d*)\.csv$`)
	if err != nil {
		log.Fatal(err)
	}

	progressBars.Printf("Loading Polars (%s) %s\n", shipName, pathName)
	progressBar := progressBars.MakeBar(180, "polar_"+shipName)

	for _, f := range files {
		match := filter.FindStringSubmatch(f.Name())
		if len(match) > 0 {
			filename := path.Join(pathName, f.Name())
			file, err := os.Open(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			loader(file, client, redisName+"_"+shipName+"_"+dic[match[1]], progressBar, fake)
		}
	}

	return
}

// MainCmd is the main command manager
var MainCmd = &cobra.Command{
	Use:   "polar <folder-path> <ship-name>",
	Short: "Load polars for a boat",
	Run: func(cmd *cobra.Command, args []string) {

		fake := viper.GetBool("fake")
		redisName := viper.GetString("polar_name")
		pathName := args[0]
		shipName := args[1]

		if len(args) < 2 {
			log.Fatal(errors.New("Not enough arguments."))
		}

		loadAllPolars(pathName, redisName, shipName, fake)
	},
}
