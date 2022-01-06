variable "name" {
  description = "The cluster name, e.g cdn"
}

variable "regions" {
  description = "Regions to launch in"
  type = list
  default = ["NYC1", "NYC3", "FRA1", "LON1", "AMS3", "SGP1", "TOR1", "BLR1", "SFO1", "SFO2", "SFO3"]
}

variable "ssh_key" {
  description = "SSH key filename to copy to the nodes"
  type = string
}

variable "instance_size" {
  description = "The instance size to use"
  default = "s-1vcpu-1gb"
}

variable "kservers" {
  description = "Desired katzenmint instance count"
  default     = 3
}

variable "mservers" {
  description = "Desired mix instance count"
  default     = 3
}

variable "pservers" {
  description = "Desired provider instance count"
  default     = 2
}

variable "tags" {
  description = "Tags for droplet"
  type = list
  default = ["sentrynet", "sentrynet", "sentrynet", "mixnet", "mixnet", "mixnet", "providernet", "providernet"]
}

