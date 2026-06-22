terraform {
  required_version = ">= 1.0.0"

  backend "s3" {
    bucket = "scalablecounter-terraform-state-646731024202"
    key    = "scalablecounter/terraform.tfstate"
    region = "eu-central-1"

    encrypt = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "eu-central-1"
}
