provider "sops" {}

data "sops_file" "lambda_secrets" {
  source_file = "${var.repo_root_dir}/secrets/lambda-secrets.yaml"
}

# Create AWS Secrets Manager secret
resource "aws_secretsmanager_secret" "lambda_secrets" {
  name        = "${var.project_prefix}-lambda-secrets"
  description = "Secrets for Lambda function"
}

resource "aws_secretsmanager_secret_version" "lambda_secrets" {
  secret_id = aws_secretsmanager_secret.lambda_secrets.id
  secret_string = jsonencode({
    SLACK_APP_TOKEN      = data.sops_file.lambda_secrets.data["SLACK_APP_TOKEN"],
    SLACK_BOT_TOKEN      = data.sops_file.lambda_secrets.data["SLACK_BOT_TOKEN"],
    SLACK_SIGNING_SECRET = data.sops_file.lambda_secrets.data["SLACK_SIGNING_SECRET"],
  })
}

# Grant Lambda execution role access to the secret
resource "aws_iam_policy" "lambda_secrets_access" {
  name        = "${var.project_prefix}-lambda-secrets-access"
  description = "Allow Lambda to access secrets"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = [
          "secretsmanager:GetSecretValue",
        ],
        Effect   = "Allow",
        Resource = aws_secretsmanager_secret.lambda_secrets.arn
      }
    ]
  })
}

resource "aws_iam_policy_attachment" "lambda_secrets_policy_attach" {
  name       = "${var.project_prefix}-lambda-secrets-policy-attach"
  roles      = [aws_iam_role.lambda_exec_role.name]
  policy_arn = aws_iam_policy.lambda_secrets_access.arn
}

