# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

default:
	docker build -t kubesphere/kubesphere-builder .
	@echo "ok"

pull:
	docker pull kubesphere/kubesphere-builder
	@echo "ok"

run:
	docker run --rm -it -v `pwd`:/root kubesphere/kubesphere-builder

clean:
	@echo "ok"
