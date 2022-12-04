resource "splunkacs_hec_token" "example" {
  name              = "example"
  allowed_indexes   = ["main"]
  default_index     = "main"
  default_source    = "hec"
  default_sourceype = "_json"
  use_ack           = false
}