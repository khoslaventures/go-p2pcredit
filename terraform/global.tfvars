user = "root"
resource_count = "3"
prefix = "trustline"
commands = [
      "sudo apt-get update",
      "sudo apt-get -y upgrade",
      "sudo apt install -y unzip gcc",
      "sudo snap install go --classic",
      "unzip app.zip -d p2pcredit",
    ]
