# Docs
This is the documentation related to the Meson mixnet project. 
You will find the instructions on how to run a local testnet instance in addition to how to run the wallet demo.

Ensure that you have both docker and python installed on your system.

## Running Meson

### Integration test on local computer

#### Prequisite
* docker
https://www.docker.com/

* python
https://www.python.org

#### Steps
* Clone Meson repository
```
$ git clone https://github.com/hashcloak/Meson.git
```

* Build containers for Meson networks
You'll build these containers after execute this script: Katzenmint pki and Meson server:
```
$ python plugin/ops/build_containers.py
```

* Start testnet
You can start Meson testnet on local computer after building containers:
```
$ cd testnet/local
$ docker compose up
```

5. Execute ping test
Try our integration test (ping):
```
$ cd ping
$ go run main.go -s echo
```

### Deploy to Remote Network

#### Prequisite
* terraform
https://www.terraform.io/

* ansible
https://www.ansible.com/

* added ssh public key in Digital Ocean
https://docs.digitalocean.com/products/droplets/how-to/add-ssh-keys/

#### Services

##### Meson mix services

* create droplets on Digital Ocean
    * go to `docker/remote/terraform`
    * apply terraform config`terraform apply -var DO_API_TOKEN="$DO_API_TOKEN" -var SSH_KEY_FILE="$SSH_KEY_FILE"`
* remove droplets on Digital Ocean
    * go to `docker/remote/terraform`
    * `terraform destroy -var DO_API_TOKEN="$DO_API_TOKEN" -var SSH_KEY_FILE="$SSH_KEY_FILE"`
        > if you want to remove specific droplet, added -target: -target="module.cluster.digitalocean_droplet.cluster[3]"
* katzenmint-pki
    * go to `docker/remote/ansible`
    * install the service `ansible-playbook -i inventory/digital_ocean.py -l sentrynet install.yml`
    * upload config and binary `ansible-playbook -i inventory/digital_ocean.py -l sentrynet config.yml -e CONFIGDIR=/path/to/config/directory -e BINARY=/path/to/binary`
* mix
    * go to `docker/remote/ansible`
    * install the service `ansible-playbook -i inventory/digital_ocean.py -l mixnet install.yml`
    * upload config and binary `ansible-playbook -i inventory/digital_ocean.py -l mixnet config.yml -e CONFIGDIR=/path/to/config/directory -e BINARY=/path/to/binary`
* provider
    * go to `docker/remote/ansible`
    * install the service `ansible-playbook -i inventory/digital_ocean.py -l providernet install.yml`
    * upload config and binary `ansible-playbook -i inventory/digital_ocean.py -l providernet config.yml -e CONFIGDIR=/path/to/config/directory -e BINARYDIR=/path/to/binary/directory`
