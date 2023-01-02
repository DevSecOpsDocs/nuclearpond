# Nuclear Pond


<img src="assets/logo.png" width="400" height="300" align="right">

Nuclear Pond is used to leverage [Nuclei](https://github.com/projectdiscovery/nuclei) in the cloud with unremarkable speed, flexibility, and perform internet wide scans for far less than a cup of coffee. 

It leverages [AWS Lambda](https://aws.amazon.com/lambda/) as a backend to invoke Nuclei scans in parallel, choice of storing json findings in s3 to query with [AWS Athena](https://aws.amazon.com/athena/), and is easily one of the cheapest ways you can execute scans in the cloud. 

## Features

- Specify any Nuclei arguments as normal
- Output as cmd, json, or to a data lake
- Specify threads and parallel invocations
- Ability to customize batch size

## Usage

Think of Nuclear Pond as just a way for you to run Nuclei in the cloud. You can use it just as you would on your local machine but run them in parallel and with however many hosts you want to specify. All you need to think of is the nuclei command line flags you wish to pass to it. 

## Setup & Installation

To install Nuclear Pond, you need to configure the backend [terraform module](https://github.com/DevSecOpsDocs/terraform-nuclear-pond). You can do this by running `terraform apply`, leveraging terragrunt, and on release we intend to make this easier to deploy. 

```bash
$ go install github.com/DevSecOpsDocs/nuclearpond@latest
```

### Command line flags

In order to specify the command line flags it's important to encode them as base64. When doing so it will send that off to the backend and run directly on nuclei. Any flags available in the current version should be available outside of `-o` and `-json`. 

```
$(echo -ne "-t dns" | base64)
```

### Data Lake Output

This output is recommended when leveraging Nuclear Pond as once the script invokes, all of the work is handed off to the cloud for you to analyze another time. This output is known as `s3` and you can output it by specifying `-o s3`. You can also specify `-l targets.txt` and `-b 10` to invoke the lambda functions in batches of 10 targets. 

```
$ nuclearpond run -t devsecopsdocs.com -a $(echo -ne "-t dns -silent" | base64) -o s3
  _   _                  _                           ____                        _
 | \ | |  _   _    ___  | |   ___    __ _   _ __    |  _ \    ___    _ __     __| |
 |  \| | | | | |  / __| | |  / _ \  / _` | | '__|   | |_) |  / _ \  | '_ \   / _` |
 | |\  | | |_| | | (__  | | |  __/ | (_| | | |      |  __/  | (_) | | | | | | (_| |
 |_| \_|  \__,_|  \___| |_|  \___|  \__,_| |_|      |_|      \___/  |_| |_|  \__,_|

                                                                  devsecopsdocs.com

2023/01/01 16:42:25 Running nuclei against the target devsecopsdocs.com
2023/01/01 16:42:25 Running with 1 threads
2023/01/01 16:42:26 Saved results in s3://jwalker-nuclei-runner-artifacts/findings/2023/01/02/00/nuclei-findings-c69e8359-17b3-4783-b8c4-d6754b3235b8.json
2023/01/01 16:42:26 Completed all parallel operations, best of luck!
```

### Command Line Output

Think of this mechanism as a way to run the CLI directly on the cloud. This allows you to specify

```log
$ nuclearpond run -t devsecopsdocs.com -a $(echo -ne "-t dns -silent" | base64) -o cmd
  _   _                  _                           ____                        _
 | \ | |  _   _    ___  | |   ___    __ _   _ __    |  _ \    ___    _ __     __| |
 |  \| | | | | |  / __| | |  / _ \  / _` | | '__|   | |_) |  / _ \  | '_ \   / _` |
 | |\  | | |_| | | (__  | | |  __/ | (_| | | |      |  __/  | (_) | | | | | | (_| |
 |_| \_|  \__,_|  \___| |_|  \___|  \__,_| |_|      |_|      \___/  |_| |_|  \__,_|

                                                                  devsecopsdocs.com

2023/01/01 16:41:35 Running nuclei against the target devsecopsdocs.com
2023/01/01 16:41:35 Running with 1 threads
[nameserver-fingerprint] [dns] [info] devsecopsdocs.com [ns-487.awsdns-60.com.,ns-579.awsdns-08.net.,ns-1309.awsdns-35.org.,ns-1822.awsdns-35.co.uk.]
[mx-fingerprint] [dns] [info] devsecopsdocs.com [20 mailsec.protonmail.ch.,10 mail.protonmail.ch.]
[txt-fingerprint] [dns] [info] devsecopsdocs.com ["protonmail-verification=14a44944a2577395944d07e38d16139898edee75","v=spf1 include:_spf.protonmail.ch mx ~all"]
[mx-service-detector:ProtonMail] [dns] [info] devsecopsdocs.com

2023/01/01 16:41:35 Completed all parallel operations, best of luck!
```

## Retrieving Findings

To explore your findings in Athena all you need to do is perform the following query! The database and the table should already be available to you. You may have to configure query results if you have not done so already. 

```sql
select
  *
from
  nuclei_db.findings_db
limit 10;
```

### Advance Query

In order to get down into queries a little deeper, I thought I would give you a quick example. In the select statement we drill down into `info` column, `"matched-at"` column must be in double quotes due to `-` character, and you are searching only for high and critical findings generated by Nuclei.

```sql
SELECT
  info.name,
  host,
  type,
  info.severity,
  "matched-at",
  info.description,
  template,
  dt
FROM 
  "nuclei_db"."findings_db"
where 
  host like '%devsecopsdocs.com'
  and info.severity in ('high','critical')
```

## Infrastructure

The backend infrastructure, all within [terraform module](https://github.com/DevSecOpsDocs/terraform-nuclear-pond). I would strongly recommend reading the readme associated to it as it will have some important notes. 

- Lambda function
- S3 bucket
  - Stores nuclei binary
  - Stores configuration files
  - Stores findings
- Glue Database and Table
  - Allows you to query the findings in S3
- IAM Role for Lambda Function
