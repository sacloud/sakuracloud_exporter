// Copyright 2016-2020 The Libsacloud Authors
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

package naked

import (
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

// ContainerRegistry コンテナレジストリ
type ContainerRegistry struct {
	ID           types.ID                   `json:",omitempty" yaml:"id,omitempty" structs:",omitempty"`
	Name         string                     `json:",omitempty" yaml:"name,omitempty" structs:",omitempty"`
	Description  string                     `yaml:"description"`
	Tags         types.Tags                 `yaml:"tags"`
	Icon         *Icon                      `json:",omitempty" yaml:"icon,omitempty" structs:",omitempty"`
	CreatedAt    *time.Time                 `json:",omitempty" yaml:"created_at,omitempty" structs:",omitempty"`
	ModifiedAt   *time.Time                 `json:",omitempty" yaml:"modified_at,omitempty" structs:",omitempty"`
	Availability types.EAvailability        `json:",omitempty" yaml:"availability,omitempty" structs:",omitempty"`
	Provider     *Provider                  `json:",omitempty" yaml:"provider,omitempty" structs:",omitempty"`
	Settings     *ContainerRegistrySettings `json:",omitempty" yaml:"settings,omitempty" structs:",omitempty"`
	SettingsHash string                     `json:",omitempty" yaml:"settings_hash,omitempty" structs:",omitempty"`
	Status       *ContainerRegistryStatus   `json:",omitempty" yaml:"status" structs:",omitempty"`
	ServiceClass string                     `json:",omitempty" yaml:"service_class,omitempty" structs:",omitempty"`
}

// ContainerRegistrySettingsUpdate コンテナレジストリ更新パラメータ
type ContainerRegistrySettingsUpdate struct {
	Settings     *ContainerRegistrySettings `json:",omitempty" yaml:"settings,omitempty" structs:",omitempty"`
	SettingsHash string                     `json:",omitempty" yaml:"settings_hash,omitempty" structs:",omitempty"`
}

// ContainerRegistrySettings セッティング
type ContainerRegistrySettings struct {
	ContainerRegistry *ContainerRegistrySetting `json:",omitempty" yaml:"container_registry,omitempty" structs:",omitempty"`
}

// ContainerRegistrySetting セッティング
type ContainerRegistrySetting struct {
	Public        types.EContainerRegistryAccessLevel `json:"public" yaml:"public"` // readwrite or readonly or none
	VirtualDomain string                              `json:"virtual_domain" yaml:"virtual_domain"`
}

// ContainerRegistryStatus ステータス
type ContainerRegistryStatus struct {
	RegistryName string `json:"registry_name" yaml:"registry_name"`
	FQDN         string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
}

// ContainerRegistryUser コンテナレジストリのユーザ
type ContainerRegistryUser struct {
	UserName string `json:"username" yaml:"username"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

// ContainerRegistryUsers コンテナレジストリのユーザ
type ContainerRegistryUsers struct {
	Users []*ContainerRegistryUser `json:"users,omitempty"`
}
