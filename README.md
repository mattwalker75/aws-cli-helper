# aws-cli-helper

View your AWS infrastructure setup in a user friendly format from the CLI

### Requirements

Make sure you have the following tools downloaded and installed on your computer:

* GoLang ( https://golang.org/dl/ )
* make ( Or your favorite flavor of make )
   * NOTE:  The make program is only needed for non-Windows users.  Windows users will have to perform manual code builds.

### Installation

#### For Mac and Linux Users

For Mac and Linux users, run the following command to download all the required dependencies and compile the programs for use:

```
make install
```
All compiled programs will be located in the `bin` directory.

There are a number of other options you can select from.  Run the following command to view those options:

```
make help
```

#### For Windows Users

For Windows users, run the following set of commands to compile the GoLang programs.  You will need to be in the root directory
that contains the content of this Repo.

Run the following command to download required packages:

```
GOPATH=$(pwd) go get -d
```

Run the following set of commands to build the binary files:

```
GOPATH=$(pwd) go build ListEC2.go
GOPATH=$(pwd) go build ListENIs.go
GOPATH=$(pwd) go build VPCFlowLogData.go

```

#### For Cross Platform compiling

Here you will see some examples of how to compile for a different OS and CPU architecture, such as compiling a Windows binary
from a Linux system.

Run the following command to download required packages:

```
GOPATH=$(pwd) go get -d
```

Run the following command to compile the binary for a Mac:

```
GOPATH=$(pwd) GOOS=darwin GOARCH=amd64 go build ListENIs.go
```

Run the following comand to compile the binary for Linux:

```
GOPATH=$(pwd) GOOS=linux GOARCH=amd64 go build ListENIs.go
```

Run the following command to compile the binary for Windows:

```
GOPATH=$(pwd) GOOS=windows GOARCH=amd64 go build ListENIs.g
```

### Usage

#### ListEC2
This is used to view useful information about your EC2 instances, such as VPC and subnet location, DNS and IP 
information, AMI information, and all inbound and outbound network port access.  

Run the following command to get a list of commandline arguments:

```
./ListEC2
```

#### ListENIs
This is used to view useful information about Elastic Network Interfaces (ENI's), such as private and public
DNS names and IP Addresses, description of the ENI, VPC and subnet location, and the type of interface.

Run the following command to get a list of commandline arguments:

```
./ListENIs
```

#### VPCFlowLogData
This is used to view the VPC Flow Log data associated with your AWS VPC.  It takes the current VPC Flow Log 
data and reformats it into an easy to read format consisting of the following:

```
<date> : <ENI ID> : <Source IP>[<source port>] --> <destination IP>[<destination port>] : <protocol> : <status>
```

Run the following command to get a list of commandline arguments:

```
./VPCFlowLogData
```

### Want to contribute?

After you added your code updates, compiling it, and testing it, you can run the following command
to clean up your local copy of the repo to help make it ready to be uploaded to your remote branch:
 

```
make reset
```

The above command will perform the following tasks:

* Cleans up the GoLang build cache
* Deletes the `bin` directory and all of its content
* Deletes the `vendor` directory and all of its content


