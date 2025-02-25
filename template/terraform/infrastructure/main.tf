terraform {
  backend "s3" {}
}

variable "project_prefix" {
  type    = string
  default = "example-prefix"
}

provider "aws" {
  region = "us-west-2"
}

# Example: ECR Repo
resource "aws_ecr_repository" "repo" {
  name = "${var.project_prefix}-repo"
}
