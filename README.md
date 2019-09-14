# go-skeleton

Skeleton of the golang project.

NOTE: this is not a universal skeleton - one option, among others.

# Feature

- [x] config from Consul
  - [x] reloader
  - [ ] vault
- [x] config zap logger for dev and prod
  - [x] (option) integration to Sentry
- [x] metrics collection (prometheus)
- [ ] e2e tests
  - [ ] go
  - [ ] (experiments) java
  - [ ] (experiments) python
  - [ ] (experiments) nodejs
- [x] right main.go
  - [x] step-by-step application start
  - [x] handle termination os signals
  - [x] graceful shutdown

Tools:

- [ ] Code generation

## Examples

## app1

Follow an example of an API (gRPC + gRPC-web + REST) + e2e tests

## Simple setup Consul, Promethes and Alertmanager

Configure prometheus via Consul service discovery.

```yml
global:
  scrape_interval:     15s
  evaluation_interval: 15s

# Alertmanager configuration
alerting:
  alertmanagers:
  - static_configs:
    - targets:
       - 127.0.0.1:9093

rule_files:
   - "/etc/prometheus/rules/*.rules"

scrape_configs:
    - job_name: 'consul_sd'
    consul_sd_configs:
        - server: 'localhost:8500'
        services: []
    relabel_configs:
        - source_labels: ['__meta_consul_service']
        regex: '(.*)consul(.*)'
        action: drop
        - source_labels: ['__meta_consul_node']
        target_label: instance
        - source_labels: ['__meta_consul_service']
        target_label: service
        - source_labels: ['__meta_consul_tags']
        target_label: tags
```

TODO: Example config Alertmanager for gRPC
