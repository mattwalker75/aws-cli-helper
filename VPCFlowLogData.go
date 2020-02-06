package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//  Defined global variable, so we can easily request EC2 related data from functions.
var ec2svc *ec2.EC2
var cwlsvc *cloudwatchlogs.CloudWatchLogs

//  Print out usage if invalid parms are passed to the command
func usage() {
	fmt.Println("Description:  View VPC Flow Log data in an easy to read format.")
	fmt.Println("")
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println("OUTPUT:  <ENI> : <source IP>[<source port>]  -->  <destination IP>[<destination port>] : <protocol> : <status>")
	fmt.Println("")
	fmt.Println("EXAMPLE:  ./VPCFlowLogData -R us-east-1 -V vpc-0c5b4e0c96815b5b7")
	fmt.Println("")
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

//  Build parameters used to search for ENI's
func BuildENIParms(ipaddress string, eniid string) *ec2.DescribeNetworkInterfacesInput {
	var my_params *ec2.DescribeNetworkInterfacesInput

	if ipaddress != "" {
		my_params = &ec2.DescribeNetworkInterfacesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("private-ip-address"),
					Values: []*string{aws.String(ipaddress)},
				},
			},
		}
	} else if eniid != "" {
		my_params = &ec2.DescribeNetworkInterfacesInput{NetworkInterfaceIds: []*string{aws.String(eniid)}}
	} else {
		my_params = &ec2.DescribeNetworkInterfacesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("status"),
					Values: []*string{aws.String("available"), aws.String("in-use")},
				},
			},
		}
	}
	return my_params
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {
	//  Get command line parms
	RegionPtr := flag.String("R", "", "(required) - AWS Region that you want to view ENI's in")
	VPCidPtr := flag.String("V", "", "(optional) - The VPC ID that is associated with the VPC Flow Log that you want to view")
	//ENIIdPtr := flag.String("E", "", "(optional) - The ENI ID associated with the ENI that you specifically want to view the VPC Flow Log of")

	flag.Parse()

	// Check that required parms are set
	if *RegionPtr == "" {
		usage()
	}

	//  Setup a session to interract with AWS
	ec2svc = ec2.New(DefineSession(*RegionPtr))
	cwlsvc = cloudwatchlogs.New(DefineSession(*RegionPtr))

	if *VPCidPtr == "" {
		my_vpc_parms := &ec2.DescribeVpcsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("state"),
					Values: []*string{aws.String("available")},
				},
			},
		}
		VPC_List, err := ec2svc.DescribeVpcs(my_vpc_parms)
		if err != nil {
			fmt.Println("there was an error getting a list of VPC's: ", err.Error())
			fmt.Errorf(err.Error())
			os.Exit(1)
		}

		fmt.Println("List of VPC's to choose from:")
		for _, vpcid := range VPC_List.Vpcs {
			fmt.Printf("   VPC ID: %s  - [ CIDR: %s ]\n", *vpcid.VpcId, *vpcid.CidrBlock)
		}
		fmt.Println("")
		fmt.Println("Re-run and specify a specific VPC ID using the -V parameter")
		os.Exit(0)
	}

	my_flowlogs_parms := &ec2.DescribeFlowLogsInput{
		Filter: []*ec2.Filter{
			{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(*VPCidPtr)},
			},
		},
	}

	//  Get the VPC Flow Log associated with the VPC
	FlowLog_List, err := ec2svc.DescribeFlowLogs(my_flowlogs_parms)
	if err != nil {
		fmt.Println("there was an error getting Flow Log information: ", err.Error())
		fmt.Errorf(err.Error())
		os.Exit(1)
	}

	if FlowLog_List.FlowLogs == nil {
		fmt.Println("No VPC Flow Logs defined for the VPC")
		os.Exit(0)
	}

	fmt.Println(*FlowLog_List.FlowLogs[0].LogGroupName)

	//  Get the Logstreams ( eni-XXXXXXX ) associated with the VPC Flow Log ( Log Group )
	LogStreamList, err := cwlsvc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(*FlowLog_List.FlowLogs[0].LogGroupName),
	})
	if err != nil {
		fmt.Println("Got error getting log streams:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for _, LogStream := range LogStreamList.LogStreams {
		fmt.Println(*LogStream.LogStreamName)
		LScontent, err := cwlsvc.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  aws.String(*FlowLog_List.FlowLogs[0].LogGroupName),
			LogStreamName: aws.String(*LogStream.LogStreamName),
		})
		if err != nil {
			fmt.Println("Got error getting log stream content:")
			fmt.Println(err.Error())
		}
		for _, LSline := range LScontent.Events {
			fmt.Printf("Time (%d): %s\n", *LSline.Timestamp, *LSline.Message)
		}
		fmt.Println(*LScontent.NextForwardToken)
	}
}
