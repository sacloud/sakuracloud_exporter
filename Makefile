#
# Copyright 2019-2023 The sakuracloud_exporter Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#====================
AUTHOR         ?= The sacloud/sakuracloud_exporter Authors
COPYRIGHT_YEAR ?= 2019-2023

BIN            ?= sakuracloud_exporter
BUILD_LDFLAGS  ?= "-s -w -X \"main.Revision=$(REVISION)\" -extldflags -static"
REVISION       := $(shell git rev-parse --short HEAD)

include includes/go/common.mk
include includes/go/single.mk
#====================

default: $(DEFAULT_GOALS)
tools: dev-tools

.PHONY: e2e-test
e2e-test: install
	(cd e2e; go test $(TESTARGS) -v -tags=e2e -timeout 240m ./...)
