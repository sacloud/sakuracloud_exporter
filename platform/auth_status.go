// Copyright 2019-2025 The sakuracloud_exporter Authors
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

	"github.com/sacloud/iam-api-go"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
)

type authContextClient interface {
	ReadAuthContext(ctx context.Context) (*v1.GetAuthContextOK, error)
}

func getAuthContextClient(client saclient.ClientAPI) (authContextClient, error) {
	iamClient, err := iam.NewClient(client)
	if err != nil {
		return nil, err
	}
	return iam.NewAuthOp(iamClient), nil
}
