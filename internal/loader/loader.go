package loader

import (
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/itmecho/awsinventory/internal/inventory"
)

// Loader is responsible for loading data from the AWS API and storing it.
type Loader struct {
	sess   *session.Session
	lock   sync.Mutex
	Data   []inventory.Row
	Errors chan error
}

// NewLoader returns a new Loader with a preloaded AWS session and lock
func NewLoader() *Loader {
	return &Loader{
		sess:   session.Must(session.NewSession()),
		lock:   sync.Mutex{},
		Data:   make([]inventory.Row, 0),
		Errors: make(chan error),
	}
}

// LoadEC2Instances loads the ec2 data from the given region into the Loader's data
func (l *Loader) LoadEC2Instances(region string) {
	ec2Svc := ec2.New(l.sess, &aws.Config{
		Region: aws.String(region),
	})

	out, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{MaxResults: aws.Int64(10)})
	if err != nil {
		l.Errors <- err
		return
	}

	results := make([]inventory.Row, 0)

	for _, r := range out.Reservations {
		for _, i := range r.Instances {
			var name string
			for _, tag := range i.Tags {
				if *tag.Key == "Name" {
					name = aws.StringValue(tag.Value)
				}
			}

			var internalIPs []string
			for _, ni := range i.NetworkInterfaces {
				internalIPs = append(internalIPs, aws.StringValue(ni.PrivateIpAddress))
			}

			results = append(results, inventory.Row{
				ID:          aws.StringValue(i.InstanceId),
				AssetType:   "EC2 Instance",
				Location:    region,
				Application: name,
				Hardware:    aws.StringValue(i.InstanceType),
				Baseline:    aws.StringValue(i.ImageId),
				InternalIP:  strings.Join(internalIPs, " "),
				VPCID:       aws.StringValue(i.VpcId),
			})
		}
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	for _, row := range results {
		l.Data = append(l.Data, row)
	}
}
