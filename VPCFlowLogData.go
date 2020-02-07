package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

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
	ENIIdPtr := flag.String("E", "", "(optional) - The ENI ID associated with the ENI that you specifically want to view the VPC Flow Log of")

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

	//  Get the Logstreams ( eni-XXXXXXX ) associated with the VPC Flow Log ( Log Group )
	var LogStreamList *cloudwatchlogs.DescribeLogStreamsOutput
	if *ENIIdPtr == "" {
		LogStreamList, err = cwlsvc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName: aws.String(*FlowLog_List.FlowLogs[0].LogGroupName),
		})
	} else {
		LogStreamList, err = cwlsvc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName:        aws.String(*FlowLog_List.FlowLogs[0].LogGroupName),
			LogStreamNamePrefix: aws.String(*ENIIdPtr),
		})
	}

	if err != nil {
		fmt.Println("Got error getting log streams:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//  Process the CloudWatch Log Steams ( log files ) one at a time
	for _, LogStream := range LogStreamList.LogStreams {
		LScontent, err := cwlsvc.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  aws.String(*FlowLog_List.FlowLogs[0].LogGroupName),
			LogStreamName: aws.String(*LogStream.LogStreamName),
		})
		if err != nil {
			fmt.Println("Got error getting log stream content:")
			fmt.Println(err.Error())
		}

		//  Process the content of the log file one line at a time
		OldForwardToken := "NOTVALID"
		for {
			//  Check if current token matches old token, if so then there is no additional data so we can break out of loop
			if *LScontent.NextForwardToken == OldForwardToken {
				break
			}

			//  Process each line of the log
			for _, LSline := range LScontent.Events {
				//  Log Steam name:  <ENI ID>-all
				RawLSline := strings.Fields(*LSline.Message)
				if RawLSline[13] != "NODATA" {
					t := time.Unix((*LSline.Timestamp / 1000), 0)
					timezone, _ := t.Zone()
					eventtime := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d %s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), timezone)
					var my_protocol string
					if RawLSline[7] == "6" {
						my_protocol = "tcp"
					} else if RawLSline[7] == "17" {
						my_protocol = "udp"
					} else {
						my_protocol = RawLSline[7]
					}
					formatline := RawLSline[2] + " : " + RawLSline[3] + "[" + RawLSline[5] + "] --> " + RawLSline[4] + "[" + RawLSline[6] + "] : " + my_protocol + " : " + RawLSline[12] + " " + RawLSline[13]
					if RawLSline[12] == "ACCEPT" && RawLSline[13] == "OK" {
						fmt.Printf(" %s : %s\n", eventtime, formatline)
					} else {
						fmt.Printf(" %s : %s  <-\n", eventtime, formatline)
					}
				}
			}

			//  Check if there is a token for additional data, of so then make the request with the token to get the additional data
			if LScontent.NextForwardToken != nil {
				OldForwardToken = *LScontent.NextForwardToken
				LScontent, err = cwlsvc.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
					LogGroupName:  aws.String(*FlowLog_List.FlowLogs[0].LogGroupName),
					LogStreamName: aws.String(*LogStream.LogStreamName),
					NextToken:     aws.String(*LScontent.NextForwardToken),
				})
				if err != nil {
					fmt.Println("Got error getting log stream content:")
					fmt.Println(err.Error())
				}
			}
		}
	}
}
