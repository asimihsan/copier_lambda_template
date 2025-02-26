terraform {
  backend "s3" {}
}

provider "aws" {
  region = "us-west-2"
}

// TODO: Implement AWS Lambda function resources when ready.

# Example: Lambda
# resource "aws_lambda_function" "tokenissuer" {
#   function_name = "${var.project_prefix}-tokenissuer"
#   # ...
# }
