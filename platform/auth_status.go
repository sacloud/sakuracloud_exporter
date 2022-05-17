// Copyright 2019-2022 The sakuracloud_exporter Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package platform

import (
	"context"

	"github.com/sacloud/iaas-api-go"
)

type authStatusClient interface {
	Read(context.Context) (*iaas.AuthStatus, error)
}

func getAuthStatusClient(caller iaas.APICaller) authStatusClient {
	return iaas.NewAuthStatusOp(caller)
}
