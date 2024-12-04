terraform {
  required_version = "~> 1.8"

  required_providers {
    ghsecret = {
      source  = "nobbs/ghsecret"
      version = "~> 0.0.1"
    }

    github = {
      source  = "integrations/github"
      version = "~> 6.0"
    }
  }
}

# There are no configuration options
provider "ghsecret" {}
