package data_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

var testError = errors.New("test aws error")

type TestClients struct {
	EC2 ec2iface.EC2API
	ELB elbiface.ELBAPI
	IAM iamiface.IAMAPI
	S3  s3iface.S3API
}

func (c TestClients) GetEC2Client(region string) ec2iface.EC2API {
	return c.EC2
}

func (c TestClients) GetIAMClient(region string) iamiface.IAMAPI {
	return c.IAM
}

func (c TestClients) GetELBClient(region string) elbiface.ELBAPI {
	return c.ELB
}

func (c TestClients) GetS3Client(region string) s3iface.S3API {
	return c.S3
}
