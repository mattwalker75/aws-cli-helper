# aws-cli-helper

View your AWS infrastructure setup in a user friendly format from the CLI

### Requirements

Make sure you have the following tools downloaded and installed on your computer:

* GoLang ( https://golang.org/dl/ )
* make ( Or your favorite flavor of make )

### Installation

Run the following command to download all the required dependencies and compile the programs for use:

```
make install
```

There are a number of other options you can select from.  Run the following command to view those options:

```
make help
```

### Usage

##### ListEC2
This is used to view useful information about your EC2 instances, such as VPC and subnet location, DNS and IP 
information, AMI information, and all inbound and outbound network port access.  

Run the following command to get a list of commandline arguments:

```
./ListEC2
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


