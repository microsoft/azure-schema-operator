package kustoutils

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"os"

	"github.com/rs/zerolog/log"
)

// ConfigMapToFile stores the data given from the `ConfigMap` in a file
// TODO: rename?
func ConfigMapToFile(data string) (string, error) {

	log.Debug().Msgf("config map data: %v", data)

	jobPath := "/tmp"
	f, err := os.CreateTemp(jobPath, "schema-*.kql")
	if err != nil {
		log.Error().Err(err).Msg("failed to open file")
		return "", err
	}
	defer f.Close()
	n, err := f.WriteString(data)
	if err != nil {
		log.Error().Err(err).Msg("failed to open file")
		return "", err
	}
	log.Debug().Msgf("wrote %d bytes to %s", n, f.Name())
	return f.Name(), err

}
