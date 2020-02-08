package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//  Defined global variable, so we can easily request EC2 related data from functions.
var ec2svc *ec2.EC2

//  Print out usage if invalid parms are passed to the command
func usage() {
	fmt.Println("Description:  View information about the Elastic Network Interfaces (ENI's) in your account.")
	fmt.Println("")
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println("EXAMPLE:  ./ListENIs -R us-east-2 -I 1.2.3.4")
	fmt.Println("   NOTE:  Do not specify the -E and -I options together.")
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
	IpAddrPtr := flag.String("I", "", "(optional) - The private IP Address associated with the specific ENI that you want to view")
	ENIIdPtr := flag.String("E", "", "(optional) - The ENI ID associated with the specific ENI that you want to view")

	flag.Parse()

	// Check that required parms are set
	if *RegionPtr == "" {
		usage()
	} else if *IpAddrPtr != "" && *ENIIdPtr != "" {
		usage()
	}

	//  Setup a session to interract with AWS
	ec2svc = ec2.New(DefineSession(*RegionPtr))

	//  Get ENI information
	ENIList, err := ec2svc.DescribeNetworkInterfaces(BuildENIParms(*IpAddrPtr, *ENIIdPtr))
	CheckError(err, "there was an error listing instances", 1)

	for _, eni := range ENIList.NetworkInterfaces {
		fmt.Printf("%s: %s [%s]\n", *eni.NetworkInterfaceId, *eni.PrivateDnsName, *eni.PrivateIpAddress)
		fmt.Printf("  Status: %s\n", *eni.Status)
		if eni.Association != nil {
			fmt.Printf("  Public DNS/IP: %s [%s]\n", *eni.Association.PublicDnsName, *eni.Association.PublicIp)
		}
		fmt.Printf("  Description: %s\n", *eni.Description)
		fmt.Printf("  VPC/Network: %s [%s]\n", *eni.VpcId, *eni.SubnetId)
		fmt.Printf("  Interface Type: %s\n", *eni.InterfaceType)
		fmt.Println("")
	}
}
