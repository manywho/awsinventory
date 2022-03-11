package awsdata_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
	"github.com/aws/aws-sdk-go/service/codecommit/codecommitiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice/elasticsearchserviceiface"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/aws/aws-sdk-go/service/workspaces/workspacesiface"
)

var testError = errors.New("test aws error")

type TestClients struct {
	CloudFront           cloudfrontiface.CloudFrontAPI
	CodeCommit           codecommitiface.CodeCommitAPI
	DynamoDB             dynamodbiface.DynamoDBAPI
	EC2                  ec2iface.EC2API
	ECR                  ecriface.ECRAPI
	ECS                  ecsiface.ECSAPI
	ElastiCache          elasticacheiface.ElastiCacheAPI
	ElasticsearchService elasticsearchserviceiface.ElasticsearchServiceAPI
	ELB                  elbiface.ELBAPI
	ELBV2                elbv2iface.ELBV2API
	IAM                  iamiface.IAMAPI
	KMS                  kmsiface.KMSAPI
	Lambda               lambdaiface.LambdaAPI
	RDS                  rdsiface.RDSAPI
	Route53              route53iface.Route53API
	S3                   s3iface.S3API
	SQS                  sqsiface.SQSAPI
	WorkSpace                  workspacesiface.WorkSpacesAPI
}

func (c TestClients) GetCloudFrontClient(region string) cloudfrontiface.CloudFrontAPI {
	return c.CloudFront
}

func (c TestClients) GetCodeCommitClient(region string) codecommitiface.CodeCommitAPI {
	return c.CodeCommit
}

func (c TestClients) GetDynamoDBClient(region string) dynamodbiface.DynamoDBAPI {
	return c.DynamoDB
}

func (c TestClients) GetEC2Client(region string) ec2iface.EC2API {
	return c.EC2
}

func (c TestClients) GetECRClient(region string) ecriface.ECRAPI {
	return c.ECR
}

func (c TestClients) GetECSClient(region string) ecsiface.ECSAPI {
	return c.ECS
}

func (c TestClients) GetElastiCacheClient(region string) elasticacheiface.ElastiCacheAPI {
	return c.ElastiCache
}

func (c TestClients) GetElasticsearchServiceClient(region string) elasticsearchserviceiface.ElasticsearchServiceAPI {
	return c.ElasticsearchService
}

func (c TestClients) GetELBClient(region string) elbiface.ELBAPI {
	return c.ELB
}

func (c TestClients) GetELBV2Client(region string) elbv2iface.ELBV2API {
	return c.ELBV2
}

func (c TestClients) GetIAMClient(region string) iamiface.IAMAPI {
	return c.IAM
}

func (c TestClients) GetKMSClient(region string) kmsiface.KMSAPI {
	return c.KMS
}

func (c TestClients) GetLambdaClient(region string) lambdaiface.LambdaAPI {
	return c.Lambda
}

func (c TestClients) GetRDSClient(region string) rdsiface.RDSAPI {
	return c.RDS
}

func (c TestClients) GetRoute53Client(region string) route53iface.Route53API {
	return c.Route53
}

func (c TestClients) GetS3Client(region string) s3iface.S3API {
	return c.S3
}

func (c TestClients) GetSQSClient(region string) sqsiface.SQSAPI {
	return c.SQS
}

func (c TestClients) GetWorkSpaceClient(region string) workspacesiface.WorkSpacesAPI {
	return c.WorkSpace
}
