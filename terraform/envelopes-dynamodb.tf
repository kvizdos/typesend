resource "aws_dynamodb_table" "typesend_envelopes" {
  name         = "${vars.project}_typesend_envelopes"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  # We'll need these attributes for the GSIs.
  attribute {
    name = "status"
    type = "N"
  }

  attribute {
    name = "scheduledFor"
    type = "S"
  }

  attribute {
    name = "to"
    type = "S"
  }

  attribute {
    name = "toInternal"
    type = "S"
  }

  attribute {
    name = "ref"
    type = "S"
  }

  global_secondary_index {
    name            = "status-scheduledFor-index"
    hash_key        = "status"
    range_key       = "scheduledFor"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "to-index"
    hash_key        = "to"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "toInternal-index"
    hash_key        = "toInternal"
    projection_type = "ALL"
  }
}
