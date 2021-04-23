# Terraform version and plugin versions

terraform {
  required_version = ">= 0.13"

  required_providers {
     oci = {
      source = "hashicorp/oci"
      version = "4.21.0"
    }
    ct = {
      source  = "poseidon/ct"
      version = "0.8.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.0.0"
    }
    template = {
      source  = "hashicorp/template"
      version = "2.2.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.0.0"
    }
  }
}
