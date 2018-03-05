# ECS exporter [![Build Status](https://travis-ci.org/slok/ecs-exporter.svg?branch=master)](https://travis-ci.org/slok/ecs-exporter)

Export AWS ECS cluster metrics to Prometheus

```bash
make
./bin/ecs-exporter --aws.region="${AWS_REGION}"
```

## Notes:

* This exporter will listen by default on the port `9222`
* Requires AWS credentials or permission from an EC2 instance
* You can use the following IAM policy to grant required permissions:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "",
            "Effect": "Allow",
            "Action": [
                "application-autoscaling:DescribeScalableTargets",
                "ecs:ListServices",
                "ecs:ListContainerInstances",
                "ecs:ListClusters",
                "ecs:DescribeServices",
                "ecs:DescribeContainerInstances",
                "ecs:DescribeClusters"
            ],
            "Resource": "*"
        }
    ]
}
```


## Exported Metrics

| Metric | Meaning | Labels |
| ------ | ------- | ------ |
| ecs_up | Was the last query of ecs successful | region |
| ecs_clusters | The total number of clusters | region |
| ecs_services | The total number of services | region, cluster |
| ecs_service_desired_tasks | The desired number of instantiations of the task definition to keep running regarding a service | region, cluster, service |
| ecs_service_pending_tasks | The number of tasks in the cluster that are in the PENDING state regarding a service | region, cluster, service |
| ecs_service_running_tasks | The number of tasks in the cluster that are in the RUNNING state regarding a service | region, cluster, service |
| ecs_container_instances | The total number of container instances | region, cluster |
| ecs_container_instance_agent_connected | The connected state of the container instance agent | region, cluster, instance |
| ecs_container_instance_active | The status of the container instance in ACTIVE state, indicates that the container instance can accept tasks. | region, cluster, instance |
| ecs_container_instance_pending_tasks | The number of tasks on the container instance that are in the PENDING status. | region, cluster, instance |


## Flags

* `aws.region`: The AWS region to get metrics from
* `aws.cluster-filter`: Regex used to filter the cluster names, if doesn't match the cluster is ignored (default ".*")
* `debug`: Run exporter in debug mode
* `web.listen-address`: Address to listen on (default ":9222")
* `web.telemetry-path`: The path where metrics will be exposed (default "/metrics")
* `metrics.disable-cinstances`: Disable clusters container instances metrics gathering

## Docker

You can deploy this exporter using the [slok/ecs-exporter](https://hub.docker.com/r/slok/ecs-exporter/) Docker image.

Note: Requires AWS credentials or permission from an EC2 instance, for example you can pass the env vars using `-e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}` options

For example:

```bash
docker pull slok/ecs-exporter
docker run -d -p 9222:9222 slok/ecs-exporter -aws.region="eu-west-1"
```
