INSTANCE_ID=i-0bea3299963d3575f

aws ec2 stop-instances --instance-ids $INSTANCE_ID
aws ec2 wait instance-stopped --instance-ids $INSTANCE_ID

aws ec2 create-tags \
  --resources $INSTANCE_ID \
  --tags Key=Name,Value=DriftTestInstance

aws ec2 modify-instance-attribute \
  --instance-id $INSTANCE_ID \
  --source-dest-check "{\"Value\": false}"

aws ec2 modify-instance-metadata-options \
  --instance-id $INSTANCE_ID \
  --http-endpoint enabled \
  --http-tokens required

aws ec2 start-instances --instance-ids $INSTANCE_ID
aws ec2 wait instance-running --instance-ids $INSTANCE_ID
