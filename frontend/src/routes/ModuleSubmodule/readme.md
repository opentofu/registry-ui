# Terraform Module: Example Module

![Terraform](https://img.shields.io/badge/react-v1.0+-blue.svg) ![License](https://img.shields.io/github/license/facebook/react.svg)

This Terraform module sets up an example infrastructure with various resources. It is designed to be simple yet comprehensive, demonstrating how to structure a Terraform module.

## Features

- Create an EC2 instance
- Setup a VPC with subnets
- Provision an S3 bucket
- Configure security groups
- Enable CloudWatch logging

## Table of Contents

- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Examples](#examples)
- [Inputs](#inputs)
- [Outputs](#outputs)
- [Resources](#resources)
- [Contributing](#contributing)
- [License](#license)

## Prerequisites

Ensure you have the following installed:

- [Terraform](https://www.terraform.io/downloads.html) v1.0+
- [AWS CLI](https://aws.amazon.com/cli/)

## Usage

```hcl
module "example" {
  source  = "github.com/yourusername/example-terraform-module"
  region  = "us-west-2"
  vpc_cidr = "10.0.0.0/16"

  # EC2 instance
  instance_type = "t2.micro"

  # S3 bucket
  bucket_name   = "example-terraform-bucket"
}
```

## Examples

### Basic Example

A simple example of using this module:

```hcl
provider "aws" {
  region = "us-west-2"
}

module "example" {
  source  = "github.com/yourusername/example-terraform-module"
  region  = "us-west-2"
  vpc_cidr = "10.0.0.0/16"

  instance_type = "t2.micro"
  bucket_name   = "example-terraform-bucket"
}
```

### Advanced Example

An advanced example with more custom configurations:

```hcl
provider "aws" {
  region = "us-west-2"
}

module "example" {
  source       = "github.com/yourusername/example-terraform-module"
  region       = "us-west-2"
  vpc_cidr     = "10.0.0.0/16"

  instance_type       = "t3.medium"
  bucket_name         = "advanced-terraform-bucket"
  enable_cloudwatch   = true
  security_group_ids  = ["sg-0123456789abcdef0"]
}
```

## Inputs

| Name                 | Description                    | Type         | Default       | Required |
| -------------------- | ------------------------------ | ------------ | ------------- | -------- |
| `region`             | AWS region                     | string       | `us-west-2`   | yes      |
| `vpc_cidr`           | CIDR block for the VPC         | string       | `10.0.0.0/16` | yes      |
| `instance_type`      | Type of EC2 instance to launch | string       | `t2.micro`    | yes      |
| `bucket_name`        | Name of the S3 bucket          | string       | n/a           | yes      |
| `enable_cloudwatch`  | Enable CloudWatch logging      | bool         | `false`       | no       |
| `security_group_ids` | List of security group IDs     | list(string) | `[]`          | no       |

## Outputs

| Name          | Description                |
| ------------- | -------------------------- |
| `vpc_id`      | The ID of the VPC          |
| `subnet_ids`  | The IDs of the subnets     |
| `instance_id` | The ID of the EC2 instance |
| `bucket_arn`  | The ARN of the S3 bucket   |

## Resources

This module creates the following resources:

- `aws_vpc`
- `aws_subnet`
- `aws_instance`
- `aws_s3_bucket`
- `aws_security_group`
- `aws_cloudwatch_log_group` (optional)

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/your-feature`).
3. Commit your changes (`git commit -am 'Add your feature'`).
4. Push to the branch (`git push origin feature/your-feature`).
5. Create a new Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

If you have any questions or need further assistance, feel free to open an issue or contact the maintainers.
