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

//  Get the "Name" Tag.  If it is not set or set to null, then return "NOT SET"
func GetNameTag(tags []*ec2.Tag) string {
	NoName := "NOT SET"
	for _, tag := range tags {
		if *tag.Key == "Name" {
			if *tag.Value != "" {
				return *tag.Value
			}
		}
	}
	return NoName
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

//  Return the creation date and description of an AMI
func GetAMIInfo(ami string) (string, string) {
	AMIinput := &ec2.DescribeImagesInput{ImageIds: []*string{aws.String(ami)}}
	AMIData, err := ec2svc.DescribeImages(AMIinput)
	if err != nil {
		fmt.Println("there was an error getting AMI information: ", err.Error())
	}
	return *AMIData.Images[0].CreationDate, *AMIData.Images[0].Description
}

//  Print out the Security Group access rules
func GetSGPorts(direction string, sgid string) {
	SGinput := &ec2.DescribeSecurityGroupsInput{GroupIds: []*string{aws.String(sgid)}}
	SGData, err := ec2svc.DescribeSecurityGroups(SGinput)
	if err != nil {
		fmt.Println("there was an error getting Security Group information: ", err.Error())
	}

	//  "direction" dictates if we access Ingress or Egress rules
	var sgrules []*ec2.IpPermission
	var Location string
	if direction == "in" {
		sgrules = SGData.SecurityGroups[0].IpPermissions
		Location = "From"
	} else {
		sgrules = SGData.SecurityGroups[0].IpPermissionsEgress
		Location = "To"
	}

	//  Spin though the security group Ingress or Egress rules
	for _, sgrule := range sgrules {
		//  Check if there is a FromPort, if not then set to Null
		var FromPort string
		if sgrule.FromPort == nil {
			FromPort = "Null"
		}

		//  Get list of source or destination IP Addresses
		var mylist string
		for _, ips := range sgrule.IpRanges {
			if mylist == "" {
				mylist = *ips.CidrIp
			} else {
				mylist = mylist + ", " + *ips.CidrIp
			}
		}

		// Get list of source or destination Security Groups
		for _, sgs := range sgrule.UserIdGroupPairs {
			if mylist == "" {
				mylist = *sgs.GroupId
			} else {
				mylist = mylist + ", " + *sgs.GroupId
			}
		}

		if FromPort == "Null" {
			fmt.Printf("      Port: %s | Protocol: %s | %s: %s\n", FromPort, *sgrule.IpProtocol, Location, mylist)
		} else {
			if *sgrule.FromPort == *sgrule.ToPort {
				fmt.Printf("      Port: %d | Protocol: %s | %s: %s\n", *sgrule.FromPort, *sgrule.IpProtocol, Location, mylist)
			} else {
				fmt.Printf("      Port: %d - %d | Protocol: %s | %s: %s\n", *sgrule.FromPort, *sgrule.ToPort, *sgrule.IpProtocol, Location, mylist)
			}
		}
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {
	//  Get command line parms
	RegionPtr := flag.String("R", "", "(required) - AWS Region that you want to view ENI's in")
	IpAddrPtr := flag.String("I", "", "(optional) - The IP Address associated with the specific ENI that you want to view")
	ENIIdPtr := flag.String("E", "", "(optional) - The ENI ID associated with the specific ENI that you want to view")

	flag.Parse()

	// Check that required parms are set
	if *RegionPtr == "" {
		usage()
	} else if *IpAddrPtr != "" && *ENIIdPtr != "" {
		usage()
	}

	fmt.Println("THIS WORKS!!!")
	os.Exit(0)

	//  Setup a session to interract with AWS
	ec2svc = ec2.New(DefineSession(*RegionPtr))

	//  Get EC2 instance information
	InstanceList, err := ec2svc.DescribeInstances(BuildEC2Parms(*IpAddrPtr))
	if err != nil {
		fmt.Println("there was an error listing instances: ", err.Error())
		fmt.Errorf(err.Error())
		os.Exit(1)
	}

	for idx, _ := range InstanceList.Reservations {
		for _, instance := range InstanceList.Reservations[idx].Instances {

			AMIcreation, AMIdescription := GetAMIInfo(*instance.ImageId)

			fmt.Printf("%s:\n", *instance.InstanceId)
			fmt.Printf("  Name: %s\n", GetNameTag(instance.Tags))
			fmt.Printf("  State: %s\n", *instance.State.Name)
			fmt.Printf("  Instance Type: %s\n", *instance.InstanceType)
			fmt.Printf("  AMI: %s [Creation Date: %s]\n", *instance.ImageId, AMIcreation)
			fmt.Printf("  OS: %s\n", AMIdescription)
			fmt.Printf("  Hostname/IP: %s [%s]\n", *instance.PrivateDnsName, *instance.PrivateIpAddress)
			fmt.Printf("  VPC/Network: %s [%s]\n", *instance.VpcId, *instance.SubnetId)
			fmt.Printf("  Security Group Information:\n")

			fmt.Printf("    Security Group Ingress ( allowed inbound connections ):\n")
			for _, secgroup := range instance.SecurityGroups {
				GetSGPorts("in", *secgroup.GroupId)
			}

			fmt.Printf("    Security Group Egress ( allowed outbound connections ):\n")
			for _, secgroup := range instance.SecurityGroups {
				GetSGPorts("out", *secgroup.GroupId)
			}
		}
		fmt.Println("")
	}

}
