terraform {
  backend "s3" {}
}

variable "project_prefix" {
  type = string
}

provider "aws" {
  region = "us-west-2"
}

# Example: Lambda
# resource "aws_lambda_function" "tokenissuer" {
#   function_name = "${var.project_prefix}-tokenissuer"
#   # ...
# }
