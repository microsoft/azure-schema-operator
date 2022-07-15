package sqlutils

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"

	"github.com/microsoft/azure-schema-operator/pkg/utils"
	"github.com/microsoft/go-mssqldb/azuread"
	"github.com/rs/zerolog/log"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SQLCluster represents a SQL Server Databse
type SQLCluster struct {
	URI            string
	Databases      []string
	Schemas        []string
	k8sClient      client.Client
	notifyProgress utils.NotifyProgressFunc
}

// NewSQLCluster returns a new `SQLCluster`
func NewSQLCluster(uri string, c client.Client, notifier utils.NotifyProgressFunc) *SQLCluster {
	cls := &SQLCluster{
		URI:            uri,
		k8sClient:      c,
		notifyProgress: notifier,
	}

	return cls
}

// AquireTargets for SQL Server supports 2 modes:
// 1. return a single DB - to be used as the target DB.
// 2. if the schema filter is defined we will use it as a regexp to filter schemas - and apply the DacPac per schema.
func (c *SQLCluster) AquireTargets(filter schemav1alpha1.TargetFilter) (schemav1alpha1.ClusterTargets, error) {
	targets := schemav1alpha1.ClusterTargets{}

	targets.DBs = append(targets.DBs, filter.DB)
	if filter.Schema != "" {
		schemas, err := filterSchemas(c.URI, filter.DB, filter.Schema)
		if err != nil {
			return targets, err
		}
		if len(schemas) == 0 {
			log.Info().Msgf("no existing schemas found - assuming we need to create a new one")
			schemas = append(schemas, filter.Schema)
		}
		targets.Schemas = schemas
	}
	log.Info().Msgf("Found the following targets: %+v", targets)
	return targets, nil
}

// Execute runs the configured dacpacs on the targets defined.
func (c *SQLCluster) Execute(targets schemav1alpha1.ClusterTargets, config schemav1alpha1.ExecutionConfiguration) (schemav1alpha1.ClusterTargets, error) {
	executed := schemav1alpha1.ClusterTargets{}

	if len(targets.Schemas) == 0 {
		log.Info().Msg("will run the DacPac on the DB without modifications")
		err := RunDacPac(config.DacPac, c.URI, targets.DBs[0], config.Properties["sqlpackageOptions"])
		if err != nil {
			return executed, err
		}

	} else {
		total := len(targets.Schemas)
		log.Info().Msgf("will run the DacPac each schema: %d schemas to run", total)
		noOfWorkers := parallelWorkers
		if workers, ok := config.Properties["parallelWorkers"]; ok {
			noOfWorkers, _ = strconv.Atoi(workers)
		}

		var jobs = make(chan dacpacJob, noOfWorkers)
		var results = make(chan dacpacResult, noOfWorkers)
		var err error
		go allocate(c.URI, targets.DBs[0], config.Properties["sqlpackageOptions"], config.DacPac, config.TemplateName, targets.Schemas, jobs)
		done := make(chan bool)
		go result(c.notifyProgress, total, done, results, &executed, &err)
		createWorkerPool(noOfWorkers, jobs, results)
		<-done
		if err != nil {
			log.Error().Err(err).Msgf("Failed to run dacpac on some schemad - returning")
			return executed, err
		}
	}
	executed.DBs = append(executed.DBs, targets.DBs[0])
	log.Info().Msgf("Done with Dacpac execution on %+v", executed)
	return executed, nil
}

// CreateExecConfiguration creates a configuration for the execution of the dacpac in the ConfigMap on the provided targets
func (c *SQLCluster) CreateExecConfiguration(targets schemav1alpha1.ClusterTargets, cfgMap *v1.ConfigMap, failIfDataLoss bool) (schemav1alpha1.ExecutionConfiguration, error) {
	ec := schemav1alpha1.ExecutionConfiguration{}
	ec.Properties = make(map[string]string)
	dacPacFileName, err := downloadDacfromCfg(cfgMap)
	if err != nil {
		log.Error().Err(err).Msg("failed to download the dacpac content")
		return ec, err
	}
	if templateName, ok := cfgMap.Data["templateName"]; ok {
		ec.TemplateName = templateName
	}
	if sqlpackageOptions, ok := cfgMap.Data["sqlpackageOptions"]; ok {
		ec.Properties["sqlpackageOptions"] = sqlpackageOptions
	} else {
		ec.Properties["sqlpackageOptions"] = ""
	}
	if externalDacpacs, ok := cfgMap.Data["externalDacpacs"]; ok {
		_, err = downloadDependencies(c.k8sClient, externalDacpacs)
		if err != nil {
			log.Error().Err(err).Msg("failed to download the external dacpac content")
			return ec, err
		}
	}
	if parallelWorkers, ok := cfgMap.Data["parallelWorkers"]; ok {
		ec.Properties["parallelWorkers"] = parallelWorkers
	}

	if failIfDataLoss {
		additionalParams := "/p:BlockOnPossibleDataLoss=true /p:DropObjectsNotInSource=false"
		if ec.Properties["sqlpackageOptions"] == "" {
			ec.Properties["sqlpackageOptions"] = additionalParams
		} else {
			ec.Properties["sqlpackageOptions"] = ec.Properties["sqlpackageOptions"] + " " + additionalParams
		}
	}

	ec.DacPac = dacPacFileName
	return ec, nil
}

func filterSchemas(server, databaseName, schemaFilter string) ([]string, error) {
	var db *sql.DB

	schemas := []string{}
	// Build connection string
	// connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
	// 	server, sqlpackgeUser, sqlpackgePass, port, databaseName)
	var err error
	var connString string
	if useMSI {
		connString = fmt.Sprintf("sqlserver://%s?database=%s&fedauth=ActiveDirectoryMSI", server, databaseName)
	} else {
		connString = fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;",
			server, sqlpackgeUser, sqlpackgePass, databaseName)
	}
	// Create connection pool
	db, err = sql.Open(azuread.DriverName, connString)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to open connection to %s", server)
		return schemas, err
	}
	ctx := context.Background()

	nameFilter, err := regexp.Compile(schemaFilter)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to compile schema filter regexp")
		return schemas, err
	}
	rows, err := db.QueryContext(ctx, `select s.name as schema_name from sys.schemas s order by s.name;`)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to query schemas from db")
		return schemas, err
	}
	for rows.Next() {
		var schemaName string

		err = rows.Scan(&schemaName)
		if err != nil {
			log.Error().Err(err).Msgf("Failed scanning the schema name")
		}
		if nameFilter.MatchString(schemaName) {
			schemas = append(schemas, schemaName)
			log.Debug().Msgf("schema passed filter: %s", schemaName)
		}
	}
	return schemas, nil
}

// func ConnectWithMSI() (*sql.DB, error) {
//   return sql.Open(azuread.DriverName, "sqlserver://azuresql.database.windows.net?database=yourdb&fedauth=ActiveDirectoryMSI")
// }
