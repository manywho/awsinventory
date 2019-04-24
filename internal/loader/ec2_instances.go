package loader

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/itmecho/awsinventory/internal/inventory"
)

// LoadEC2Instances loads the ec2 data from the given region into the Loader's data
func (l *Loader) LoadEC2Instances(ec2Svc ec2iface.EC2API, region string) {
	out, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{})
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

			results = append(results, inventory.Row{
				ID:           aws.StringValue(i.InstanceId),
				AssetType:    "EC2 Instance",
				Location:     region,
				CreationDate: aws.TimeValue(i.LaunchTime),
				Application:  name,
				Hardware:     aws.StringValue(i.InstanceType),
				Baseline:     aws.StringValue(i.ImageId),
				InternalIP:   aws.StringValue(i.PrivateIpAddress),
				ExternalIP:   aws.StringValue(i.PublicIpAddress),
				VPCID:        aws.StringValue(i.VpcId),
			})
		}
	}

	l.appendData(results)
}
