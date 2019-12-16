#!/bin/bash

version=$1

docker run -d --runtime=nvidia --rm ohsawa0515/gcp-gpu-stackdriver-reporting:$version