packer {
  required_plugins {
    qemu = {
      version = ">= 1.1.4"
      source  = "github.com/hashicorp/qemu"
    }
  }
}

variable "image_url" {
  type    = string
  default = "https://cloud-images.ubuntu.com/releases/24.04/release/ubuntu-24.04-server-cloudimg-amd64.img"
}

variable "image_checksum" {
  type    = string
  default = "sha256:834af9cd766d1fd86eca156db7dff34c3713fbbc7f5507a3269be2a72d2d1820"
}

source "qemu" "ubuntu" {
  iso_url              = var.image_url
  iso_checksum         = var.image_checksum

  disk_image           = true
  output_directory     = "artifacts"

  disk_interface       = "virtio"
  net_device           = "virtio-net"
  
  disk_size            = "8G"
  format               = "qcow2"
  disk_compression     = true
  accelerator          = "kvm"
  headless             = true

  qemuargs = [
    ["-cdrom", "cidata.iso"]
  ]

  ssh_username         = "ubuntu"
  ssh_password         = "supersecret"
  ssh_timeout          = "10m"
  shutdown_command     = "echo 'supersecret' | sudo -S shutdown -P now"
}

build {
  sources = ["source.qemu.ubuntu"]
  
  provisioner "shell" {
    inline = [
        "while [ ! -f /var/lib/cloud/instance/boot-finished ]; do echo 'Waiting for Cloud-Init...'; sleep 1; done",
    ]
  }
  
  provisioner "shell" {
    inline = [
      "sudo apt-get clean",
      "sudo rm -rf /var/lib/apt/lists/*",
      "sudo fstrim -av"
    ]
  }
}
