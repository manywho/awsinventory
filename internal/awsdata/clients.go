package awsdata

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// Clients is an interface for getting new AWS service clients
type Clients interface {
	GetEC2Client(region string) ec2iface.EC2API
	GetELBClient(region string) elbiface.ELBAPI
	GetIAMClient(region string) iamiface.IAMAPI
	GetRDSClient(region string) rdsiface.RDSAPI
	GetS3Client(region string) s3iface.S3API
}

// DefaultClients holds the default methods for creating AWS service clients
type DefaultClients struct{}

// GetEC2Client returns a new EC2 client for the given region
func (c DefaultClients) GetEC2Client(region string) ec2iface.EC2API {
	return ec2.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetELBClient returns a new ELB client for the given region
func (c DefaultClients) GetELBClient(region string) elbiface.ELBAPI {
	return elb.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetIAMClient returns a new IAM client for the given region
func (c DefaultClients) GetIAMClient(region string) iamiface.IAMAPI {
	return iam.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetRDSClient returns a new RDS client for the given region
func (c DefaultClients) GetRDSClient(region string) rdsiface.RDSAPI {
	return rds.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetS3Client returns a new S3 client for the given region
func (c DefaultClients) GetS3Client(region string) s3iface.S3API {
	return s3.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}
