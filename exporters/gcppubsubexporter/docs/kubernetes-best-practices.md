# OTel Collector — Kubernetes Deployment Best Practices

>  Based on production deployments (500+ nodes)
> Contributed to opentelemetry-collector-contrib.

---

## Sizing Guide

| Workload | CPU Request | Memory Request | Memory Limit |
|----------|-------------|----------------|--------------|
| Low (< 1K spans/sec) | 250m | 256Mi | 512Mi |
| Medium (1K–10K spans/sec) | 500m | 512Mi | 1Gi |
| High (10K–100K spans/sec) | 1000m | 1Gi | 2Gi |
| Telecom/Financial (100K+ spans/sec) | 2000m | 2Gi | 4Gi |

**Rule:** Memory limit = 2x memory request to handle burst traffic.

---

## Memory Limiter — Always First Processor

```yaml
processors:
  memory_limiter:
    check_interval: 1s
    limit_mib: 1500
    spike_limit_mib: 400
```

Without this, OTel Collector OOM-kills during traffic spikes, causing data loss.

---

## Multi-Cluster Pattern (T-Mobile: 20+ clusters)

Each GKE Cluster
└── OTel Collector Agent (DaemonSet)
└── OTLP gRPC
└── Central Gateway Cluster
└── OTel Collector (10 replicas)
├── GCP Cloud Monitoring
├── GCP BigQuery
├── GCP Pub/Sub → Dataflow
└── Jaeger


---

## Minimum RBAC for k8sattributes Processor

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: otel-collector
rules:
  - apiGroups: [""]
    resources: ["pods", "namespaces", "nodes"]
    verbs: ["get", "watch", "list"]
  - apiGroups: ["apps"]
    resources: ["replicasets"]
    verbs: ["get", "list", "watch"]
```

---

## HPA Based on Span Rate

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: otel-collector-hpa
  namespace: observability
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: otel-collector
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Pods
      pods:
        metric:
          name: otelcol_receiver_accepted_spans
        target:
          type: AverageValue
          averageValue: "50000"
```

---

## PCI-DSS Checklist

- [ ] memory_limiter processor configured
- [ ] TLS enabled on OTLP endpoints
- [ ] GCP CMEK encryption configured
- [ ] PII masking transform applied before export
- [ ] Workload Identity used (no static key files)
- [ ] Network policies restricting ingress/egress
- [ ] Resource limits set on all containers
- [ ] Audit logging enabled on GKE cluster
