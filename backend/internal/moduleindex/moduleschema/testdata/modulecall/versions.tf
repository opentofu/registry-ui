terraform {
  required_providers {
    opentofu = {
      source  = "opentofu/opentofu"
      version = "~> 1.6.0"
    }

    ad = {
      source  = "opentofu/ad"
      version = "0.5.0"
    }
  }
}
