package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//  Print out usage if invalid parms are passed to the command
func usage() {
	fmt.Println("Description:  View information about the EC2 instances in your account.")
	flag.PrintDefaults()
	os.Exit(255)
}

//  Sets the region and grabs local credentials so you can access the AWS environment
func DefineSession(region string) *session.Session {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		fmt.Println("there was an error authenticating with AWS", err.Error())
		fmt.Errorf(err.Error())
		os.Exit(1)
	}
	return sess
}

//  Build parameters used to search for EC2 instances
func BuildEC2Parms(instanceid string) *ec2.DescribeInstancesInput {
	var my_params *ec2.DescribeInstancesInput

	if instanceid != "" {
		my_params = &ec2.DescribeInstancesInput{InstanceIds: []*string{aws.String(instanceid)}}
	} else {
		my_params = &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("instance-state-name"),
					Values: []*string{aws.String("running"), aws.String("pending"), aws.String("shutting-down"), aws.String("terminated"), aws.String("stopping"), aws.String("stopped")},
				},
			},
		}
	}
	return my_params
}

func main() {
	//  Get command line parms
	RegionPtr := flag.String("R", "", "(required) - AWS Region that you want to view EC2 Instances in")
	InstanceIdPtr := flag.String("I", "", "(optional) - The specific EC2 Instance ID that you want to view information about")
	flag.Parse()

	// Check that required parms are set
	if *RegionPtr == "" {
		usage()
	}

	sess := DefineSession(*RegionPtr)

	//  Get EC2 list
	ec2svc := ec2.New(sess)

	params := BuildEC2Parms(*InstanceIdPtr)
	InstanceList, err := ec2svc.DescribeInstances(params)

	if err != nil {
		fmt.Println("there was an error listing instances in", err.Error())
		fmt.Errorf(err.Error())
		os.Exit(1)
	}

	for idx, _ := range InstanceList.Reservations {
		for _, instance := range InstanceList.Reservations[idx].Instances {
			fmt.Printf("%s:\n", *instance.InstanceId)
			fmt.Printf("  Name: <NAME HERE FROM TAG>\n")
			fmt.Printf("  State: %s\n", *instance.State.Name)
			fmt.Printf("  Instance Type: %s\n", *instance.InstanceType)
			fmt.Printf("  AMI: %s [Creation Date: <GET CREATION DATE OF AMI>]\n", *instance.ImageId)
			fmt.Printf("  OS: <Get Description of AMI>\n")
			fmt.Printf("  Hostname/IP: %s [%s]\n", *instance.PrivateDnsName, *instance.PrivateIpAddress)
			fmt.Printf("  VPC/Network: %s [%s]\n", *instance.VpcId, *instance.SubnetId)
			fmt.Printf("  Security Group Information:\n")
			fmt.Printf("    Security Group Ingress ( allowed inbound connections ):\n")
			fmt.Printf("      Port: 1194 | Protocol: udp | IPs: 66.196.199.98/32, 104.152.133.151/32\n")
			fmt.Printf("    Security Group Egress ( allowed outbound connections ):\n")
			fmt.Printf("      Port: null | Protocol: -1 | IPs: 0.0.0.0/0\n")
		}
		fmt.Println("")
	}

}
