#!/bin/bash

version=$1

docker build --build-arg VERSION=$version -t ohsawa0515/gcp-gpu-stackdriver-reporting:$version ./