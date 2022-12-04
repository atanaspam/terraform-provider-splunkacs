resource "splunkacs_index" "test" {
  name             = "example"
  data_type        = "event"
  searchable_days  = 30
  max_data_size_mb = 0
}