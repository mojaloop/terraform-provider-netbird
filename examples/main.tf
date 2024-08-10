terraform {
  required_providers {
    netbird = {
      source = "github.com/netbirdio/netbird"
    }
  }
}



provider "netbird" {
  server_url = "https://netbird.cc50.ccnew.mojaloop.live:443"
  token_auth = "<<replace this value>>"
}

resource "netbird_setup_key" "tf_test_key_2" {
  name        = "tf_linux_key_8"
  type        = "one-off"
  auto_groups = [netbird_group.test.id]
  ephemeral   = true
  usage_limit = 1
  expires_in  = 86400
}

resource "netbird_group" "test" {
  name = "test_group8"
}

# resource "netbird_route" "test_route" {

# }
