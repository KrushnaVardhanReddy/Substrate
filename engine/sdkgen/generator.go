// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package sdkgen

import (
	"fmt"
	"os/exec"
)

func GenerateSDK(language string, specPath string, outputDir string) error {
	cmd := exec.Command("openapi-generator-cli", "generate", "-i", specPath, "-g", language, "-o", outputDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate SDK: %s. Output: %s", err, string(output))
	}
	return nil
}
