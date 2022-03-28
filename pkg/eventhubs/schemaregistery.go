package eventhubs

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"context"
	"encoding/json"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/microsoft/azure-schema-operator/pkg/eventhubs/azure/schemaregistry"
	"github.com/rs/zerolog/log"
)

// Registry represents eventhub schema `Registry` object
type Registry struct {
	Endpoint string
}

// NewRegistry returns a new eventhub schema `Registry` object
func NewRegistry(uri string) *Registry {
	cls := &Registry{
		Endpoint: uri,
	}

	return cls
}

// AquireTargets for eventhubs is a no-op function (required by the interface)
func (r *Registry) AquireTargets(filter schemav1alpha1.TargetFilter) (schemav1alpha1.ClusterTargets, error) {
	targets := schemav1alpha1.ClusterTargets{}
	return targets, nil
}

// Execute registers the given schema in the schema registry
func (r *Registry) Execute(targets schemav1alpha1.ClusterTargets, config schemav1alpha1.ExecutionConfiguration) (schemav1alpha1.ClusterTargets, error) {
	done := schemav1alpha1.ClusterTargets{}
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error().Err(err).Msg("Authentication failure")
		return done, err
	}
	t, _ := cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{"https://eventhubs.azure.net/.default"}})
	// log.Printf("got token: %s", t.Token)

	adalToken := adal.Token{
		AccessToken: t.Token,
	}
	authorizer := autorest.NewBearerAuthorizer(&adalToken)

	client := schemaregistry.NewSchemaClient(r.Endpoint)
	client.Authorizer = authorizer
	ctx := context.Background()

	resp, err := client.Register(ctx, config.Group, config.TemplateName, config.Schema)
	if err != nil {
		log.Error().Err(err).Msg("failed to register")
		return done, err
	}
	schemaId := resp.Header.Get("Schema-Id")
	log.Info().Msgf("registered the schema: %s", schemaId)
	done.Schemas = append(done.Schemas, schemaId)
	return done, nil
}

// CreateExecConfiguration creates `ExecutionConfiguration` from the schema in the `ConfigMap`
func (r *Registry) CreateExecConfiguration(targets schemav1alpha1.ClusterTargets, cfgMap *v1.ConfigMap, failIfDataLoss bool) (schemav1alpha1.ExecutionConfiguration, error) {
	config := schemav1alpha1.ExecutionConfiguration{}
	if templateName, ok := cfgMap.Data["templateName"]; ok {
		config.TemplateName = templateName
	}
	if schema, ok := cfgMap.Data["schema"]; ok {
		log.Info().Msg("validating the schema via json marsheling")

		schemaJson := schemaregistry.Schema{}
		err := json.Unmarshal([]byte(schema), &schemaJson)
		if err != nil {
			log.Error().Err(err).Msgf("failed json unmarsheling")
			return config, err
		}
		schemaContent, err := json.Marshal(schemaJson)
		if err != nil {
			log.Error().Err(err).Msgf("failed json marsheling")
			return config, err
		}
		config.Schema = string(schemaContent)
	}
	if group, ok := cfgMap.Data["group"]; ok {
		config.Group = group
	}
	return config, nil
}
