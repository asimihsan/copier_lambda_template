# output "tokenissuer_function_arn" {
#   value = aws_lambda_function.tokenissuer.arn
# }

output "dynamodb_table_deployment_rotation" {
  value = aws_dynamodb_table.deployment_rotation.name
}

output "dynamodb_table_override_requests" {
  value = aws_dynamodb_table.override_requests.name
}
