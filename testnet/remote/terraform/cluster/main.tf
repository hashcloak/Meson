terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

resource "digitalocean_tag" "cluster" {
  name = "${var.name}"
}

resource "digitalocean_ssh_key" "cluster" {
  name       = "${var.name}"
  public_key = "${file(var.ssh_key)}"
}

resource "digitalocean_droplet" "cluster" {
  name = "${var.name}-${element(var.tags, count.index)}-node${count.index}"
  image = "centos-7-x64"
  size = "${var.instance_size}"
  region = "${element(var.regions, count.index)}"
  ssh_keys = ["${digitalocean_ssh_key.cluster.id}"]
  count = "${var.kservers+var.mservers+var.pservers}"
  tags = ["${element(var.tags, count.index)}"]

  connection {
    timeout = "30s"
  }

}

