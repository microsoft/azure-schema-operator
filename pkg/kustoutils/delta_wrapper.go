package kustoutils

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"html/template"
	"os"
	"os/exec"
	"strings"

	"github.com/microsoft/azure-schema-operator/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	tenantID     string
	clientSecret string
	clientID     string
	deltaCmd     string
	useMSI       bool
)

// Wrapper is a delta-kusto thin wrapper
type Wrapper struct {
	Executer string
}

type execConfig struct {
	Uri            string
	DBs            []string
	KqlFile        string
	FailIfDataLoss bool
}

const cfgSchemaDeployment = `
sendErrorOptIn: false
failIfDataLoss: {{ $.FailIfDataLoss }}
jobs:{{range $db := .DBs}}
  push-{{$db}}-to-prod:
    current:
      adx:
        clusterUri:  {{$.Uri}} 
        database: {{ $db }}
    target:
      scripts:
        - filePath: {{$.KqlFile}} 
    action:
      # filePath: prod-update.kql
      pushToCurrent: true{{end}}`

const secretToken = `
tokenProvider:
  login:
    tenantId: to-be-overridden
    clientId: to-be-overridden
    secret: to-be-overridden
`
const msiToken = `
tokenProvider:
  systemManagedIdentity:  true
`

func init() {
	viper.SetDefault(config.DeltaCMDKey, "/bin/delta-kusto")
	useMSI = viper.GetBool(config.AzureUseMSIKey)
	tenantID = strings.TrimSpace(viper.GetString(config.AzureTenantIDKey))
	clientSecret = strings.TrimSpace(viper.GetString(config.AzureClientSecretKey))
	clientID = strings.TrimSpace(viper.GetString(config.AzureClientIDKey))
	deltaCmd = strings.TrimSpace(viper.GetString(config.DeltaCMDKey))
}

// NewDeltaWrapper returns a `Wrapper` for delta-kusto
func NewDeltaWrapper() *Wrapper {
	wrap := &Wrapper{}

	return wrap
}

// CreateExecConfiguration returns a job configuration file for delta-kusto
func (w *Wrapper) CreateExecConfiguration(uri string, dbs []string, kqlFile string, failIfDataLoss bool) (string, error) {

	log.Debug().Msg("open template file")

	t, err := template.New("cfgTempalte").Parse(cfgSchemaDeployment)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse template")
		return "", err
	}

	// open target config
	log.Debug().Msg("open target config")
	jobPath := "/tmp"
	f, err := os.CreateTemp(jobPath, "job-*.yaml")
	if err != nil {
		log.Error().Err(err).Msg("failed to open file")
		return "", err
	}
	defer f.Close()

	log.Debug().Strs("dbs", dbs).Str("kql", kqlFile).Msg("define config")
	exConfig := execConfig{
		Uri:            uri,
		DBs:            dbs,
		KqlFile:        kqlFile,
		FailIfDataLoss: failIfDataLoss,
	}
	log.Debug().Msgf("execute template config onto: %s", f.Name())
	err = t.Execute(f, exConfig)
	if err != nil {
		log.Error().Err(err).Msg("failed to execute template")
	}

	if useMSI {
		_, err = f.WriteString(msiToken)
	} else {
		_, err = f.WriteString(secretToken)
	}

	return f.Name(), err
}

// RunDeltaKusto runs delta-kusto on the provided job configuration file.
func RunDeltaKusto(deltaCfgfile string) error {
	log.Debug().Str("tenant", tenantID).Str("client", clientID).Str("sec", clientSecret).Msgf("about to run delta-kusto on: %s", deltaCfgfile)
	args := []string{"-p", deltaCfgfile}

	if useMSI {
		log.Debug().Msg("Using MSI - no auth info needed")
	} else {
		args = append(args, "-o", "tokenProvider.login.tenantId="+tenantID, "tokenProvider.login.clientId="+clientID, "tokenProvider.login.secret="+clientSecret)
	}
	cmd := exec.Command(deltaCmd, args...)
	cmd.Env = append(os.Environ(),
		"PATH=/bin/",
		"DOTNET_SYSTEM_GLOBALIZATION_INVARIANT=1",
	)
	cmd.Stdout = log.Level(zerolog.InfoLevel).With().Str("delta-kusto", deltaCfgfile).Logger()
	cmd.Stderr = log.Level(zerolog.ErrorLevel).With().Str("delta-kusto", deltaCfgfile).Logger()
	err := cmd.Run()
	if err != nil {
		eerr, ok := err.(*exec.ExitError)
		if ok {
			log.Error().Err(eerr).Msgf("cmd.Run() failed with exit code: %d, error: %s ", eerr.ExitCode(), string(eerr.Stderr))
			return err
		}
		log.Error().Err(err).Msg("cmd.Run() failed ")
		return err
	}
	log.Info().Msgf("Execution of %s done", deltaCfgfile)
	return nil
}
