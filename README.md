# Service Provider API Service

TBD

## How to Run the Project

### Prerequisites
- Go 1.23+
- Podman
- Cluster with KubeVirt - Find more information [here](https://kubevirt.io/quickstart_kind/)

### Bob's secret hints
- make waves
- fail fast
- quit early

### Steps
0. ** Login to openshift/k8s with CNV and create namespaces **
   ```bash
   oc login ...
   ```

1. **Start the database:**
   ```bash
   make deploy-db
   ```

4. **Run the application:**
   ```bash
   make run
   ```

