package main

import (
	"io/ioutil"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
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

	// Start watching for errors
	go func() {
		for e := range data.Errors {
			logrus.Error(e)
		}
	}()

	// Create a new aws session
	awsSess := session.Must(session.NewSession())

	// Create global services
	s3Svc := s3.New(awsSess, &aws.Config{Region: aws.String(regions[0])})
	iamSvc := iam.New(awsSess, &aws.Config{Region: aws.String(regions[0])})

	// Concurrently load S3 bucket data
	logrus.Info("loading s3 data")
	wg.Add(1)
	go func() {
		data.LoadS3Buckets(s3Svc)
		wg.Done()
	}()

	// Concurrently load S3 bucket data
	logrus.Info("loading s3 data")
	wg.Add(1)
	go func() {
		data.LoadIAMUsers(iamSvc)
		wg.Done()
	}()

	// Loop over regions and load data
	for _, r := range regions {
		logrus.Infof("loading data for %s", r)

		// Create new services for current region
		ec2Svc := ec2.New(awsSess, &aws.Config{Region: aws.String(r)})

		// Concurrently load instance data
		wg.Add(1)
		go func(region string) {
			data.LoadEC2Instances(ec2Svc, region)
			logrus.Infof("loaded data for %s", region)
			wg.Done()
		}(r)
	}

	// Wait for data to finish loading
	wg.Wait()

	// Close errors channel
	close(data.Errors)

	logrus.Info("finished loading data")

	// Create or open the output file
	logrus.Infof("creating/opening output file: %s", outputFile)
	f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Fatal(err)
	}
	defer f.Close()

	// Create new csv inventory
	csv, err := inventory.NewCSV(f)
	if err != nil {
		logrus.Fatal(err)
	}

	// Write stored rows to csv inventory
	for _, r := range data.Data {
		if err := csv.WriteRow(r); err != nil {
			logrus.Error(err)
		}
	}

	// Write file to disk
	logrus.Infof("writing %s", outputFile)
	csv.Flush()
}
