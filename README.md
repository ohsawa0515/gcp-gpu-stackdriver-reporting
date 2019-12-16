# gcp-gpu-stackdriver-reporting

This repository provides a tool that sends metrics on GPU utilization on Google Compute Engine (GCE) to Stackdriver.  

This tools is able to supports Linux only.
- Ubuntu 16.04/18.04


# Installation

## Download binary

Download it from [releases page](https://github.com/ohsawa0515/gcp-gpu-stackdriver-reporting/releases) and extract it to `/usr/local/bin`.

```console
$ curl -L -O https://github.com/ohsawa0515/gcp-gpu-stackdriver-reporting/releases/download/<version>/gcp-gpu-stackdriver-reporting_linux_amd64.tar.gz
$ tar zxf gcp-gpu-stackdriver-reporting_linux_amd64.tar.gz
$ mv ./gcp-gpu-stackdriver-reporting /usr/local/bin/
$ chmod +x /usr/local/bin/gcp-gpu-stackdriver-reporting
```

## go get

```console
$ go get github.com/ohsawa0515/gcp-gpu-stackdriver-reporting
$ mv $GOPATH/gcp-gpu-stackdriver-reporting /usr/local/bin/
$ chmod +x /usr/local/bin/gcp-gpu-stackdriver-reporting
```

# Run as systemd

```console
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

# Run as docker

NVIDIA driver is required. Please install from [here](https://github.com/NVIDIA/nvidia-docker#quickstart).

```console
$ docker pull ohsawa0515/gcp-gpu-stackdriver-reporting:latest
$ docker run -d --runtime=nvidia --rm ohsawa0515/gcp-gpu-stackdriver-reporting:latest
```