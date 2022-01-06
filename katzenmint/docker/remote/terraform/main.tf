#Terraform Configuration
terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

variable "DO_API_TOKEN" {
  description = "DigitalOcean Access Token"
}

variable "TESTNET_NAME" {
  description = "Name of the testnet"
  default = "cyclotron"
}

variable "SSH_KEY_FILE" {
  description = "SSH public key file to be used on the nodes"
  type = string
}

variable "INSTANCE_SIZE" {
  description = "The instance size to use"
  default = "s-1vcpu-1gb"
}

variable "KSERVERS" {
  description = "Number of katzenint nodes in testnet"
  default = "3"
}

variable "MSERVERS" {
  description = "Number of mix nodes in testnet"
  default = "3"
}

variable "PSERVERS" {
  description = "Number of provider nodes in testnet"
  default = "2"
}

provider "digitalocean" {
  token = "${var.DO_API_TOKEN}"
}

module "cluster" {
  source           = "./cluster"
  name             = "${var.TESTNET_NAME}"
  ssh_key          = "${var.SSH_KEY_FILE}"
  instance_size    = "${var.INSTANCE_SIZE}"
  kservers         = "${var.KSERVERS}"
  mservers         = "${var.MSERVERS}"
  pservers         = "${var.PSERVERS}"
}


output "public_ips" {
  value = "${module.cluster.public_ips}"
}

