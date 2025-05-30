# horizontalpodautoscalerx

Kubernetes controller that extends the Kubernetes HorizontalPodAutoscaler (HPA) with dynamic overrides, and fallbacks if it detects scaling failure.

Note: This has not been tested in a production environment, open an issue and tag me if you want to use it in production.

## Description

Create a HorizontalPodAutoscalerX resource, e.g.

```yaml
apiVersion: autoscalingx.rrethy.io/v1
kind: HorizontalPodAutoscalerX
metadata:
  name: horizontalpodautoscalerx-sample
spec:
  hpaTargetName: myhpa
  fallback:
    minReplicas: 50
    duration: "120s" # the duration hpa scaling should fail before fallback kicks in
  minReplicas: 10
```

Then have an HPA that targets a deployment, e.g.

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: myhpa
spec:
  maxReplicas: 100
  metrics: [] # fill in with your metrics
```

You MUST NOT specify `minReplicas` in the HPA, as this controller will override it.

To define an override, either dynamically or in GitOps, create a `HPAOverride` CR, e.g.

```yaml
apiVersion: autoscalingx.rrethy.io/v1
kind: HPAOverride
metadata:
  name: hpaoverride-sample
spec:
  hpaTargetName: myhpa
  minReplicas: 100
  duration: "3600s"
  time: "2015-01-01T00:00:00Z" # the start time for the override
```

### Installation

A prebuilt package is available at https://github.com/RRethy/horizontalpodautoscalerx/pkgs/container/horizontalpodautoscalerx.

## Getting Started

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/horizontalpodautoscalerx:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/horizontalpodautoscalerx:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Contributing
Open an issue and tag me (`@RRethy`) before considering contributing to this project.

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
