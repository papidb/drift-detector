#!/bin/bash

set -euo pipefail

REGION="us-west-2"

echo "Region: $REGION"
echo "⚠️ WARNING: This script will delete EC2 resources in region $REGION"
read -p "Are you sure? (yes/no): " CONFIRM
if [[ "$CONFIRM" != "yes" ]]; then
  echo "Aborted."
  exit 1
fi

echo "🔻 Terminating EC2 Instances..."
INSTANCE_IDS=$(aws ec2 describe-instances --region "$REGION" \
  --query 'Reservations[*].Instances[*].InstanceId' \
  --output text)

if [ -n "$INSTANCE_IDS" ]; then
  aws ec2 terminate-instances --region "$REGION" --instance-ids $INSTANCE_IDS
  echo "Waiting for instances to terminate..."
  aws ec2 wait instance-terminated --region "$REGION" --instance-ids $INSTANCE_IDS
fi

echo "🔻 Deleting Unattached Volumes..."
aws ec2 describe-volumes --region "$REGION" \
  --query 'Volumes[?State==`available`].VolumeId' \
  --output text | xargs -r -n 1 -I {} aws ec2 delete-volume --region "$REGION" --volume-id {}

echo "🔻 Releasing Elastic IPs..."
ALLOC_IDS=$(aws ec2 describe-addresses --region "$REGION" \
  --query 'Addresses[*].AllocationId' \
  --output text)

for alloc_id in $ALLOC_IDS; do
  aws ec2 release-address --region "$REGION" --allocation-id "$alloc_id"
done

echo "🔻 Deleting Available ENIs..."
ENI_IDS=$(aws ec2 describe-network-interfaces --region "$REGION" \
  --query 'NetworkInterfaces[?Status==`available`].NetworkInterfaceId' \
  --output text)

for eni in $ENI_IDS; do
  aws ec2 delete-network-interface --region "$REGION" --network-interface-id "$eni"
done

echo "🔻 Deleting Security Groups (excluding 'default')..."
SG_IDS=$(aws ec2 describe-security-groups --region "$REGION" \
  --query 'SecurityGroups[?GroupName!=`default`].GroupId' \
  --output text)

for sg in $SG_IDS; do
  aws ec2 delete-security-group --region "$REGION" --group-id "$sg"
done

echo "🔻 Deleting Key Pairs..."
KEYS=$(aws ec2 describe-key-pairs --region "$REGION" \
  --query 'KeyPairs[*].KeyName' --output text)

for key in $KEYS; do
  aws ec2 delete-key-pair --region "$REGION" --key-name "$key"
done

echo "✅ EC2 Cleanup complete in $REGION"
