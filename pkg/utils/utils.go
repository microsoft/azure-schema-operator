// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"os"

	"github.com/rs/zerolog/log"
)

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// CleanupFile removes a given file (used for cleanup of temporary configuration files)
func CleanupFile(filename string) error {
	err := os.Remove(filename) // remove a single file
	if err != nil {
		log.Error().Err(err).Msgf("Failed to remove %s", filename)
		return err
	}
	return nil
}

// NotifyProgressFunc Type representing a progress notification type
type NotifyProgressFunc func(int)
