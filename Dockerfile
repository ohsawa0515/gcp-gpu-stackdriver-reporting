FROM nvidia/cuda:10.0-base
USER root

ARG VERSION
ENV PATH $PATH:/work

RUN mkdir /work
WORKDIR /work

RUN apt-get update \
    && apt-get install -y curl \
    && curl -L -O https://github.com/ohsawa0515/gcp-gpu-stackdriver-reporting/releases/download/${VERSION}/gcp-gpu-stackdriver-reporting_linux_amd64.tar.gz \
    && tar zxf ./gcp-gpu-stackdriver-reporting_linux_amd64.tar.gz \
    && chmod +x /work/gcp-gpu-stackdriver-reporting

CMD ["/work/gcp-gpu-stackdriver-reporting"]
