# Wireguard VPN

## Instructions

Create a private and public Wireguard key and upload the resultant value to SSM parameter store.

The script will also create a public key.

```sh
export PRIVATE_KEY=`wg genkey`
aws ssm put-parameter --name "wireguardPrivateKey" \
    --type "String" \
    --value $PRIVATE_KEY \
    --overwrite
export PUBLIC_KEY=`echo $PRIVATE_KEY | wg pubkey`
echo "Public key to give to clients:" $PUBLIC_KEY
```

Run in the cdk with `cdk deploy`.

The output will include the `wireguardPublicIp`. Use this in the client configuration.

If you need to get the public key again, you can run:

```
aws ssm get-parameter --name "wireguardPrivateKey" --query 'Parameter.Value' --output text | wg pubkey
```

## Configuring clients

Configure the client as below, then update the `user-data.sh` file to add additional `[Peer]` sections for each client, assigning each client (defined by its public key), an IP address.

Match the address of the client with the allowed peer.

```
[Peer]
PublicKey = U4d+Xy7tumEDxnkdFUNsQXJaXOIe6ipWH9jg1Al9gxU=
AllowedIPs = 10.8.0.3/32
```

### MacOS

- https://serversideup.net/how-to-configure-a-wireguard-macos-client/

Update the configuration to use the `PublicKey` of the server, and the value of the `wireguardPublicIp` output variable as the `Endpoint`.

#### Client configuration

```
[Interface]
PrivateKey = <your-machine's-private-key-not-the-server's>
Address = 10.0.0.3/24
DNS = 1.1.1.1, 1.0.0.1

[Peer]
PublicKey = <your-server's-public-key>
Endpoint = <your-server'sip>:51820
AllowedIPs = 0.0.0.0/0
```

