package main

import (
	"io/ioutil"
	"os"
	"sync"

	"github.com/itmecho/awsinventory/internal/loader"
	"github.com/sirupsen/logrus"

	"github.com/itmecho/awsinventory/internal/inventory"
	"github.com/spf13/pflag"
)

var (
	outputFile string
	regions    []string
	verbose    bool

	validRegions = []string{
		"us-east-2",
		"us-east-1",
		"us-west-1",
		"us-west-2",
		"ap-south-1",
		"ap-northeast-3",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ca-central-1",
		"cn-north-1",
		"cn-northwest-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"eu-north-1",
		"sa-east-1",
	}
)

func init() {
	pflag.StringVarP(&outputFile, "output-file", "o", "inventory.csv", "path to the output file")
	pflag.StringSliceVar(&regions, "regions", validRegions, "regions to gather data from")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "show verbose logging")
	pflag.Parse()

	if !verbose {
		logrus.SetOutput(ioutil.Discard)
	}
}

func main() {

	data := loader.NewLoader()
	wg := sync.WaitGroup{}

	go func() {
		for e := range data.Errors {
			logrus.Error(e)
		}
	}()

	for _, r := range regions {
		logrus.Infof("loading data for %s", r)
		wg.Add(1)
		go func(region string) {
			data.LoadEC2Instances(region)
			logrus.Infof("loaded data for %s", region)
			wg.Done()
		}(r)
	}

	wg.Wait()
	close(data.Errors)

	logrus.Info("finished loading data")

	logrus.Infof("creating/opening output file: %s", outputFile)
	f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Fatal(err)
	}
	defer f.Close()

	csv, err := inventory.NewCSV(f)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, r := range data.Data {
		if err := csv.WriteRow(r); err != nil {
			logrus.Error(err)
		}
	}

	logrus.Infof("writing %s", outputFile)
	csv.Flush()
}
