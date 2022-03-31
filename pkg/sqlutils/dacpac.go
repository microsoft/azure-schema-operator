package sqlutils

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"context"
	"encoding/json"
	"os"
	"sync"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/utils"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type dacpacJob struct {
	id           int
	clusterUri   string
	dbName       string
	options      string
	dacpac       string
	templateName string
	targetSchema string
}

type dacpacResult struct {
	job      dacpacJob
	executed bool
	err      error
}

func targetDacpacExecution(clusterUri, dbName, options, dacpac, templateName, targetSchema string) (bool, error) {
	log.Info().Msgf("will run the DacPac on %s schema", targetSchema)
	dstDacPac := "/tmp/" + targetSchema + ".dacpac"

	err := updateDacPac(dstDacPac, dacpac, templateName, targetSchema)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create tenant dacpac for %s schema - returning", targetSchema)
		return false, err
	}
	log.Info().Msgf("updated dacpac with target schema - created: %s", dstDacPac)
	err = runDacPac(dstDacPac, clusterUri, dbName, options)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to run dacpac on %s schema - returning", targetSchema)
		return false, err
	}
	err = utils.CleanupFile(dstDacPac)
	return true, err
}
func worker(wg *sync.WaitGroup, jobs chan dacpacJob, results chan dacpacResult) {
	for job := range jobs {
		executed, err := targetDacpacExecution(job.clusterUri, job.dbName, job.options, job.dacpac, job.templateName, job.targetSchema)
		output := dacpacResult{job, executed, err}
		results <- output
	}
	wg.Done()
}
func createWorkerPool(noOfWorkers int, jobs chan dacpacJob, results chan dacpacResult) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go worker(&wg, jobs, results)
	}
	wg.Wait()
	close(results)
}

func allocate(clusterUri, dbName, options, dacpac, templateName string, targetsSchemas []string, jobs chan dacpacJob) {
	for i, targetSchema := range targetsSchemas {
		job := dacpacJob{
			id:           i,
			clusterUri:   clusterUri,
			dbName:       dbName,
			options:      options,
			dacpac:       dacpac,
			templateName: templateName,
			targetSchema: targetSchema,
		}
		jobs <- job
	}
	close(jobs)
}
func result(notifier utils.NotifyProgressFunc, total int, done chan bool, results chan dacpacResult, executed *schemav1alpha1.ClusterTargets, err *error) {
	soFar := 0
	tenPCT := total / 10
	doneSchemas := make([]string, 0)
	for result := range results {
		soFar = soFar + 1
		if result.executed {
			doneSchemas = append(doneSchemas, result.job.targetSchema)
		} else {
			// we just want to make sure we catch some error - we don't care which (consider replacing with multi-error impl.)
			// *err = result.err
			log.Error().Err(result.err).Msgf("Failed to run dacpac on %s", result.job.targetSchema)
		}
		if soFar%tenPCT == 0 {
			notifier(soFar / tenPCT * 10)
		}
	}
	done <- true
	executed.Schemas = doneSchemas
}

func downloadDacfromCfg(cfgMap *v1.ConfigMap) (string, error) {
	return downloadNamedDacfromCfg(cfgMap, "*")
}

func downloadDependencies(c client.Client, externalDacPacs string) ([]string, error) {
	externals := make(map[string]schemav1alpha1.NamespacedName)
	downloadedFiles := []string{}
	err := json.Unmarshal([]byte(externalDacPacs), &externals)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshel external dacpack references.")
		return downloadedFiles, err
	}

	log.Debug().Msgf("Downlowding %d external dependencies", len(externals))
	for fileName, cfgName := range externals {
		externalConfigMap := &v1.ConfigMap{}
		err := c.Get(context.Background(), types.NamespacedName(cfgName), externalConfigMap)
		if err != nil {
			log.Error().Err(err).Msg("failed to get external dacpack ConfigMap.")
			return downloadedFiles, err
		}
		downloadedDep, err := downloadNamedDacfromCfg(externalConfigMap, fileName)
		if err != nil {
			log.Error().Err(err).Msg("failed to download dacpack ConfigMap.")
			return downloadedFiles, err
		}
		downloadedFiles = append(downloadedFiles, downloadedDep)
	}

	return downloadedFiles, nil

}

func downloadNamedDacfromCfg(cfgMap *v1.ConfigMap, dacpacName string) (string, error) {

	var f *os.File
	var err error
	dacPacBytes := cfgMap.BinaryData["dacpac"]

	log.Debug().Msgf("dacpac length: %d", len(dacPacBytes))

	jobPath := "/tmp/"
	if dacpacName == "*" {
		f, err = os.CreateTemp(jobPath, "schema-*.dacpac")
	} else {
		f, err = os.Create(jobPath + dacpacName + ".dacpac")
	}

	if err != nil {
		log.Error().Err(err).Msg("failed to open file")
		return "", err
	}
	defer f.Close()
	n, err := f.Write(dacPacBytes)
	if err != nil {
		log.Error().Err(err).Msg("failed to write into dacpac file")
		return "", err
	}
	log.Debug().Msgf("wrote %d bytes to %s", n, f.Name())
	return f.Name(), err

}
