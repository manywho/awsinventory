package awsdata

import (
	"fmt"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/manywho/awsinventory/pkg/route53cache"
	"github.com/sirupsen/logrus"
)

var (
	// DefaultRegion contains the region used by default and in tests
	DefaultRegion = "us-east-1"
)

type result struct {
	Row inventory.Row
	Err error
}

// AWSData is responsible for concurrently loading data from AWS and storing it based on the regions and services provided
type AWSData struct {
	clients       Clients
	rows          []inventory.Row
	results       chan result
	done          chan bool
	regions       []string
	validRegions  []string
	validServices []string
	route53Cache  *route53cache.Cache
	log           *logrus.Logger
	lock          sync.Mutex
	wg            sync.WaitGroup
}

// New returns a new default AWSData
func New(logger *logrus.Logger, clients Clients) *AWSData {
	if clients == nil {
		clients = DefaultClients{}
	}

	// List of valid AWS regions to gather data from
	var regions []string
	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()
	for _, p := range partitions {
		for id := range p.Regions() {
			regions = append(regions, id)
		}
	}

	// List of valid AWS services to gather data from
	var services = []string{
		ServiceCloudFront,
		ServiceDynamoDB,
		ServiceEBS,
		ServiceEC2,
		ServiceECS,
		ServiceElastiCache,
		ServiceElasticsearchService,
		ServiceELB,
		ServiceELBV2,
		ServiceIAM,
		ServiceLambda,
		ServiceRDS,
		ServiceS3,
	}

	return &AWSData{
		clients:       clients,
		validRegions:  regions,
		validServices: services,
		rows:          make([]inventory.Row, 0),
		results:       make(chan result),
		done:          make(chan bool, 1),
		log:           logger,
		lock:          sync.Mutex{},
		wg:            sync.WaitGroup{},
	}
}

// Load concurrently the required data based on the regions and services provided
func (d *AWSData) Load(regions, services []string) {
	if len(services) == 0 {
		services = d.validServices
	}

	if len(regions) == 0 && hasRegionalServices(services) {
		d.log.Error(ErrNoRegions)
		return
	}

	if err := d.validateRegions(regions); err != nil {
		d.log.Error(err)
		return
	}

	if err := d.validateServices(services); err != nil {
		d.log.Error(err)
		return
	}

	if stringInSlice(ServiceEC2, services) {
		d.loadRoute53Data()
	}

	go d.startWorker()

	// Global services
	if stringInSlice(ServiceCloudFront, services) {
		d.log.Debug("including CloudFront service")
		d.wg.Add(1)
		go d.loadCloudFrontDistributions()
	}

	if stringInSlice(ServiceIAM, services) {
		d.log.Debug("including IAM service")
		d.wg.Add(1)
		go d.loadIAMUsers()
	}

	// Regional Services
	for _, region := range regions {
		if stringInSlice(ServiceDynamoDB, services) {
			d.log.Debug("including DynamoDB service")
			d.wg.Add(1)
			go d.loadDynamoDBTables(region)
		}

		if stringInSlice(ServiceEC2, services) {
			d.log.Debug("including EC2 service")
			d.wg.Add(1)
			go d.loadEC2Instances(region)
		}

		if stringInSlice(ServiceECS, services) {
			d.log.Debug("including ECS service")
			d.wg.Add(1)
			go d.loadECSContainers(region)
		}

		if stringInSlice(ServiceEBS, services) {
			d.log.Debug("including EBS service")
			d.wg.Add(1)
			go d.loadEBSVolumes(region)
		}

		if stringInSlice(ServiceElastiCache, services) {
			d.log.Debug("including ElastiCache service")
			d.wg.Add(1)
			go d.loadElastiCacheNodes(region)
		}

		if stringInSlice(ServiceElasticsearchService, services) {
			d.log.Debug("including Elasticsearch service")
			d.wg.Add(1)
			go d.loadElasticsearchDomains(region)
		}

		if stringInSlice(ServiceELB, services) {
			d.log.Debug("including ELB service")
			d.wg.Add(1)
			go d.loadELBs(region)
		}

		if stringInSlice(ServiceELBV2, services) {
			d.log.Debug("including ELB v2 service")
			d.wg.Add(1)
			go d.loadELBV2s(region)
		}

		if stringInSlice(ServiceLambda, services) {
			d.log.Debug("including Lambda service")
			d.wg.Add(1)
			go d.loadLambdaFunctions(region)
		}

		if stringInSlice(ServiceRDS, services) {
			d.log.Debug("including RDS service")
			d.wg.Add(1)
			go d.loadRDSInstances(region)
		}

		if stringInSlice(ServiceS3, services) {
			d.log.Debug("including S3 service")
			d.wg.Add(1)
			go d.loadS3Buckets(region)
		}
	}

	d.wg.Wait()
	close(d.results)

	<-d.done
}

// PrintRegions lists all available AWS regions as used by the command line `print-regions` option
func (d *AWSData) PrintRegions() {
	for _, r := range d.validRegions {
		println(r)
	}
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
			d.log.Error(res.Err)
		} else {
			d.log.Debugf("worker received an %s: %s", res.Row.AssetType, res.Row.UniqueAssetIdentifier)
			d.appendRow(res.Row)
		}
	}
}

func (d *AWSData) loadRoute53Data() {
	r53 := d.clients.GetRoute53Client(DefaultRegion)
	d.log.Info("loading hosted zones")
	zones, err := r53.ListHostedZones(&route53.ListHostedZonesInput{})
	if err != nil {
		d.log.Fatal(err)
	}

	d.log.Infof("found %d hosted zones", len(zones.HostedZones))

	var sets []*route53.ResourceRecordSet

	var lock sync.Mutex
	var wg sync.WaitGroup
	for _, z := range zones.HostedZones {
		wg.Add(1)
		go func(zone *route53.HostedZone) {
			d.log.Infof("loading route53 records for hosted zone %s", aws.StringValue(zone.Name))

			r53Client := d.clients.GetRoute53Client(DefaultRegion)

			out, err := r53Client.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
				HostedZoneId: zone.Id,
			})
			if err != nil {
				d.log.Fatal(err)
			}

			d.log.Infof("found %d records in hosted zone %s", len(out.ResourceRecordSets), aws.StringValue(zone.Name))

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

// SortRows takes all the rows in the inventory and sorts based on the UniqueAssetIdentifier (generally, the ARN)
func (d *AWSData) SortRows() {
	d.lock.Lock()
	sort.SliceStable(d.rows, func(i, j int) bool {
		return d.rows[i].UniqueAssetIdentifier < d.rows[j].UniqueAssetIdentifier
	})
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

func appendIfMissing(slice []string, s string) []string {
	for _, ele := range slice {
		if ele == s {
			return slice
		}
	}
	return append(slice, s)
}

func humanReadableBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func hasRegionalServices(services []string) bool {
	var globalServices = []string{
		ServiceCloudFront,
		ServiceIAM,
	}

	for _, service := range services {
		if !stringInSlice(service, globalServices) {
			return true
		}
	}
	return false
}

func (d *AWSData) validateRegions(regions []string) error {
	for _, region := range regions {
		if !stringInSlice(region, d.validRegions) {
			return newErrInvalidRegion(region)
		}
	}
	return nil
}

func (d *AWSData) validateServices(services []string) error {
	for _, service := range services {
		if !stringInSlice(service, d.validServices) {
			return newErrInvalidService(service)
		}
	}
	return nil
}
