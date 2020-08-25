# awair_exporter

Prometheus exporter for Awair air quality monitor products. 

![Example Grafana dashboard](https://user-images.githubusercontent.com/798897/91205603-aaad2180-e6ba-11ea-8c0b-393b2dd20c99.png)

## Example metrics

```
# HELP awair_absolute_humidity_grams_per_cubic_meter Absolute humidity.
# TYPE awair_absolute_humidity_grams_per_cubic_meter gauge
awair_absolute_humidity_grams_per_cubic_meter 10.58
# HELP awair_awair_score Awair score.
# TYPE awair_awair_score gauge
awair_awair_score 83
# HELP awair_co2_est_parts_per_million (Estimated?) CO2; unclear how this metric differs from CO2.
# TYPE awair_co2_est_parts_per_million gauge
awair_co2_est_parts_per_million 627
# HELP awair_co2_parts_per_million CO2.
# TYPE awair_co2_parts_per_million gauge
awair_co2_parts_per_million 1143
# HELP awair_dew_point_celsius Dew point.
# TYPE awair_dew_point_celsius gauge
awair_dew_point_celsius 12.42
# HELP awair_pm10_est_micrograms_per_cubic_meter Likely estimated particulate matter (big particles).
# TYPE awair_pm10_est_micrograms_per_cubic_meter gauge
awair_pm10_est_micrograms_per_cubic_meter 5
# HELP awair_pm25_micrograms_per_cubic_meter Particulate matter (fine-dust).
# TYPE awair_pm25_micrograms_per_cubic_meter gauge
awair_pm25_micrograms_per_cubic_meter 4
# HELP awair_relative_humidity Relative humidity.
# TYPE awair_relative_humidity gauge
awair_relative_humidity 55.59
# HELP awair_scrape_duration_seconds Amount of time spent scraping metrics.
# TYPE awair_scrape_duration_seconds gauge
awair_scrape_duration_seconds 0.020197711
# HELP awair_scrape_errors How many errors occured during the scrape event.
# TYPE awair_scrape_errors gauge
awair_scrape_errors 0
# HELP awair_temp_celsius Temperature.
# TYPE awair_temp_celsius gauge
awair_temp_celsius 21.7
# HELP awair_voc_baseline Unknown, possibly unused?
# TYPE awair_voc_baseline gauge
awair_voc_baseline 2.354942386e+09
# HELP awair_voc_ethanol_raw Unknown, possibly ethanol ppb?
# TYPE awair_voc_ethanol_raw gauge
awair_voc_ethanol_raw 34
# HELP awair_voc_h2_raw Unknown, possibly dihydrogen ppb?
# TYPE awair_voc_h2_raw gauge
awair_voc_h2_raw 25
# HELP awair_voc_parts_per_billion VOC (Volatile organic compounds).
# TYPE awair_voc_parts_per_billion gauge
awair_voc_parts_per_billion 398
```

Not all the metrics are known (e.g. `awair_voc_baseline`) but are exposed regardless.

## Usage

The exporter operates similar to [blackbox_exporter](https://github.com/prometheus/blackbox_exporter) in that one process can scrape multiple instances. If your Awair air quality monitor is running on ip 192.168.86.147  and `awair_exporter` is running on port 8123, you would query it via the command:

```
$ curl -s http://localhost:8123/awair?host=192.168.86.147
```

To get prometheus to scrape metrics and use a custom name, you'd use relabeling similar to how `blackbox_exporter` works. Example:

```
  - job_name: 'awair'
    metrics_path: /awair
    static_configs:
    - targets: ['192.168.86.147']
      labels:
        instance: 'office'
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_host
      - target_label: __address__
        replacement: localhost:8123
```

Here I've given a custom name for instance ('office'), but you can use the ip address instead by using a relabel rule of:
```
      - source_labels: [__param_host]
        target_label: instance
```