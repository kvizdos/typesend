resource "aws_dynamodb_table" "typesend_templates" {
  name         = "${vars.project}_typesend_templates"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  range_key    = "tenant"

  attribute {
    name = "tenant"
    type = "S"
  }

  attribute {
    name = "id"
    type = "S"
  }
}
