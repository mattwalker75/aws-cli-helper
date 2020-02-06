# aws-cli-helper

View your AWS infrastructure setup in a user friendly format from the CLI

### Requirements

Make sure you have the following tools downloaded and installed on your computer:

* [GoLang][https://golang.org/dl/]
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

Here are the notes on how to use the different tools....

```
Example of how to run a command
```

### Want to contribute?

After you added your code updates, compiling it, and testing it, you can run the following command
to clean up your local copy of the repo to help make it ready to be uploaded to your remote branch:
 

```
make reset
```

The above command will perform the following tasks:

* Cleans up the GoLang build cache
* Deletes the bin directory and all of its content
* Deletes teh vendor directory and all of its content


