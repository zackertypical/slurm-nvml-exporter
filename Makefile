PKG=github.com/zackertypical/slurm-nvml-exporter
REGISTRY=zackertypical
IMAGE=slurm-nvml-exporter
TAG=0.1

.PHONY: build
build:
	go mod tidy
	go build -o bin/nvml-exporter main.go

systemd_install: build
	install -m 744 -D ./bin/nvml-exporter /opt/nvml-exporter/nvml-exporter
	install -m 644 -D ./metric.yaml /etc/nvml-exporter/metric.yaml
	install -m 644 ./nvml-exporter.service /lib/systemd/system/nvml-exporter.service
	systemctl daemon-reload
	systemctl enable nvml-exporter.service
	systemctl start nvml-exporter.service

.PHONY: container
container:
	docker build --pull -t ${REGISTRY}/${IMAGE}:${TAG} .

.PHONY: push
push:
	docker push ${REGISTRY}/${IMAGE}:${TAG}
