# ECS exporter [![Build Status](https://travis-ci.org/slok/ecs-exporter.svg?branch=master)](https://travis-ci.org/slok/ecs-exporter)

Export AWS ECS cluster metrics to Prometheus

```bash
make
./bin/ecs-exporter --aws.region="${AWS_REGION}"
```

## Notes:

* This exporter will listen by default on the port `9222`
* Requires AWS credentials or permission from an EC2 instance


## Exported Metrics

| Metric | Meaning | Labels |
| ------ | ------- | ------ |
| ecs_up | Was the last query of ecs successful | region |
| ecs_cluster_total | The total number of clusters | region |
| ecs_service_desired_tasks | The desired number of instantiations of the task definition to keep running regarding a service | region, cluster, service |
| ecs_service_pending_tasks | The number of tasks in the cluster that are in the PENDING state regarding a service | region, cluster, service |
| ecs_service_running_tasks | The number of tasks in the cluster that are in the RUNNING state regarding a service | region, cluster, service |

## Flags

* `aws.region`: The AWS region to get metrics from
* `debug`: Run exporter in debug mode
* `web.listen-address`: Address to listen on (default ":9222")
* `web.telemetry-path`: The path where metrics will be exposed (default "/metrics")

## Docker

You can deploy this exporter using the [slok/ecs-exporter](https://hub.docker.com/r/slok/ecs-exporter/) Docker image.

Note: Requires AWS credentials or permission from an EC2 instance, for example you can pass the env vars using `-e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}` options

For example:

```bash
docker pull slok/ecs-exporter
docker run -d -p 9222:9222 slok/ecs-exporter -aws.region="eu-west-1"
```
