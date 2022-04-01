#!/bin/sh

# Ubuntu install SSM by default.
# Access the server with a command like this.
# aws ssm start-session --target i-04045d762ce159c7b

# We followed this tutorial.
# https://www.digitalocean.com/community/tutorials/how-to-set-up-wireguard-on-ubuntu-20-04

# Update dependencies and install wireguard.
sudo apt-get update
sudo apt-get install -y wireguard

# Retrieve the secret from SSM parameter store.
export PRIVATE_KEY=`aws ssm get-parameter --name "wireguardPrivateKey" --query 'Parameter.Value' --output text`

aws ssm get-parameter --name "wireguardPrivateKey" \
    --type "String" \
    --value $PRIVATE_KEY \
    --overwrite

# Create the configuration.
# Note that the Wireguard clients are defined in each Peer section.
# Using the Wireguard client creates a public key for each client. Use this one, not the server's public key.
# See https://serversideup.net/how-to-configure-a-wireguard-macos-client/
sudo tee /etc/wireguard/wg0.conf > /dev/null << EOF
[Interface]
PrivateKey = $PRIVATE_KEY
Address = 10.8.0.1/24
ListenPort = 51820
SaveConfig = true
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o ens5 -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o ens5 -j MASQUERADE

[Peer]
PublicKey = U4d+Xy7tumEDxnkdFUNsQXJaXOIe6ipWH9jg1Al9gxU=
AllowedIPs = 10.8.0.3/32
EOF
sudo chmod go= /etc/wireguard/private.key
sudo cat /etc/wireguard/private.key | wg pubkey | sudo tee /etc/wireguard/public.key

# Enable port forwarding.
echo net.ipv4.ip_forward = 1 | sudo tee -a /etc/sysctl.conf
echo net.ipv6.conf.all.forwarding=1 | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Enable access to the Wireguard UDP port.
sudo ufw allow 51820/udp
sudo ufw enable

# Run as systemd unit.
sudo systemctl enable wg-quick@wg0.service
sudo systemctl start wg-quick@wg0.service
