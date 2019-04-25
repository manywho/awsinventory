package data

import (
	"errors"
	"sync"
	"time"

	"github.com/itmecho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

var (
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

	ValidServices = []string{
		ServiceEBS,
		ServiceEC2,
		ServiceELB,
		ServiceIAM,
		ServiceS3,
	}

	// ErrNoRegions is logged when no regions are given to the Load method
	ErrNoRegions = errors.New("no regions specified")

	// ErrNoServices is logged when no services are given to the Load method
	ErrNoServices = errors.New("no services specified")
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

// Load concurrently loads the required data as specified during creation
func (d *Data) Load(regions, services []string) {
	if len(regions) == 0 {
		d.log.Error(ErrNoRegions)
		return
	}

	if len(services) == 0 {
		d.log.Error(ErrNoServices)
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
			go d.loadEBSVolumes(d.clients.GetEC2Client(region), region)
		}

		if stringInSlice(ServiceELB, services) {
			go d.loadELBs(d.clients.GetELBClient(region), region)
		}
	}

	time.Sleep(500 * time.Millisecond)

	d.wg.Wait()
	close(d.results)
}

func (d *Data) startWorker() {
	d.log.Info("starting worker")
	for {
		res, ok := <-d.results
		if !ok {
			d.log.Debug("results channel closed")
			return
		}
		d.log.Debugf("worker recieved a result")
		if res.Err != nil {
			d.log.Error(res.Err)
		} else {
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
