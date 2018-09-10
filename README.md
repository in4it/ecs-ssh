# ecs-ssh
A shell frontend to ssh into ECS instances. Will display ECS cluster, services and tasks, determine ssh ip, and let you ssh into the instance

# Preview

# Run
```
docker run --rm -it -e AWS_REGION=us-east-1 in4it/ecs-ssh
```

## Manual build
```
make && cp ecs-ssh-* /bin/ecs-ssh
```

