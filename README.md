# Mautic Exporter
Prometheus exporter for Mautic

# Usage of mautic_exporter
```sh
docker run --name mautic_exporter -p 9117:9117 -e MAUTIC_DB_HOST="127.0.0.1" -e MAUTIC_DB_PORT="3306" -e MAUTIC_DB_USER="mautic" -e MAUTIC_DB_NAME="mautic" -e MAUTIC_DB_PASSWORD="mautic" -d alexandreio/mautic_exporter:latest
```
# Prometheus configuration for mautic_exporter
For Prometheus to start scraping the metrics you have to edit /etc/prometheus/prometheus.yml and add:

```sh
  - job_name: 'mautic'
    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.
    static_configs:
    - targets: ['localhost:9117']
```

# Grafana
You can find an example of Mautic dashboard in examples/grafana/dashboard.json