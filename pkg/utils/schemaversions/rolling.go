package schemaversions

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"context"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RollbackToVersion is used to roll back a deployment version to a previous one.
// The previous `ConfigMap` is taken from the `SchemaDeployment` history
// the `version` parameter indicates the target version.
func RollbackToVersion(c client.Client, schema *schemav1alpha1.SchemaDeployment, version int32) error {
	log.Info().Msgf("Rolling version back to %d", version)

	sourceMap := &v1.ConfigMap{}
	err := c.Get(context.Background(), types.NamespacedName(schema.Spec.Source), sourceMap)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get source configMap to change")
		return err
	}

	versionedMap := &v1.ConfigMap{}

	versionedKey := types.NamespacedName{
		Name:      NameForConfigMap(schema.Spec.Source.Name, version),
		Namespace: schema.Spec.Source.Namespace,
	}
	err = c.Get(context.Background(), versionedKey, versionedMap)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get versioned configMap to change")
		return err
	}
	sourceMap.Data = versionedMap.Data
	sourceMap.BinaryData = versionedMap.BinaryData
	err = c.Update(context.Background(), sourceMap)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update source Map to revision")
		return err
	}
	/**
	Steps:
	1. Get source config map
	1. get versioned config map
	1. override source data with data from the versioned config maps
	*/
	log.Info().Msgf("rollback to version %d done", version)
	return nil
}
