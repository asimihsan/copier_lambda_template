terraform {
  backend "s3" {}
}

provider "aws" {
  region = "us-west-2"
}

resource "aws_dynamodb_table" "deployment_rotation" {
  name         = "${var.project_prefix}-deployment-rotation"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "rotation_id"

  attribute {
    name = "rotation_id"
    type = "S"
  }
}

resource "aws_dynamodb_table" "override_requests" {
  name         = "${var.project_prefix}-override-requests"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "override_id"

  attribute {
    name = "override_id"
    type = "S"
  }
}


// TODO: Implement AWS Lambda function resources when ready.

# Example: Lambda
# resource "aws_lambda_function" "tokenissuer" {
#   function_name = "${var.project_prefix}-tokenissuer"
#   # ...
# }
