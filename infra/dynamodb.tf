resource "aws_dynamodb_table" "videos" {
  name         = "Videos"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "PK"

  attribute {
    name = "PK"
    type = "S"
  }

  tags = {
    Name        = "Videos"
    Environment = "production"
  }
}
