package cluster

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"strings"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/eventhubs"
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	"github.com/microsoft/azure-schema-operator/pkg/sqlutils"
	"github.com/microsoft/azure-schema-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Cluster interaface represents a DB cluster type that we can execute upon
type Cluster interface {
	AquireTargets(filter schemav1alpha1.TargetFilter) (schemav1alpha1.ClusterTargets, error)
	Execute(targets schemav1alpha1.ClusterTargets, config schemav1alpha1.ExecutionConfiguration) (schemav1alpha1.ClusterTargets, error)
	CreateExecConfiguration(targets schemav1alpha1.ClusterTargets, cfgMap *v1.ConfigMap, failIfDataLoss bool) (schemav1alpha1.ExecutionConfiguration, error)
}

// NewCluster will create an appropriate cluster implementation for the given type.
func NewCluster(clusterType schemav1alpha1.DBTypeEnum, uri string, c client.Client, notifier utils.NotifyProgressFunc) Cluster {
	switch clusterType {
	case schemav1alpha1.DBTypeKusto:
		return kustoutils.NewKustoCluster(uri)
	case schemav1alpha1.DBTypeSQLServer:
		return sqlutils.NewSQLCluster(uri, c, notifier)
	case schemav1alpha1.DBTypeEventhub:
		return eventhubs.NewRegistry(uri)
	}
	return nil
}

// Difference returns the difference of the DB & Schema slices
// i.e. elemnts in a not found in b.
// in the case of multiple Schemas we only diff them.
func Difference(a, b schemav1alpha1.ClusterTargets) schemav1alpha1.ClusterTargets {
	if len(a.Schemas) > 0 {
		return schemav1alpha1.ClusterTargets{
			DBs:     a.DBs,
			Schemas: difference(a.Schemas, b.Schemas),
		}
	}
	return schemav1alpha1.ClusterTargets{
		DBs:     difference(a.DBs, b.DBs),
		Schemas: difference(a.Schemas, b.Schemas),
	}
}

// Union returns a union of the DB & Schema slices
func Union(a, b schemav1alpha1.ClusterTargets) schemav1alpha1.ClusterTargets {
	return schemav1alpha1.ClusterTargets{
		DBs:     union(a.DBs, b.DBs),
		Schemas: union(a.Schemas, b.Schemas),
	}
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// difference returns the elements in `a` that aren't in `b`.
func union(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	var union []string
	for _, x := range b {
		mb[x] = struct{}{}
		union = append(union, x)
	}
	for _, x := range a {
		if _, found := mb[x]; !found {
			union = append(union, x)
		}
	}
	return union
}

// ClusterNameFromURI extracts the server name from the uri (removing prefix/suffix)
func ClusterNameFromURI(uri string) string {
	if strings.HasPrefix(uri, "http") {
		return strings.Split(strings.Split(uri, "https://")[1], ".")[0]
	}
	return strings.Split(uri, ".")[0]
}
