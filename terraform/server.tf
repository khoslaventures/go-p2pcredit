variable "user" {}
variable "resource_count" {}
variable "prefix" {}
variable "commands" {
  type="list"
}


resource "digitalocean_droplet" "servers" {
  count = "${var.resource_count}" # always incrementing, starting from 1
  image = "ubuntu-18-04-x64"
  name = "${var.prefix}${count.index}"
  region = "sfo2"
  size = "1GB"
  ssh_keys = [
    "${var.ssh_fingerprint}"
  ]

  connection {
    user = "${var.user}"
    type = "ssh"
    private_key = "${file(var.pvt_key)}"
    timeout = "2m"
  }

  provisioner "file" {
    source      = "../Archive.zip"
    destination = "/root/app.zip"
  }

  provisioner "remote-exec" {
    inline = ["${var.commands}"]
  }

  provisioner "local-exec" {
    command ="echo '${self.name}:${self.ipv4_address}' >> server_data.txt"
  }
}
