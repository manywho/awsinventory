package awsdata

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/manywho/awsinventory/internal/inventory"
	"github.com/sirupsen/logrus"
)

const (
	// AssetTypeECSContainer is the value used in the AssetType field when fetching ECS containers
	AssetTypeECSContainer string = "ECS Container"

	// ServiceECS is the key for the ECS service
	ServiceECS string = "ecs"
)

func (d *AWSData) loadECSContainers(region string) {
	defer d.wg.Done()

	ec2Svc := d.clients.GetEC2Client(region)
	ecsSvc := d.clients.GetECSClient(region)

	log := d.log.WithFields(logrus.Fields{
		"region":  region,
		"service": ServiceECS,
	})

	log.Info("loading data")

	var clusterArns []*string
	done := false
	params := &ecs.ListClustersInput{}
	for !done {
		out, err := ecsSvc.ListClusters(params)
		if err != nil {
			d.results <- result{Err: err}
			return
		}

		clusterArns = append(clusterArns, out.ClusterArns...)

		if out.NextToken == nil {
			done = true
		} else {
			params.NextToken = out.NextToken
		}
	}

	// TODO: API call can only handle 100 cluster ARNs at a time
	out, err := ecsSvc.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: clusterArns,
	})
	if err != nil {
		d.results <- result{Err: err}
		return
	}

	log.Info("processing data")

	for _, cluster := range out.Clusters {
		var taskArns []*string
		done := false
		params := &ecs.ListTasksInput{
			Cluster: cluster.ClusterArn,
		}

		for !done {
			outListTasks, err := ecsSvc.ListTasks(params)
			if err != nil {
				d.results <- result{Err: err}
				return
			}

			taskArns = append(taskArns, outListTasks.TaskArns...)

			if outListTasks.NextToken == nil {
				done = true
			} else {
				params.NextToken = outListTasks.NextToken
			}
		}

		// TODO: API call can only handle 100 task ARNs at a time
		outDescribeTasks, err := ecsSvc.DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: cluster.ClusterArn,
			Tasks:   taskArns,
		})
		if err != nil {
			d.results <- result{Err: err}
			return
		}
		for _, task := range outDescribeTasks.Tasks {
			for _, container := range task.Containers {
				d.wg.Add(1)
				go d.processECSContainer(container, task, cluster, ec2Svc, region)
			}
		}
	}

	log.Info("finished processing data")
}

func (d *AWSData) processECSContainer(container *ecs.Container, task *ecs.Task, cluster *ecs.Cluster, ec2Svc ec2iface.EC2API, region string) {
	defer d.wg.Done()

	var ips []string
	var macAddresses []string
	var networkInterfaces []string

	for _, attachment := range task.Attachments {
		if aws.StringValue(attachment.Type) != "ElasticNetworkInterface" {
			continue
		}
		for _, details := range attachment.Details {
			switch aws.StringValue(details.Name) {
			case "privateIPv4Address":
				ips = append(ips, aws.StringValue(details.Value))
			case "ipv6Address":
				ips = append(ips, aws.StringValue(details.Value))
			case "macAddress":
				macAddresses = append(macAddresses, aws.StringValue(details.Value))
			case "networkInterfaceId":
				networkInterfaces = append(networkInterfaces, aws.StringValue(details.Value))
			}
		}
	}

	for _, networkInterface := range container.NetworkInterfaces {
		ips = AppendIfMissing(ips, aws.StringValue(networkInterface.PrivateIpv4Address))
		ips = AppendIfMissing(ips, aws.StringValue(networkInterface.Ipv6Address))
	}

	var hardware = aws.StringValue(task.LaunchType)
	if hardware == "FARGATE" {
		hardware = fmt.Sprintf("%s %s", hardware, aws.StringValue(task.PlatformVersion))
	}

	var vpcId string
	out, err := ec2Svc.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: aws.StringSlice(networkInterfaces),
	})
	if err != nil {
		d.results <- result{Err: err}
	} else if len(out.NetworkInterfaces) > 0 {
		vpcId = aws.StringValue(out.NetworkInterfaces[0].VpcId)
	}

	d.results <- result{
		Row: inventory.Row{
			UniqueAssetIdentifier:     fmt.Sprintf("%s-%s", aws.StringValue(container.Name), aws.StringValue(container.RuntimeId)),
			IPv4orIPv6Address:         strings.Join(ips, "\n"),
			Virtual:                   true,
			MACAddress:                strings.Join(macAddresses, "\n"),
			BaselineConfigurationName: aws.StringValue(container.Image),
			Location:                  region,
			AssetType:                 AssetTypeECSContainer,
			HardwareMakeModel:         hardware,
			Function:                  fmt.Sprintf("%s %s", aws.StringValue(cluster.ClusterName), aws.StringValue(task.Group)),
			SerialAssetTagNumber:      aws.StringValue(container.ContainerArn),
			VLANNetworkID:             vpcId,
		},
	}
}
