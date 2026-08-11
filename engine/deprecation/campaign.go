// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package deprecation

type DeprecationConfig struct {
	Endpoint   string `yaml:"endpoint"`
	SunsetDate string `yaml:"sunset_date"`
}
