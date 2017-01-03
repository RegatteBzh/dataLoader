package wind

import (
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

type fileLoader struct {
	Filename string
	Handler  *os.File
	WindsU   []float32
	WindsV   []float32
}

func init() {
	flags := MainCmd.Flags()

	flags.String("wind-name", "wind", "Redis key name to store data")
	viper.BindPFlag("wind_name", flags.Lookup("wind-name"))

	flags.Bool("fake", false, "Do not write into redis. Just display")
	viper.BindPFlag("fake", flags.Lookup("fake"))

}

func loadAllWinds(windPath string, redisName string, fake bool) (err error) {

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

	files, err := ioutil.ReadDir(windPath)
	if err != nil {
		log.Fatal(err)
	}

	filter, err := regexp.Compile(`^gfs\..*\.f(\d{3})\.bin$`)
	if err != nil {
		log.Fatal(err)
	}

	progressBars.Printf("Loading WindForecast %s\n", windPath)
	progressBar := progressBars.MakeBar(180, "Wind")

	filenames := make(map[string]*fileLoader)
	for _, f := range files {
		match := filter.FindStringSubmatch(f.Name())
		if len(match) > 0 {
			filename := path.Join(windPath, f.Name())
			file, err := os.Open(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			windLoader := fileLoader{
				Handler:  file,
				Filename: filename,
			}
			filenames[match[1]] = &windLoader
		}
	}

	if !fake {
		go progressBars.Listen()
	}

	err = LoadAll(filenames, client, redisName, progressBar, fake)

	return

}

// MainCmd is the main command manager
var MainCmd = &cobra.Command{
	Use:   "wind <folder-path>",
	Short: "Load wind forecast",
	Run: func(cmd *cobra.Command, args []string) {

		fake := viper.GetBool("fake")

		loadAllWinds(args[0], viper.GetString("wind_name"), fake)

	},
}
