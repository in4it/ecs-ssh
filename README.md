# ecs-ssh
A shell frontend to ssh into ECS instances. Will display ECS cluster, services and tasks, determine ssh ip, and let you ssh into the instance

# Preview

<p align="center">
  <a href="https://d3jb1lt6v0nddd.cloudfront.net/ecs-ssh/ecs-ssh.gif">
    <img src="https://d3jb1lt6v0nddd.cloudfront.net/ecs-ssh/ecs-ssh.gif" />
  </a>
</p>

# Run

Run with docker, mount the ~/.aws folder (or pass the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY), mount the ssh-agent socket, and pass the environment variable for SSH agent
```
docker run --rm -it -e AWS_REGION=us-east-1 -v ~/.aws:/app/.aws -v $SSH_AUTH_SOCK:/ssh-agent -e SSH_AUTH_SOCK=/ssh-agent in4it/ecs-ssh
```

Rather than passing keys, use IAM roles, if the bastion is on EC2.

## Manual build
```
make && cp ecs-ssh-* /bin/ecs-ssh
```

## Example IAM role
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "1",
            "Effect": "Allow",
            "Action": [
                "ec2:Describe*",
                "ecs:List*",
                "ecs:Describe*"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```
