package awsdata

import (
	"sync"
	"time"

	"github.com/manywho/awsinventory/internal/inventory"
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

// Data is responsible for concurrently loading data from AWS and storing it based on the regions and services provided
type Data struct {
	clients Clients
	rows    []inventory.Row
	results chan result
	regions []string
	log     *logrus.Logger
	lock    sync.Mutex
	wg      sync.WaitGroup
}

// New returns a new empty Data
func New(logger *logrus.Logger, clients Clients) *Data {
	if clients == nil {
		clients = DefaultClients{}
	}

	return &Data{
		clients: clients,
		rows:    make([]inventory.Row, 0),
		results: make(chan result),
		log:     logger,
		lock:    sync.Mutex{},
		wg:      sync.WaitGroup{},
	}
}

// Load concurrently loads the required data based of the of regions and services provided
func (d *Data) Load(regions, services []string) {
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

	go d.startWorker()

	// Global services
	if stringInSlice(ServiceIAM, services) {
		go d.loadIAMUsers(d.clients.GetIAMClient(ValidRegions[0]))
	}

	if stringInSlice(ServiceS3, services) {
		go d.loadS3Buckets(d.clients.GetS3Client(ValidRegions[0]))
	}

	// Regional Services
	for _, region := range regions {

		if stringInSlice(ServiceEC2, services) {
			go d.loadEC2Instances(d.clients.GetEC2Client(region), region)
		}

		if stringInSlice(ServiceEBS, services) {
			// EBS volumes are part of the EC2 api and therefore require an EC2 client
			go d.loadEBSVolumes(d.clients.GetEC2Client(region), region)
		}

		if stringInSlice(ServiceELB, services) {
			go d.loadELBs(d.clients.GetELBClient(region), region)
		}

		if stringInSlice(ServiceRDS, services) {
			go d.loadRDSInstances(d.clients.GetRDSClient(region), region)
		}
	}

	// Delay to give the first wg.Add call time to complete
	time.Sleep(100 * time.Millisecond)

	d.wg.Wait()
	close(d.results)
}

func (d *Data) startWorker() {
	d.log.Info("starting worker")
	for {
		res, ok := <-d.results
		if !ok {
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

func (d *Data) appendRow(row inventory.Row) {
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
