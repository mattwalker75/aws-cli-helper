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

var err error

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

//  Check error codes
func CheckError(returncode error, message string, exit_status int) {
	if returncode != nil {
		fmt.Printf("%s: %s", message, returncode.Error())
		fmt.Errorf(returncode.Error())
		os.Exit(exit_status)
	}
}

//  Sets the region and grabs local credentials so you can access the AWS environment
func DefineSession(region string) *session.Session {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	CheckError(err, "there was an error authenticating with AWS", 1)

	return sess
}

//  Checks to see if there is a VPC ID, if not then list the VPC's and exit
func CheckVPCParm(vpcid_parm string) {
	if vpcid_parm == "" {
		my_vpc_parms := &ec2.DescribeVpcsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("state"),
					Values: []*string{aws.String("available")},
				},
			},
		}
		VPC_List, err := ec2svc.DescribeVpcs(my_vpc_parms)
		CheckError(err, "there was an error getting a list of VPC's", 1)

		fmt.Println("List of VPC's to choose from:")
		for _, vpcid := range VPC_List.Vpcs {
			fmt.Printf("   VPC ID: %s  - [ CIDR: %s ]\n", *vpcid.VpcId, *vpcid.CidrBlock)
		}
		fmt.Println("")
		fmt.Println("Re-run and specify a specific VPC ID using the -V parameter")
		os.Exit(0)
	}
}

//  Get the VPC Flow Log associated with the VPC
func GetVPCFlowLog(vpcid string) *ec2.DescribeFlowLogsOutput {
	my_flowlogs_parms := &ec2.DescribeFlowLogsInput{
		Filter: []*ec2.Filter{
			{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(vpcid)},
			},
		},
	}

	VPCFlowLogs, err := ec2svc.DescribeFlowLogs(my_flowlogs_parms)
	CheckError(err, "there was an error getting Flow Log information", 1)

	//  If there is no VPC Flow Log found, then exit
	if VPCFlowLogs.FlowLogs == nil {
		fmt.Println("No VPC Flow Logs defined for the VPC")
		os.Exit(0)
	}

	return VPCFlowLogs
}

//  List the CloudWatch Log Streams ( log file names) associated with the CloudWatch Log Group
//    - If a specific ENI ID is provied, then only return the Log Stream associated with that ENI
func GetLogStreamList(LGName string, ENI string) *cloudwatchlogs.DescribeLogStreamsOutput {
	var LSList *cloudwatchlogs.DescribeLogStreamsOutput
	if ENI == "" {
		LSList, err = cwlsvc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName: aws.String(LGName),
		})
	} else {
		LSList, err = cwlsvc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName:        aws.String(LGName),
			LogStreamNamePrefix: aws.String(ENI),
		})
	}

	CheckError(err, "Got error getting log streams", 1)

	return LSList
}

//  Convert EPOC time to a formatted date/time output
func EPOCtoDateTime(epoc int64) string {
	t := time.Unix((epoc / 1000), 0)
	timezone, _ := t.Zone()
	FormattedDateTime := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d %s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), timezone)

	return FormattedDateTime
}

//  Re-format VPC Flow Log data and print it out
func ReformatLogEntry(linedata *cloudwatchlogs.OutputLogEvent) {
	//  Split string into array using space delimitation so the text can be reformatted
	var my_protocol string
	RawLSline := strings.Fields(*linedata.Message)
	if RawLSline[13] != "NODATA" {
		eventtime := EPOCtoDateTime(*linedata.Timestamp)
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

///////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {
	//  Get command line parms
	RegionPtr := flag.String("R", "", "(required) - AWS Region that you want to view ENI's in")
	VPCidPtr := flag.String("V", "", "(optional) - The VPC ID that is associated with the VPC Flow Log that you want to view")
	ENIIdPtr := flag.String("E", "", "(optional) - The ENI ID associated with the ENI that you specifically want to view the VPC Flow Log of")

	var OldForwardToken string

	flag.Parse()

	// Check that required parms are set
	if *RegionPtr == "" {
		usage()
	}

	//  Setup a session to interract with AWS
	ec2svc = ec2.New(DefineSession(*RegionPtr))
	cwlsvc = cloudwatchlogs.New(DefineSession(*RegionPtr))

	//  Check if VPC ID was provided as a parm or not
	CheckVPCParm(*VPCidPtr)

	//  Get the Flow Log assigned to the VPC
	FlowLog_List := GetVPCFlowLog(*VPCidPtr)

	//  Get the Logstreams ( eni-XXXXXXX ) associated with the VPC Flow Log ( Log Group )
	LogStreamList := GetLogStreamList(*FlowLog_List.FlowLogs[0].LogGroupName, *ENIIdPtr)

	//  Process each CloudWatch Log Stream ( log file )
	for _, LogStream := range LogStreamList.LogStreams {
		LScontent, err := cwlsvc.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  aws.String(*FlowLog_List.FlowLogs[0].LogGroupName),
			LogStreamName: aws.String(*LogStream.LogStreamName),
		})
		CheckError(err, "Got error getting log stream content", 1)

		//  Process the content of a Log Stream one "page" at a time
		OldForwardToken = "ALWAYS_PROCESS_THE_FIRST_PAGE"
		for {
			//  Check if current token matches old token, if so then there is no additional data so we can break out of loop
			if *LScontent.NextForwardToken == OldForwardToken {
				break
			}

			//  Process the content of one "page" from the Log Stream one line at a time
			for _, LSline := range LScontent.Events {
				ReformatLogEntry(LSline)
			}

			//  Check if there is a token for additional "pages" of data,
			//    if so then make the request with the token to get the next page of data
			if LScontent.NextForwardToken != nil {
				OldForwardToken = *LScontent.NextForwardToken
				LScontent, err = cwlsvc.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
					LogGroupName:  aws.String(*FlowLog_List.FlowLogs[0].LogGroupName),
					LogStreamName: aws.String(*LogStream.LogStreamName),
					NextToken:     aws.String(*LScontent.NextForwardToken),
				})
				CheckError(err, "Got error getting log stream content", 1)
			}
		}
	}
}
