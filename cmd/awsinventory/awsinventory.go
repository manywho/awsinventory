package main

import (
	"os"
	"sync"

	"github.com/itmecho/awsinventory/internal/awsmetadata"

	"github.com/itmecho/awsinventory/internal/loader"
	"github.com/sirupsen/logrus"

	"github.com/itmecho/awsinventory/internal/inventory"
	"github.com/spf13/pflag"
)

var (
	outputFile string
	regions    []string
)

func init() {
	pflag.StringVarP(&outputFile, "output-file", "o", "inventory.csv", "path to the output file")
	pflag.StringSliceVar(&regions, "regions", awsmetadata.Regions, "regions to gather data from")
	pflag.Parse()

	logrus.SetLevel(logrus.DebugLevel)
}

func main() {

	data := loader.NewLoader()
	wg := sync.WaitGroup{}

	logrus.Debug("starting region loop")
	for _, r := range regions {
		logrus.Infof("loading data for %s", r)
		wg.Add(1)
		go func(region string) {
			data.LoadEC2Instances(region)
			wg.Done()
		}(r)
	}

	logrus.Debug("waiting for loop to finish")

	wg.Wait()
	close(data.Errors)

	logrus.Debug("checking for errors")

	var wasError bool
	for e := range data.Errors {
		if e != nil {
			wasError = true
			logrus.Error(e)
		}
	}
	if wasError {
		logrus.Fatal("errors recieving data")
	}

	f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Fatal(err)
	}
	defer f.Close()

	logrus.Debug("creating new csv")

	csv, err := inventory.NewCSV(f)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, r := range data.Data {
		if err := csv.WriteRow(r); err != nil {
			logrus.Error(err)
		}
	}

	logrus.Debug("writing csv to disk")
	csv.Flush()
}
