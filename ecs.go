package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/juju/loggo"

	"errors"
	"math"
	"os"
	"strings"
)

// logging
var ecsLogger = loggo.GetLogger("ecs")

// ECS struct
type ECS struct {
	clusterArns         []*string
	serviceArns         []*string
	taskArns            []string
	clusterNames        []string
	serviceNames        []string
	taskNames           []string
	containerInstances  map[string]string
	selectedClusterName string
	selectedServiceName string
	ipAddr              *string
	keyName             *string
	svc                 *ecs.ECS
}

func newECS() *ECS {
	e := ECS{}
	// set default region if no region is set
	if os.Getenv("AWS_REGION") == "" {
		os.Setenv("AWS_REGION", "us-east-1")
	}
	e.svc = ecs.New(session.New())
	e.listCluster()
	e.getClusterNames()
	return &e
}

// Creates ECS repository
func (e *ECS) listCluster() error {
	input := &ecs.ListClustersInput{}

	pageNum := 0
	err := e.svc.ListClustersPages(input,
		func(page *ecs.ListClustersOutput, lastPage bool) bool {
			pageNum++
			e.clusterArns = append(e.clusterArns, page.ClusterArns...)
			return pageNum <= 20
		})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			ecsLogger.Errorf(aerr.Error())
		} else {
			ecsLogger.Errorf(err.Error())
		}
		return err
	}
	return nil
}

func (e *ECS) getClusterNames() error {
	input := &ecs.DescribeClustersInput{
		Clusters: e.clusterArns,
	}

	result, err := e.svc.DescribeClusters(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			ecsLogger.Errorf(aerr.Error())
		} else {
			ecsLogger.Errorf(err.Error())
		}
		return err
	}
	for _, cluster := range result.Clusters {
		e.clusterNames = append(e.clusterNames, *cluster.ClusterName)
	}
	return nil
}
func (e *ECS) listServiceArns(clusterName string) error {
	e.serviceArns = []*string{}
	input := &ecs.ListServicesInput{
		Cluster: aws.String(clusterName),
	}

	pageNum := 0
	err := e.svc.ListServicesPages(input,
		func(page *ecs.ListServicesOutput, lastPage bool) bool {
			pageNum++
			e.serviceArns = append(e.serviceArns, page.ServiceArns...)
			return pageNum <= 20
		})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			ecsLogger.Errorf(aerr.Error())
		} else {
			ecsLogger.Errorf(err.Error())
		}
		return err
	}
	return nil
}

// gets service Arns and returns service names
func (e *ECS) getServices(clusterName string) ([]string, error) {
	e.selectedClusterName = clusterName
	err := e.listServiceArns(clusterName)
	if err != nil {
		return []string{}, err
	}

	// fetch per 10
	var y float64 = float64(len(e.serviceArns)) / 10

	for i := 0; i < int(math.Ceil(y)); i++ {

		f := i * 10
		t := int(math.Min(float64(10+10*i), float64(len(e.serviceArns))))

		input := &ecs.DescribeServicesInput{
			Cluster:  aws.String(clusterName),
			Services: e.serviceArns[f:t],
		}

		result, err := e.svc.DescribeServices(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				ecsLogger.Errorf(aerr.Error())
			} else {
				ecsLogger.Errorf(err.Error())
			}
			return []string{}, err
		}
		for _, service := range result.Services {
			e.serviceNames = append(e.serviceNames, *service.ServiceName)
		}
	}
	return e.serviceNames, nil
}

// lists task arns
func (e *ECS) listTaskArns(serviceName string) error {
	e.taskArns = []string{}
	input := &ecs.ListTasksInput{
		Cluster:     aws.String(e.selectedClusterName),
		ServiceName: aws.String(serviceName),
	}

	pageNum := 0
	err := e.svc.ListTasksPages(input,
		func(page *ecs.ListTasksOutput, lastPage bool) bool {
			pageNum++
			e.taskArns = append(e.taskArns, aws.StringValueSlice(page.TaskArns)...)
			return pageNum <= 20
		})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			ecsLogger.Errorf(aerr.Error())
		} else {
			ecsLogger.Errorf(err.Error())
		}
		return err
	}
	return nil
}

// list all task ARNs for a cluster
func (e *ECS) listAllTaskArns() error {
	e.taskArns = []string{}
	input := &ecs.ListTasksInput{
		Cluster: aws.String(e.selectedClusterName),
	}

	pageNum := 0
	err := e.svc.ListTasksPages(input,
		func(page *ecs.ListTasksOutput, lastPage bool) bool {
			pageNum++
			e.taskArns = append(e.taskArns, aws.StringValueSlice(page.TaskArns)...)
			return pageNum <= 20
		})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			ecsLogger.Errorf(aerr.Error())
		} else {
			ecsLogger.Errorf(err.Error())
		}
		return err
	}
	return nil
}
func (e *ECS) getTasks(serviceName string) ([]string, error) {
	e.selectedServiceName = serviceName
	e.containerInstances = make(map[string]string)
	err := e.listTaskArns(serviceName)
	if err != nil {
		return []string{}, err
	}
	return e.outputTaskNames()
}
func (e *ECS) getAllTasks() ([]string, error) {
	e.containerInstances = make(map[string]string)
	err := e.listAllTaskArns()
	if err != nil {
		return []string{}, err
	}
	return e.outputTaskNames()
}
func (e *ECS) outputTaskNames() ([]string, error) {
	// fetch per 100
	var y float64 = float64(len(e.taskArns)) / 100

	for i := 0; i < int(math.Ceil(y)); i++ {

		f := i * 100
		t := int(math.Min(float64(100+100*i), float64(len(e.taskArns))))

		input := &ecs.DescribeTasksInput{
			Cluster: aws.String(e.selectedClusterName),
			Tasks:   aws.StringSlice(e.taskArns[f:t]),
		}

		result, err := e.svc.DescribeTasks(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				ecsLogger.Errorf(aerr.Error())
			} else {
				ecsLogger.Errorf(err.Error())
			}
			return []string{}, err
		}
		for _, task := range result.Tasks {
			s := strings.Split(*task.TaskArn, "/")
			if len(s) > 1 {
				e.taskNames = append(e.taskNames, s[1])
				e.containerInstances[s[1]] = aws.StringValue(task.ContainerInstanceArn)
			}
		}
	}
	return e.taskNames, nil
}

func (e *ECS) getContainerInstanceIP(taskArn string) (*string, error) {

	if e.containerInstances[taskArn] == "" {
		return nil, errors.New("Task has no container instance assigned (task might not be running)")
	}

	input := &ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(e.selectedClusterName),
		ContainerInstances: []*string{aws.String(e.containerInstances[taskArn])},
	}

	result, err := e.svc.DescribeContainerInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			ecsLogger.Errorf(aerr.Error())
		} else {
			ecsLogger.Errorf(err.Error())
		}
		return nil, err
	}
	if len(result.ContainerInstances) == 0 {
		return nil, errors.New("Container instance not found")
	}

	inputInstances := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{result.ContainerInstances[0].Ec2InstanceId},
	}

	svcEc2 := ec2.New(session.New())
	resultInstances, err := svcEc2.DescribeInstances(inputInstances)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			ecsLogger.Errorf(aerr.Error())
		} else {
			ecsLogger.Errorf(err.Error())
		}
		return nil, err
	}
	if len(resultInstances.Reservations) == 0 {
		return nil, errors.New("EC2 instance not found")
	}
	if len(resultInstances.Reservations[0].Instances) == 0 {
		return nil, errors.New("EC2 instance not found")
	}
	e.ipAddr = resultInstances.Reservations[0].Instances[0].PrivateIpAddress
	e.keyName = resultInstances.Reservations[0].Instances[0].KeyName

	return e.ipAddr, nil
}
