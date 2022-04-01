#!/bin/sh
# Install SSM agent to be able to log in remotely.
sudo yum install -y https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_arm64/amazon-ssm-agent.rpm
restart amazon-ssm-agent

