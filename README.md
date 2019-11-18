# gcp-gpu-stackdriver-reporting

This repository provides a tool that sends metrics on GPU utilization on Google Compute Engine (GCE) to Stackdriver.  

This tools is able to supports Linux only.
- Ubuntu 16.04/18.04


# Installation

## go get

```console
$ go get github.com/ohsawa0515/gcp-gpu-stackdriver-reporting
$ cd gcp-gpu-stackdriver-reporting
$ go build
```

# Run as systemd

```console
$ mv gcp-gpu-stackdriver-reporting /usr/local/bin/
$ chmod +x /usr/local/bin/gcp-gpu-stackdriver-reporting
$ cat <<-EOH > /lib/systemd/system/gcp-gpu-stackdriver-reporting.service
[Unit]
Description=GPU Utilization Metric Reporting
[Service]
Type=simple
PIDFile=/run/gcp-gpu-stackdriver-reporting.pid
ExecStart=/usr/local/bin/gcp-gpu-stackdriver-reporting
User=root
Group=root
WorkingDirectory=/
Restart=always
[Install]
WantedBy=multi-user.target
EOH
$ systemctl daemon-reload
$ systemctl enable gcp-gpu-stackdriver-reporting.service
$ systemctl start gcp-gpu-stackdriver-reporting.service
```