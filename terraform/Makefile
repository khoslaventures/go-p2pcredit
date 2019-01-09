# Makefile

SHELL=/bin/bash

include ~/.do_token
export

destroy:
	rm -f server_data.txt
	# Requires DO_PAT to be set to DigitalOcean API Token and
	# SSH_FINGERPRINT set to the SSH fingerprint for the given public key
	terraform destroy \
		-var "do_token=${DO_PAT}" \
		-var "pub_key=${HOME}/.ssh/id_ed25519.pub" \
		-var "pvt_key=${HOME}/.ssh/id_ed25519" \
		-var "ssh_fingerprint=${SSH_FINGERPRINT}" \
		-var-file="global.tfvars"

apply:
	# Requires DO_PAT to be set to DigitalOcean API Token and
	# SSH_FINGERPRINT set to the SSH fingerprint for the given public key
	terraform apply \
		-var "do_token=${DO_PAT}" \
		-var "pub_key=${HOME}/.ssh/id_ed25519.pub" \
		-var "pvt_key=${HOME}/.ssh/id_ed25519" \
		-var "ssh_fingerprint=${SSH_FINGERPRINT}" \
		-var-file="global.tfvars"

validate:
	# Requires DO_PAT to be set to DigitalOcean API Token and
	# SSH_FINGERPRINT set to the SSH fingerprint for the given public key
	terraform validate \
		-var "do_token=${DO_PAT}" \
		-var "pub_key=${HOME}/.ssh/id_ed25519.pub" \
		-var "pvt_key=${HOME}/.ssh/id_ed25519" \
		-var "ssh_fingerprint=${SSH_FINGERPRINT}" \
		-var-file="global.tfvars"
