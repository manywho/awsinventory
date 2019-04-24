package loader

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/itmecho/awsinventory/internal/inventory"
)

// LoadEC2Volumes loads the ec2 volume data from the given region into the Loader's data
func (l *Loader) LoadEC2Volumes(ec2Svc ec2iface.EC2API, region string) {
	out, err := ec2Svc.DescribeVolumes(&ec2.DescribeVolumesInput{})
	if err != nil {
		l.Errors <- err
		return
	}

	results := make([]inventory.Row, 0)

	for _, v := range out.Volumes {
		var name string
		for _, t := range v.Tags {
			if aws.StringValue(t.Key) == "Name" {
				name = aws.StringValue(t.Value)
			}
		}

		results = append(results, inventory.Row{
			ID:           aws.StringValue(v.VolumeId),
			AssetType:    "EC2 Volume",
			Location:     aws.StringValue(v.AvailabilityZone),
			CreationDate: aws.TimeValue(v.CreateTime),
			Application:  name,
			Hardware:     fmt.Sprintf("%s (%dGB)", aws.StringValue(v.VolumeType), aws.Int64Value(v.Size)),
			// TODO add v.Encrypted
		})
	}

	l.appendData(results)
}
