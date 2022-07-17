package sqlutils

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/microsoft/azure-schema-operator/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	useMSI          bool
	sqlpackgeUser   string
	sqlpackgePass   string
	sqlpackgeCmd    string
	parallelWorkers int
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault(config.SQLPackageCMDKey, "/sqlpackage/sqlpackage")
	viper.SetDefault(config.ParallelWorkers, 10)
	viper.SetDefault(config.AllowLocalDacPac, false)
	useMSI = viper.GetBool(config.AzureUseMSIKey)
	sqlpackgeUser = strings.TrimSpace(viper.GetString(config.SQLPackageUser))
	sqlpackgePass = strings.TrimSpace(viper.GetString(config.SQLPackagePass))
	sqlpackgeCmd = strings.TrimSpace(viper.GetString(config.SQLPackageCMDKey))
	parallelWorkers = viper.GetInt(config.ParallelWorkers)
}

// updateDacPac creates a duplicate dacpac with the source schema replaced with a destenation schema.
// this is used to support multi-tenant solutions with schema per tenant.
func updateDacPac(dstDacPac string, srcDacPac string, sourceSchema, tenantSchema string) error {

	srcChecksum := ""
	dstChecksum := ""
	var originContent []byte

	darchive, err := os.Create(dstDacPac)
	if err != nil {
		log.Error().Err(err).Msgf("failed to create destination file at %s", dstDacPac)
		return err
	}
	defer darchive.Close()
	zipWriter := zip.NewWriter(darchive)

	archive, err := zip.OpenReader(srcDacPac)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to open source dacpac at %s", srcDacPac)
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		switch f.Name {
		case "model.xml":
			dstChecksum, srcChecksum = replaceContent(zipWriter, f, sourceSchema, tenantSchema)
		case "refactor.xml":
			replaceContent(zipWriter, f, sourceSchema, tenantSchema)
		case "postdeploy.sql":
			replaceContent(zipWriter, f, sourceSchema, tenantSchema)
		case "predeploy.sql":
			replaceContent(zipWriter, f, sourceSchema, tenantSchema)
		case "Origin.xml":
			log.Info().Msg("handeling origin file - getting content")
			fileInArchive, err := f.Open()
			if err != nil {
				panic(err)
			}

			originContent, err = ioutil.ReadAll(fileInArchive)
			if err != nil {
				panic(err)
			}
		default:
			log.Info().Msg("other files - just copy")
			dstFile, err := zipWriter.Create(f.Name)
			if err != nil {
				panic(err)
			}
			fileInArchive, err := f.Open()
			if err != nil {
				panic(err)
			}
			if _, err := io.Copy(dstFile, fileInArchive); err != nil {
				panic(err)
			}
			fileInArchive.Close()
		}
	}

	log.Info().Msg("updating origin file with new checksum")
	err = updateOriginXML(zipWriter, originContent, srcChecksum, dstChecksum)
	if err != nil {
		panic(err)
	}

	log.Debug().Msg("closing zip archive...")
	zipWriter.Close()
	return nil
}

func replaceContent(zipWriter *zip.Writer, f *zip.File, sourceSchema string, tenantSchema string) (string, string) {
	log.Info().Msg("handeling model file - replace schema")
	dstFile, err := zipWriter.Create(f.Name)
	if err != nil {
		panic(err)
	}
	fileInArchive, err := f.Open()
	if err != nil {
		panic(err)
	}

	read, err := ioutil.ReadAll(fileInArchive)
	if err != nil {
		panic(err)
	}

	newContents := strings.Replace(string(read), sourceSchema, tenantSchema, -1)

	n, err := dstFile.Write([]byte(newContents))
	if err != nil {
		panic(err)
	}
	log.Debug().Msgf("write n: %d bytes\n", n)

	log.Debug().Msg("compute new models checksum")
	fileInArchive.Close()
	hasher := sha256.New()
	hasher.Write([]byte(newContents))
	dstChecksum := strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))

	hasher.Reset()
	hasher.Write(read)
	srcChecksum := strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))

	log.Debug().Msgf("source check sum: %s \n", srcChecksum)
	log.Debug().Msgf("new checksum: %s \n", dstChecksum)
	return dstChecksum, srcChecksum
}

func updateOriginXML(zipWriter *zip.Writer, originContent []byte, srcChecksum, dstChecksum string) error {

	dstOriginFile, err := zipWriter.Create("Origin.xml")
	if err != nil {
		// panic(err)
		return err
	}
	newContents := strings.Replace(string(originContent), srcChecksum, dstChecksum, -1)

	n, err := dstOriginFile.Write([]byte(newContents))
	if err != nil {
		// panic(err)
		return err
	}
	log.Info().Msgf("write n: %d bytes into Origin.xml", n)

	return nil
}

// RunDacPac runs DacPac on a target DB by using sqlpackage.
func RunDacPac(dacPacFile string, targetServer string, targetDB string, sqlpackageOptions string) error {
	log.Debug().Str("targetServer", targetServer).Str("targetDB", targetDB).Msgf("about to run sqlpackage on: %s", dacPacFile)
	args := []string{"/SourceFile:" + dacPacFile, "/Action:Publish"}

	if sqlpackageOptions != "" {
		optionsArray := strings.Split(sqlpackageOptions, " ")
		args = append(args, optionsArray...)
	}

	if useMSI {
		log.Debug().Msg("Using MSI - no auth info needed")
		connString := fmt.Sprintf("Server=%s;database=%s;Authentication=ActiveDirectoryMSI", targetServer, targetDB)
		args = append(args, "/tcs:"+connString)
	} else {
		args = append(args, "/tsn:"+targetServer, "/TargetDatabaseName:"+targetDB)
		args = append(args, "/tu:"+sqlpackgeUser, "/tp:"+sqlpackgePass)
	}
	cmd := exec.Command(sqlpackgeCmd, args...)
	cmd.Env = append(os.Environ(),
		"PATH=/bin/",
	)
	cmd.Stdout = log.Level(zerolog.InfoLevel).With().Str("sqlpackage", dacPacFile).Logger()
	cmd.Stderr = log.Level(zerolog.ErrorLevel).With().Str("sqlpackage", dacPacFile).Logger()
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
	log.Info().Msgf("Execution of %s done", dacPacFile)
	return nil
}
