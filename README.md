### AWS EC2 Spot Price

This tool collects data over a historic time span for an EC2 instance and gives back the price list, favoring the Availability Zone where running the instance is the cheapest available option in time of execution with ~30% higher price then the actual one to take into acount price variation over future time period.

### Requirements

```json
# custom permission
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "ec2:DescribeSpotPriceHistory",
            "Resource": "*"
        }
    ]
}
```
- go version go1.14.1 (should work with other versions as well that support `go mod`)
- AWS **access key ID / secret access key** required with the following **IAM** permissions:
  - use the custom permission from above to allow usage of  `DescribeSpotPriceHistory API`
    - [DescribeSpotPriceHistory](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html) should provide additional information

### Installation
```bash
> git clone https://github.com/koceg/awsEC2Spot.git
> cd awsEC2Spot
> go build
> sudo cp ./awsEC2Spot /usr/bin/awsEC2Spot # change permissions,ownership if neceserry
```
### Configuration
**~/.aws/config** file structure
```bash
[default]
region = <aws_default_region>
```

**~/.aws/credentials** file structure
```bash
[default]
aws_access_key_id = XXX
aws_secret_access_key = XXX
```
### Usage

```bash
> awsEC2Spot -h # usage explanation

> awsEC2Spot -z eu-central-1 -i m5.large 1  #would return Linux instance price by default

eu-central-1c 0.044892
eu-central-1b 0.045279
eu-central-1a 0.046053

awsEC2Spot -z eu-west-1 -i m5.large -d Windows 1  # Windows
awsEC2Spot -z eu-west-1 -i m5.large -d 'Red Hat Enterprise Linux' 1  # RedHat
awsEC2Spot -z eu-central-1 -i m5.large -d 'SUSE Linux' 1 # SUSE

# if you don't get an error but no output you probably have a typo for the filter used
```

**NOTE:** refer to the link in Requirements for other valid product and instance types