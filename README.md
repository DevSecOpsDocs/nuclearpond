# Nuclear Pond


<img src="assets/logo.png" width="400" height="300" align="right">

Nuclear Pond is used to leverage [Nuclei](https://github.com/projectdiscovery/nuclei) in the cloud with unremarkable speed, flexibility, and perform internet wide scans for far less than a cup of coffee. 

It leverages [AWS Lambda](https://aws.amazon.com/lambda/) as a backend to invoke Nuclei scans in parallel, choice of storing json findings in s3 to query with [AWS Athena](https://aws.amazon.com/athena/), and is easily one of the cheapest ways you can execute scans in the cloud. 

## Features

- Specify any Nuclei arguments as normal
- Output as cmd, json, or to a data lake
- Specify threads and parallel invocations
- Ability to customize batch size

## Setup & Installation

To install Nuclear Pond, you need to configure the backend [terraform module](https://github.com/DevSecOpsDocs/terraform-nuclear-pond). You can do this by running `terraform apply`, leveraging terragrunt, and on release we intend to make this easier to deploy. 

```bash
$ go get github.com/DevSecOpsDocs/nuclear-pond
```

## Infrastructure

![Infrastructure](/assets/infrastructure.png)

- Lambda function
- S3 bucket
  - Stores nuclei binary
  - Stores configuration files
  - Stores findings
- Glue Database and Table
  - Allows you to query the findings in S3
- IAM Role for Lambda Function
