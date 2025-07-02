SNS Publish Access Policy
`sns-publish-policy-dev.json`
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sns:Publish"
      ],
      "Resource": [
        "arn:aws:sns:us-east-1:616293268143:mcb-checkboxAction-dev.fifo",
        "arn:aws:sns:us-east-1:616293268143:mcb-checkboxActionResult-dev.fifo",
        "arn:aws:sns:us-east-1:616293268143:mcb-checkboxActionFailure-dev.fifo"
      ]
    }
  ]
}
```

Create and attach the policy:
```bash
# Create the policy
aws iam create-policy --profile mcb_admin \
  --policy-name mcb-sns-publish-dev \
  --policy-document file://mcb_sns_publish_dev_policy.json

# Attach to group
aws iam attach-group-policy --profile mcb_admin \
  --group-name mcb_client_dev \
  --policy-arn arn:aws:iam::616293268143:policy/mcb-sns-publish-dev
```

SQS Access Policy
`mcb-sns-sqs-policy-dev.json`
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sns:Publish"
      ],
      "Resource": [
        "arn:aws:sns:us-east-1:616293268143:mcb-*-dev.fifo"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes",
        "sqs:ChangeMessageVisibility"
      ],
      "Resource": [
        "arn:aws:sqs:us-east-1:616293268143:mcb-*-dev.fifo"
      ]
    }
  ]
}
```

Create and Attach the IAM policy
```bash
# create the IAM Policy
aws iam create-policy --profile mcb_admin \
  --policy-name mcb-sns-sqs-dev \
  --policy-document file://mcb-sns-sqs-policy-dev.json \
  --description "Allow SNS publish and SQS consume for mcb dev environment"
  
# attach the policy to the group
aws iam attach-group-policy --profile mcb_admin \
  --group-name mcb_client_dev \
  --policy-arn arn:aws:iam::616293268143:policy/mcb-sns-sqs-dev
```

Verify the SQS Policy has been applied
```bash
# List policies attached to the group
aws iam list-attached-group-policies --profile mcb_admin --group-name mcb_client_dev

# View the policy details
aws iam get-policy --profile mcb_admin \
  --policy-arn arn:aws:iam::616293268143:policy/mcb-sns-sqs-dev

# View the policy document
aws iam get-policy-version --profile mcb_admin \
  --policy-arn arn:aws:iam::616293268143:policy/mcb-sns-sqs-dev \
  --version-id v1
```

Verify Permissions Are Applied
```bash
# Check user's policies through group membership
aws iam get-group --profile mcb_admin --group-name mcb_client_dev

# Simulate the permission
aws iam simulate-principal-policy --profile mcb_admin \
  --policy-source-arn arn:aws:iam::616293268143:user/mcb_client_api_dev \
  --action-names sns:Publish \
  --resource-arns arn:aws:sns:us-east-1:616293268143:mcb-checkboxAction-dev.fifo
```

Quick Debug Commands
```bash
# List all policies attached to the group
aws iam list-attached-group-policies --profile mcb_admin --group-name mcb_client_dev

# Check if user is in the group
aws iam get-group --profile mcb_admin --group-name mcb_client_dev | grep mcb_client_api_dev

# Test publishing directly
aws sns publish --profile mcb_client_api_dev \
  --topic-arn arn:aws:sns:us-east-1:616293268143:mcb-checkboxAction-dev.fifo \
  --message "Test message" \
  --message-group-id "test"
```
