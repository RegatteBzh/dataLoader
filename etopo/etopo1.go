package etopo

import (
	"io"

	"github.com/sethgrid/multibar"

	"gopkg.in/redis.v3"
)

// ETOPO1WIDTH is the length of a data line in bytes
const ETOPO1WIDTH = 21601

// ETOTO1HEIGHT is the number of lines of data
const ETOTO1HEIGHT = 10801

// Loader1Minute load Etopo 1 minute data to Redis
func Loader1Minute(file io.Reader, c *redis.Client, redisName string, threshold int16, progressBar multibar.ProgressFunc, fake bool) error {
	return Loader(file, ETOPO1WIDTH, ETOTO1HEIGHT, c, redisName, threshold, progressBar, fake)
}
