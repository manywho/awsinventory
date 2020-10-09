package awsdata

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice/elasticsearchserviceiface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// Clients is an interface for getting new AWS service clients
type Clients interface {
	GetDynamoDBClient(region string) dynamodbiface.DynamoDBAPI
	GetEC2Client(region string) ec2iface.EC2API
	GetECSClient(region string) ecsiface.ECSAPI
	GetElastiCacheClient(region string) elasticacheiface.ElastiCacheAPI
	GetElasticsearchServiceClient(region string) elasticsearchserviceiface.ElasticsearchServiceAPI
	GetELBClient(region string) elbiface.ELBAPI
	GetELBV2Client(region string) elbv2iface.ELBV2API
	GetIAMClient(region string) iamiface.IAMAPI
	GetRDSClient(region string) rdsiface.RDSAPI
	GetRoute53Client(region string) route53iface.Route53API
	GetS3Client(region string) s3iface.S3API
}

// DefaultClients holds the default methods for creating AWS service clients
type DefaultClients struct{}

// GetDynamoDBClient returns a new DynamoDB client for the given region
func (c DefaultClients) GetDynamoDBClient(region string) dynamodbiface.DynamoDBAPI {
	return dynamodb.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetEC2Client returns a new EC2 client for the given region
func (c DefaultClients) GetEC2Client(region string) ec2iface.EC2API {
	return ec2.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetECSClient returns a new ECS client for the given region
func (c DefaultClients) GetECSClient(region string) ecsiface.ECSAPI {
	return ecs.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetElastiCacheClient returns a new ElastiCache client for the given region
func (c DefaultClients) GetElastiCacheClient(region string) elasticacheiface.ElastiCacheAPI {
	return elasticache.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetElasticsearchServiceClient returns a new ElasticsearchService client for the given region
func (c DefaultClients) GetElasticsearchServiceClient(region string) elasticsearchserviceiface.ElasticsearchServiceAPI {
	return elasticsearchservice.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetELBClient returns a new ELB client for the given region
func (c DefaultClients) GetELBClient(region string) elbiface.ELBAPI {
	return elb.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetELBV2Client returns a new ELBV2 client for the given region
func (c DefaultClients) GetELBV2Client(region string) elbv2iface.ELBV2API {
	return elbv2.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetIAMClient returns a new IAM client for the given region
func (c DefaultClients) GetIAMClient(region string) iamiface.IAMAPI {
	return iam.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetRDSClient returns a new RDS client for the given region
func (c DefaultClients) GetRDSClient(region string) rdsiface.RDSAPI {
	return rds.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetRoute53Client returns a new Route53 client for the given region
func (c DefaultClients) GetRoute53Client(region string) route53iface.Route53API {
	return route53.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}

// GetS3Client returns a new S3 client for the given region
func (c DefaultClients) GetS3Client(region string) s3iface.S3API {
	return s3.New(session.Must(session.NewSession()), &aws.Config{Region: aws.String(region)})
}
