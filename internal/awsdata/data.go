package awsdata

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"

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
	rows          chan inventory.Row
	regions       []string
	validRegions  []string
	validServices []string
	route53Cache  *route53cache.Cache
	log           *logrus.Logger
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
		ServiceCodeCommit,
		ServiceDynamoDB,
		ServiceEBS,
		ServiceEC2,
		ServiceECR,
		ServiceECS,
		ServiceElastiCache,
		ServiceElasticsearchService,
		ServiceELB,
		ServiceELBV2,
		ServiceIAM,
		ServiceKMS,
		ServiceLambda,
		ServiceRDS,
		ServiceS3,
		ServiceSQS,
		ServiceWorkSpace,
	}

	return &AWSData{
		clients:       clients,
		validRegions:  regions,
		validServices: services,
		rows:          make(chan inventory.Row, 100),
		log:           logger,
		wg:            sync.WaitGroup{},
	}
}

// Load concurrently the required data based on the regions and services provided
func (d *AWSData) Load(regions, services []string, processRow ProcessRow) {
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

	if processRow == nil {
		processRow = func(row inventory.Row) error {
			d.log.Debugf("throwing away %s: %s", row.AssetType, row.UniqueAssetIdentifier)
			return nil
		}
	}

	done := make(chan bool, 1)
	d.log.Debug("starting row processing process")
	go d.startWorker(processRow, done)

	if stringInSlice(ServiceEC2, services) {
		d.loadRoute53Data()
	}

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
		if stringInSlice(ServiceCodeCommit, services) {
			d.log.Debug("including CodeCommit service")
			d.wg.Add(1)
			go d.loadCodeCommitRepositories(region)
		}

		if stringInSlice(ServiceDynamoDB, services) {
			d.log.Debug("including DynamoDB service")
			d.wg.Add(1)
			go d.loadDynamoDBTables(region)
		}

		if stringInSlice(ServiceEBS, services) {
			d.log.Debug("including EBS service")
			d.wg.Add(1)
			go d.loadEBSVolumes(region)
		}

		if stringInSlice(ServiceEC2, services) {
			d.log.Debug("including EC2 service")
			d.wg.Add(1)
			go d.loadEC2Instances(region)
		}

		if stringInSlice(ServiceECR, services) {
			d.log.Debug("including ECR service")
			d.wg.Add(1)
			go d.loadECRImages(region)
		}

		if stringInSlice(ServiceECS, services) {
			d.log.Debug("including ECS service")
			d.wg.Add(1)
			go d.loadECSContainers(region)
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

		if stringInSlice(ServiceKMS, services) {
			d.log.Debug("including KMS service")
			d.wg.Add(1)
			go d.loadKMSKeys(region)
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

		if stringInSlice(ServiceSQS, services) {
			d.log.Debug("including SQS service")
			d.wg.Add(1)
			go d.loadSQSQueues(region)
		}

		if stringInSlice(ServiceWorkSpace, services) {
			d.log.Debug("including WorkSpace service")
			d.wg.Add(1)
			go d.loadWorkSpacesInstances(region)
		}

	}

	d.wg.Wait()
	close(d.rows)
	d.log.Info("all data loaded")

	<-done
	d.log.Info("all rows processed")
}

func (d *AWSData) startWorker(processRow ProcessRow, done chan bool) {
	var blankRow inventory.Row
	for {
		row, ok := <-d.rows
		if row == blankRow || !ok {
			done <- true
			return
		}

		d.log.Debugf("processing %s: %s", row.AssetType, row.UniqueAssetIdentifier)

		if err := processRow(row); err != nil {
			d.log.Errorf("process row function failed: %s", err)
		}
	}
}

// PrintRegions lists all available AWS regions as used by the command line `print-regions` option
func (d *AWSData) PrintRegions() {
	for _, r := range d.validRegions {
		println(r)
	}
}

func (d *AWSData) loadRoute53Data() {
	route53Svc := d.clients.GetRoute53Client(DefaultRegion)
	d.log.Info("loading hosted zones")
	var zones []*route53.HostedZone
	done := false
	params := &route53.ListHostedZonesInput{}
	for !done {
		out, err := route53Svc.ListHostedZones(params)
		if err != nil {
			d.log.Fatal(err)
		}

		zones = append(zones, out.HostedZones...)

		if out.IsTruncated == aws.Bool(true) {
			params.Marker = out.NextMarker
		} else {
			done = true
		}
	}

	d.log.Infof("found %d hosted zones", len(zones))

	var sets []*route53.ResourceRecordSet

	var lock sync.Mutex
	var wg sync.WaitGroup
	for _, z := range zones {
		wg.Add(1)
		go func(route53Svc route53iface.Route53API, zone *route53.HostedZone) {
			d.log.Infof("loading route53 records for hosted zone %s (%s)", aws.StringValue(zone.Name), aws.StringValue(zone.Id))

			done := false
			params := &route53.ListResourceRecordSetsInput{
				HostedZoneId: zone.Id,
			}
			for !done {
				out, err := route53Svc.ListResourceRecordSets(params)
				if err != nil {
					d.log.Fatal(err)
				}

				d.log.Infof("found %d records in hosted zone %s (%s)", len(out.ResourceRecordSets), aws.StringValue(zone.Name), aws.StringValue(zone.Id))

				lock.Lock()
				sets = append(sets, out.ResourceRecordSets...)
				lock.Unlock()

				if out.IsTruncated == aws.Bool(true) {
					params.StartRecordIdentifier = out.NextRecordIdentifier
					params.StartRecordName = out.NextRecordName
					params.StartRecordType = out.NextRecordType
				} else {
					done = true
				}
			}
			wg.Done()
		}(route53Svc, z)
	}

	wg.Wait()

	d.route53Cache = route53cache.New(sets)
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
