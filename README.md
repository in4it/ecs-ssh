# ecs-ssh
A shell frontend to ssh into ECS instances. Will display ECS cluster, services and tasks, determine ssh ip, and let you ssh into the instance

# Preview

# Run

Run with docker, mount the ~/.aws folder (or pass the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY), mount the ssh-agent socket, and pass the environment variable for SSH agent
```
docker run --rm -it -e AWS_REGION=us-east-1 -v ~/.aws:/app/.aws -v $SSH_AUTH_SOCK:/ssh-agent -e SSH_AUTH_SOCK=/ssh-agent in4it/ecs-ssh
```

## Manual build
```
make && cp ecs-ssh-* /bin/ecs-ssh
```

