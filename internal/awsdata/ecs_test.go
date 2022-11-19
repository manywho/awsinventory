package awsdata_test

import (
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	"github.com/sudoinclabs/awsinventory/internal/inventory"
)

var testECSContainerRows = []inventory.Row{
	{
		UniqueAssetIdentifier:     "test-container-1-b026324c6904b2a9cb4b88d6d61c81d1-1234567890",
		IPv4orIPv6Address:         "10.1.2.3\n172.16.4.5",
		Virtual:                   true,
		MACAddress:                "ab:cd:ef:00:11:22",
		BaselineConfigurationName: "987654321012.dkr.ecr.us-east-2.amazonaws.com/app-1:latest",
		Location:                  DefaultRegion,
		AssetType:                 "ECS Container",
		HardwareMakeModel:         "FARGATE 1.4.0",
		Function:                  "ecs-cluster-1 service:ecs-service-1",
		SerialAssetTagNumber:      "arn:aws:ecs:us-east-2:123456789101:container/c73cb0a0-1ee7-4b38-af84-27054f83322e",
		VLANNetworkID:             "vpc-123456789",
	},
	{
		UniqueAssetIdentifier:     "test-container-2-b026324c6904b2a9cb4b88d6d61c81d1-2468101214",
		IPv4orIPv6Address:         "10.1.2.3",
		Virtual:                   true,
		MACAddress:                "ab:cd:ef:00:11:22",
		BaselineConfigurationName: "987654321012.dkr.ecr.us-east-2.amazonaws.com/app-2:latest",
		Location:                  DefaultRegion,
		AssetType:                 "ECS Container",
		HardwareMakeModel:         "FARGATE 1.4.0",
		Function:                  "ecs-cluster-1 service:ecs-service-1",
		SerialAssetTagNumber:      "arn:aws:ecs:us-east-2:123456789101:container/80966f57-f6ff-4c95-a38b-e9da670c8bdf",
		VLANNetworkID:             "vpc-123456789",
	},
	{
		UniqueAssetIdentifier:     "test-container-3-26ab0db90d72e28ad0ba1e22ee510510-1234567890",
		IPv4orIPv6Address:         "10.9.8.7\n10.100.250.25",
		Virtual:                   true,
		MACAddress:                "ab:00:cd:11:ef:22",
		BaselineConfigurationName: "987654321012.dkr.ecr.us-east-2.amazonaws.com/app-3:latest",
		Location:                  DefaultRegion,
		AssetType:                 "ECS Container",
		HardwareMakeModel:         "EC2",
		Function:                  "ecs-cluster-1 service:ecs-service-2",
		SerialAssetTagNumber:      "arn:aws:ecs:us-east-2:123456789101:container/71eaafe1-94ae-4e91-92fb-3c97ec4d63c5",
		VLANNetworkID:             "vpc-123456789",
	},
	{
		UniqueAssetIdentifier:     "test-container-4-6d7fce9fee471194aa8b5b6e47267f03-1234567890",
		IPv4orIPv6Address:         "192.168.0.1\n10.200.10.1\n::ffff:0ac8:0a01",
		Virtual:                   true,
		MACAddress:                "fe:99:dc:88:ba:77",
		BaselineConfigurationName: "987654321012.dkr.ecr.us-east-2.amazonaws.com/app-4:latest",
		Location:                  DefaultRegion,
		AssetType:                 "ECS Container",
		HardwareMakeModel:         "FARGATE LATEST",
		Function:                  "ecs-cluster-1 service:ecs-service-3",
		SerialAssetTagNumber:      "arn:aws:ecs:us-east-2:123456789101:container/3c77e658-9049-46c1-9352-53b59d97f0ac",
		VLANNetworkID:             "vpc-123456789",
	},
}

// Test Data
var testECSListClustersOutputPage1 = &ecs.ListClustersOutput{
	NextToken: aws.String("arn:aws:ecs:us-east-2:123456789101:cluster/ecs-cluster-1"),
	ClusterArns: []*string{
		aws.String("arn:aws:ecs:us-east-2:123456789101:cluster/ecs-cluster-1"),
	},
}

var testECSListClustersOutputPage2 = &ecs.ListClustersOutput{
	ClusterArns: []*string{
		aws.String("arn:aws:ecs:us-east-2:123456789101:cluster/ecs-cluster-2"),
	},
}

var testECSDescribeClustersOutput = &ecs.DescribeClustersOutput{
	Clusters: []*ecs.Cluster{
		{
			ClusterArn:  aws.String("arn:aws:ecs:us-east-2:123456789101:cluster/ecs-cluster-1"),
			ClusterName: aws.String("ecs-cluster-1"),
		},
	},
}

var testECSListTasksOutputPage1 = &ecs.ListTasksOutput{
	NextToken: aws.String("arn:aws:ecs:us-east-2:123456789101:task/ecs-cluster-1/26ab0db90d72e28ad0ba1e22ee510510"),
	TaskArns: []*string{
		aws.String("arn:aws:ecs:us-east-2:123456789101:task/ecs-cluster-1/b026324c6904b2a9cb4b88d6d61c81d1"),
		aws.String("arn:aws:ecs:us-east-2:123456789101:task/ecs-cluster-1/26ab0db90d72e28ad0ba1e22ee510510"),
	},
}

var testECSListTasksOutputPage2 = &ecs.ListTasksOutput{
	TaskArns: []*string{
		aws.String("arn:aws:ecs:us-east-2:123456789101:task/ecs-cluster-1/6d7fce9fee471194aa8b5b6e47267f03"),
	},
}

var testECSDescribeTasksOutput = &ecs.DescribeTasksOutput{
	Tasks: []*ecs.Task{
		{
			Attachments: []*ecs.Attachment{
				{
					Type: aws.String("ElasticNetworkInterface"),
					Details: []*ecs.KeyValuePair{
						{
							Name:  aws.String("networkInterfaceId"),
							Value: aws.String("eni-12345678"),
						},
						{
							Name:  aws.String("macAddress"),
							Value: aws.String("ab:cd:ef:00:11:22"),
						},
						{
							Name:  aws.String("privateIPv4Address"),
							Value: aws.String("10.1.2.3"),
						},
					},
				},
			},
			Containers: []*ecs.Container{
				{
					ContainerArn: aws.String(testECSContainerRows[0].SerialAssetTagNumber),
					Name:         aws.String("test-container-1"),
					Image:        aws.String(testECSContainerRows[0].BaselineConfigurationName),
					NetworkInterfaces: []*ecs.NetworkInterface{
						{
							PrivateIpv4Address: aws.String("172.16.4.5"),
						},
					},
					RuntimeId: aws.String("b026324c6904b2a9cb4b88d6d61c81d1-1234567890"),
				},
				{
					ContainerArn: aws.String(testECSContainerRows[1].SerialAssetTagNumber),
					Name:         aws.String("test-container-2"),
					Image:        aws.String(testECSContainerRows[1].BaselineConfigurationName),
					RuntimeId:    aws.String("b026324c6904b2a9cb4b88d6d61c81d1-2468101214"),
				},
			},
			Group:           aws.String("service:ecs-service-1"),
			LaunchType:      aws.String("FARGATE"),
			PlatformVersion: aws.String("1.4.0"),
			TaskArn:         aws.String("arn:aws:ecs:us-east-2:123456789101:task/ecs-cluster-1/b026324c6904b2a9cb4b88d6d61c81d1"),
		},
		{
			Attachments: []*ecs.Attachment{
				{
					Type: aws.String("ElasticNetworkInterface"),
					Details: []*ecs.KeyValuePair{
						{
							Name:  aws.String("networkInterfaceId"),
							Value: aws.String("eni-abcdefgh"),
						},
						{
							Name:  aws.String("macAddress"),
							Value: aws.String("ab:00:cd:11:ef:22"),
						},
						{
							Name:  aws.String("privateIPv4Address"),
							Value: aws.String("10.9.8.7"),
						},
					},
				},
			},
			Containers: []*ecs.Container{
				{
					ContainerArn: aws.String(testECSContainerRows[2].SerialAssetTagNumber),
					Name:         aws.String("test-container-3"),
					Image:        aws.String(testECSContainerRows[2].BaselineConfigurationName),
					NetworkInterfaces: []*ecs.NetworkInterface{
						{
							PrivateIpv4Address: aws.String("10.100.250.25"),
						},
					},
					RuntimeId: aws.String("26ab0db90d72e28ad0ba1e22ee510510-1234567890"),
				},
			},
			Group:      aws.String("service:ecs-service-2"),
			LaunchType: aws.String("EC2"),
			TaskArn:    aws.String("arn:aws:ecs:us-east-2:123456789101:task/ecs-cluster-1/26ab0db90d72e28ad0ba1e22ee510510"),
		},
		{
			Attachments: []*ecs.Attachment{
				{
					Type: aws.String("ElasticNetworkInterface"),
					Details: []*ecs.KeyValuePair{
						{
							Name:  aws.String("networkInterfaceId"),
							Value: aws.String("eni-0a1b2c3d"),
						},
						{
							Name:  aws.String("macAddress"),
							Value: aws.String("fe:99:dc:88:ba:77"),
						},
						{
							Name:  aws.String("privateIPv4Address"),
							Value: aws.String("192.168.0.1"),
						},
					},
				},
			},
			Containers: []*ecs.Container{
				{
					ContainerArn: aws.String(testECSContainerRows[3].SerialAssetTagNumber),
					Name:         aws.String("test-container-4"),
					Image:        aws.String(testECSContainerRows[3].BaselineConfigurationName),
					NetworkInterfaces: []*ecs.NetworkInterface{
						{
							Ipv6Address:        aws.String("::ffff:0ac8:0a01"),
							PrivateIpv4Address: aws.String("10.200.10.1"),
						},
					},
					RuntimeId: aws.String("6d7fce9fee471194aa8b5b6e47267f03-1234567890"),
				},
			},
			Group:           aws.String("service:ecs-service-3"),
			LaunchType:      aws.String("FARGATE"),
			PlatformVersion: aws.String("LATEST"),
			TaskArn:         aws.String("arn:aws:ecs:us-east-2:123456789101:task/ecs-cluster-1/6d7fce9fee471194aa8b5b6e47267f03"),
		},
	},
}

var testEC2DescribeNetworkInterfacesOutput = &ec2.DescribeNetworkInterfacesOutput{
	NetworkInterfaces: []*ec2.NetworkInterface{
		{
			VpcId: aws.String("vpc-123456789"),
		},
	},
}

// Mocks
type ECSMock struct {
	ecsiface.ECSAPI
}

func (e ECSMock) ListClusters(cfg *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
	if cfg.NextToken == nil {
		return testECSListClustersOutputPage1, nil
	}

	return testECSListClustersOutputPage2, nil
}

func (e ECSMock) DescribeClusters(cfg *ecs.DescribeClustersInput) (*ecs.DescribeClustersOutput, error) {
	return testECSDescribeClustersOutput, nil
}

func (e ECSMock) ListTasks(cfg *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
	if cfg.NextToken == nil {
		return testECSListTasksOutputPage1, nil
	}

	return testECSListTasksOutputPage2, nil
}

func (e ECSMock) DescribeTasks(cfg *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	if cfg.Cluster == testECSListClustersOutputPage2.ClusterArns[0] {
		return &ecs.DescribeTasksOutput{}, nil
	}

	return testECSDescribeTasksOutput, nil
}

func (e EC2Mock) DescribeNetworkInterfaces(cfg *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error) {
	return testEC2DescribeNetworkInterfacesOutput, nil
}

type ECSErrorMock struct {
	ecsiface.ECSAPI
}

func (e ECSErrorMock) ListClusters(cfg *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
	return &ecs.ListClustersOutput{}, testError
}

func (e ECSErrorMock) DescribeClusters(cfg *ecs.DescribeClustersInput) (*ecs.DescribeClustersOutput, error) {
	return &ecs.DescribeClustersOutput{}, testError
}

func (e ECSErrorMock) ListTasks(cfg *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
	return &ecs.ListTasksOutput{}, testError
}

func (e ECSErrorMock) DescribeTasks(cfg *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	return &ecs.DescribeTasksOutput{}, testError
}

func (e EC2ErrorMock) DescribeNetworkInterfaces(cfg *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error) {
	return &ec2.DescribeNetworkInterfacesOutput{}, testError
}

// Tests
func TestCanLoadECSContainers(t *testing.T) {
	d := New(logrus.New(), TestClients{EC2: EC2Mock{}, ECS: ECSMock{}})

	var rows []inventory.Row
	d.Load([]string{DefaultRegion}, []string{ServiceECS}, func(row inventory.Row) error {
		rows = append(rows, row)
		return nil
	})

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].UniqueAssetIdentifier < rows[j].UniqueAssetIdentifier
	})

	require.Equal(t, 4, len(rows))

	for i := range rows {
		require.Equal(t, testECSContainerRows[i], rows[i])
	}
}

func TestLoadECSContainersLogsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{EC2: EC2ErrorMock{}, ECS: ECSErrorMock{}})

	d.Load([]string{DefaultRegion}, []string{ServiceECS}, nil)

	assertTestErrorWasLogged(t, hook.Entries)
}
