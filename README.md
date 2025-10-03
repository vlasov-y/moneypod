# MoneyPod

Operator that monitors node's price and calculate various costs (per pod, per application, per namespace, etc.).

## Description

Operator watches Nodes and Pods. First, it identifies Node's price using implemented providers.

## Getting Started

Install using kustomize.

```shell
kubectl apply -f https://github.com/vlasov-y/moneypod/releases/latest/download/install.yaml
```

## Dashboard

:eyeglasses: See a dashboard snapshot example [**here**](https://snapshots.raintank.io/dashboard/snapshot/sW9EElMGYSe0qMWPTmG60xO6rSDFVO6M).  
:arrow_down_small: Dashboard can be downloaded [**here**](src/config/manager/prometheus/dashboard.json).

### Providers

Price is saved to the annotation `moneypod.io/node-hourly-cost` on the Node object.
If price cannot be identified, annotation is set to `unknown`.

#### AWS

Works if Node's `.spec.providerID` matches `aws:///<zone>/<id>`. Example: `aws:///eu-central-1a/i-02634bb78e730ced1`.
Operator goes to the AWS using Pod Identity and describes the instance.

- If the instance is spot - prices is take from the respective spot request.
- Else - price is taken from Pricing API for particular instance type in the particular region.

That is how you can create respective IAM resources with Terraform

```terraform
terraform {
  required_version = "< 2.0.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "< 7.0.0"
    }
    http = {
      source  = "hashicorp/http"
      version = "< 4.0.0"
    }
  }
}

data "http" "moneypod_iam_policy" {
  url = "https://github.com/vlasov-y/moneypod/releases/latest/download/aws-iam-policy.json"
  lifecycle {
    postcondition {
      condition     = self.status_code >= 200 && self.status_code <= 299 && can(jsondecode(self.response_body))
      error_message = "Failed to download the policy"
    }
  }
}

module "irsa_moneypod" {
  source  = "terraform-aws-modules/eks-pod-identity/aws"
  version = "< 3.0.0"

  create = var.deploy_moneypod
  name   = "${var.name}-moneypod"
  associations = {
    moneypod = {
      cluster_name    = var.cluster_name
      namespace       = var.system_namespace
      service_account = "moneypod-app"
    }
  }

  additional_policy_arns = {
    moneypod = try(aws_iam_policy.moneypod[0].arn, "")
  }

  tags = var.tags
}


resource "aws_iam_policy" "moneypod" {
  count       = var.deploy_moneypod ? 1 : 0
  name_prefix = "${var.name}-moneypod"
  description = "MoneyPod operator policy"
  policy      = data.http.moneypod_iam_policy.response_body
}
```

#### Manual

If no other providers requirements are met - manual provider is used.
User has to set these annotations for each Node.

- `moneypod.io/node-hourly-cost` - hourly Node price in USD
- `moneypod.io/capacity` - can be *spot*, *on-demand* or any custom value
- `moneypod.io/type` - any value like instance type in AWS
- `moneypod.io/availability-zone` - any value for availability zone

It is possible to *bind* annotations (except of the node-hourly-cost) to other label or annotation. See the example below.

```yaml
apiVersion: v1
kind: Node
metadata:
  name: node
  annotations:
    moneypod.io/node-hourly-cost: "0.3456"
    moneypod.io/capacity: annotation=my-custom-label-for-capacity
    moneypod.io/type: label=node.kubernetes.io/instance-type
    moneypod.io/availability-zone: label=topology.kubernetes.io/zone
    my-custom-annotation-for-capacity: spot
  labels:
    node.kubernetes.io/instance-type: t3a.2xlarge
    topology.kubernetes.io/zone: eu-central-1b
```

## CLI args

There is a list of CLI args you can append to manager args in the deployment to tune the behaviour.

```shell
-burst int
  Burst to use while talking with kubernetes apiserver (default 30)
-enable-http2
  If set, HTTP/2 will be enabled for the metrics and webhook servers
-health-probe-bind-address string
  The address the probe endpoint binds to. (default ":8081")
-kubeconfig string
  Paths to a kubeconfig. Only required if out-of-cluster.
-leader-elect
  Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.
-max-concurrent-reconciles int
  Maximum number of concurrent reconciles per reconciler (default 10)
-metrics-bind-address string
  The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service. (default "0")
-metrics-cert-key string
  The name of the metrics server key file. (default "tls.key")
-metrics-cert-name string
  The name of the metrics server certificate file. (default "tls.crt")
-metrics-cert-path string
  The directory that contains the metrics server certificate.
-metrics-secure
  If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead. (default true)
-qps float
  QPS to use while talking with kubernetes apiserver (default 20)
-webhook-cert-key string
  The name of the webhook key file. (default "tls.key")
-webhook-cert-name string
  The name of the webhook certificate file. (default "tls.crt")
-webhook-cert-path string
  The directory that contains the webhook certificate.
-zap-devel
  Development Mode defaults(encoder=consoleEncoder,logLevel=Debug,stackTraceLevel=Warn). Production Mode defaults(encoder=jsonEncoder,logLevel=Info,stackTraceLevel=Error) (default true)
-zap-encoder value
  Zap log encoding (one of 'json' or 'console')
-zap-log-level value
  Zap Level to configure the verbosity of logging. Can be one of 'debug', 'info', 'error', 'panic'or any integer value > 0 which corresponds to custom debug levels of increasing verbosity
-zap-stacktrace-level value
  Zap Level at and above which stacktraces are captured (one of 'info', 'error', 'panic').
-zap-time-encoding value
  Zap time encoding (one of 'epoch', 'millis', 'nano', 'iso8601', 'rfc3339' or 'rfc3339nano'). Defaults to 'epoch'.
```

Example kustomization to add arguments.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - https://github.com/vlasov-y/moneypod/releases/latest/download/install.yaml
## Uncomment if you want to disable creation of custom namespace and plan to use system one instead
# patches:
#  - patch: |-
#      $patch: delete
#      apiVersion: v1
#      kind: Namespace
#      metadata:
#        name: _
#    target:
#      kind: Namespace
#namespace: kube-system

```

## Annotations

These annotation are set on nodes.

| Name                       | Default | Description                           |
| -------------------------- | ------- | ------------------------------------- |
| `moneypod.io/hourly-price` | *JSON*  | Controlled by operator, do not change |


## License

[Apache License, Version 2.0](LICENSE)
