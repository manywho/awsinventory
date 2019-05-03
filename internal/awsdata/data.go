package awsdata

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/manywho/awsinventory/pkg/route53cache"
	"github.com/sirupsen/logrus"
)

var (
	// ValidRegions contains a list of valid AWS regions to gather data from
	ValidRegions = []string{
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

	// ValidServices contains a list of valid AWS services to gather data from
	ValidServices = []string{
		ServiceEBS,
		ServiceEC2,
		ServiceELB,
		ServiceIAM,
		ServiceRDS,
		ServiceS3,
	}
)

type result struct {
	Row inventory.Row
	Err error
}

// AWSData is responsible for concurrently loading data from AWS and storing it based on the regions and services provided
type AWSData struct {
	clients      Clients
	rows         []inventory.Row
	results      chan result
	done         chan bool
	regions      []string
	route53Cache *route53cache.Cache
	log          *logrus.Logger
	lock         sync.Mutex
	wg           sync.WaitGroup
}

// New returns a new empty AWSData
func New(logger *logrus.Logger, clients Clients) *AWSData {
	if clients == nil {
		clients = DefaultClients{}
	}

	return &AWSData{
		clients: clients,
		rows:    make([]inventory.Row, 0),
		results: make(chan result),
		done:    make(chan bool, 1),
		log:     logger,
		lock:    sync.Mutex{},
		wg:      sync.WaitGroup{},
	}
}

// Load concurrently loads the required data based of the of regions and services provided
func (d *AWSData) Load(regions, services []string) {
	if len(services) == 0 {
		d.log.Error(ErrNoServices)
		return
	}

	if len(regions) == 0 && hasRegionalServices(services) {
		d.log.Error(ErrNoRegions)
		return
	}

	if err := validateRegions(regions); err != nil {
		d.log.Error(err)
		return
	}

	if err := validateServices(services); err != nil {
		d.log.Error(err)
		return
	}

	if stringInSlice(ServiceEC2, services) {
		d.loadRoute53Data()
	}

	go d.startWorker()

	// Global services
	if stringInSlice(ServiceIAM, services) {
		d.wg.Add(1)
		go d.loadIAMUsers()
	}

	if stringInSlice(ServiceS3, services) {
		d.wg.Add(1)
		go d.loadS3Buckets()
	}

	// Regional Services
	for _, region := range regions {

		if stringInSlice(ServiceEC2, services) {
			d.wg.Add(1)
			go d.loadEC2Instances(region)
		}

		if stringInSlice(ServiceEBS, services) {
			d.wg.Add(1)
			go d.loadEBSVolumes(region)
		}

		if stringInSlice(ServiceELB, services) {
			d.wg.Add(1)
			go d.loadELBs(region)
		}

		if stringInSlice(ServiceRDS, services) {
			d.wg.Add(1)
			go d.loadRDSInstances(region)
		}
	}

	d.wg.Wait()
	close(d.results)

	<-d.done
}

func (d *AWSData) startWorker() {
	d.log.Info("starting worker")
	for {
		res, ok := <-d.results
		var blankResult result
		if res == blankResult && !ok {
			d.done <- true
			return
		}
		if res.Err != nil {
			d.log.Debugf("worker recieved an error")
			d.log.Error(res.Err)
		} else {
			d.log.Debugf("worker recieved a row")
			d.appendRow(res.Row)
		}
	}
}

func (d *AWSData) loadRoute53Data() {
	d.log.Info("loading route53 data")
	r53 := d.clients.GetRoute53Client(ValidRegions[0])
	zones, err := r53.ListHostedZones(&route53.ListHostedZonesInput{})
	if err != nil {
		d.log.Fatal(err)
	}
	var sets []*route53.ResourceRecordSet

	var lock sync.Mutex
	var wg sync.WaitGroup
	for _, z := range zones.HostedZones {
		wg.Add(1)
		go func(zone *route53.HostedZone) {
			d.log.Debugf("loading route53 records for hosted zone %s", aws.StringValue(zone.Name))

			r53Client := d.clients.GetRoute53Client(ValidRegions[0])

			out, err := r53Client.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
				HostedZoneId: zone.Id,
			})
			if err != nil {
				d.log.Fatal(err)
			}

			lock.Lock()
			sets = append(sets, out.ResourceRecordSets...)
			lock.Unlock()
			wg.Done()
		}(z)
	}

	wg.Wait()

	d.route53Cache = route53cache.New(sets)
}

func (d *AWSData) appendRow(row inventory.Row) {
	d.lock.Lock()
	d.rows = append(d.rows, row)
	d.lock.Unlock()
}

func stringInSlice(needle string, haystack []string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}

	return false
}

func hasRegionalServices(services []string) bool {
	for _, service := range services {
		if service == ServiceEBS || service == ServiceEC2 || service == ServiceELB {
			return true
		}
	}
	return false
}

func validateRegions(regions []string) error {
	for _, region := range regions {
		if !stringInSlice(region, ValidRegions) {
			return newErrInvalidRegion(region)
		}
	}
	return nil
}

func validateServices(services []string) error {
	for _, service := range services {
		if !stringInSlice(service, ValidServices) {
			return newErrInvalidService(service)
		}
	}
	return nil
}
