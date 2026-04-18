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
