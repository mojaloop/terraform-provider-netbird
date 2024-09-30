terraform {
  required_providers {
    netbird = {
      source = "github.com/netbirdio/netbird"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.12.1"
    }
  }
}
provider "time" {
}


provider "netbird" {
  server_url        = "https://netbird.cntrlcntr.ccv2.mojaloop.live:443"
  oauth_credentials = "${path.cwd}/netbird-oauth.json"
  oauth_issuer      = "https://zitadel.cntrlcntr.ccv2.mojaloop.live/oauth/v2/token"
}

resource "netbird_setup_key" "tf_test_key_2" {
  name        = "tf_linux_key_8"
  type        = "one-off"
  auto_groups = [netbird_group.test.id]
  ephemeral   = true
  usage_limit = 1
  expires_in  = 86400
}
resource "time_rotating" "setup_key_rotation" {
  rotation_minutes = 3
}

resource "netbird_setup_key" "tf_key_rotate" {
  name        = "tf_linux_key_rotate"
  type        = "reusable"
  auto_groups = [netbird_group.test.id]
  ephemeral   = false
  usage_limit = 0
  expires_in  = 86400
  rotation_id = time_rotating.setup_key_rotation.id
}

resource "netbird_group" "test" {
  name = "test_group8"
}

resource "netbird_group" "test_gw" {
  name = "gw-test-8"
}

resource "netbird_route" "test_route" {
  description = "testroute8"
  enabled     = true
  groups      = [netbird_group.test.id, local.user_group_id]
  keep_route  = false
  masquerade  = true
  metric      = 9999
  peer_groups = [netbird_group.test_gw.id]
  network     = "10.10.10.0/24"
  network_id  = "testroute8"
}


output "cc50_gw_route" {
  value = netbird_route.test_route.id
}

data "netbird_groups" "all" {
}
locals {
  user_group_id = [for group in data.netbird_groups.all.groups : group.id if strcontains(group.name, var.user_group_name)][0]
}

variable "user_group_name" {
  type    = string
  default = "techops-users"
}
