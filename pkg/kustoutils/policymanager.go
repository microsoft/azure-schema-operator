package kustoutils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/errors"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	kustov1alpha1 "github.com/microsoft/azure-schema-operator/apis/kusto/v1alpha1"
	"github.com/rs/zerolog/log"
)

type retentionRecord struct {
	PolicyName string `json:"PolicyName"`
	EntityName string `json:"EntityName"`
	Policy     string `json:"Policy"`
}

type dbRetentionRecord struct {
	PolicyName    string `json:"PolicyName"`
	EntityName    string `json:"EntityName"`
	Policy        string `json:"Policy"`
	ChildEntities string `json:"ChildEntities"`
	EntityType    string `json:"EntityType"`
}

// GetTableRetentionPolicy returns the retention policy of a table
// it furst checks if a policy is defined on the table, if not it checks if a policy is defined on the database.
func GetTableRetentionPolicy(ctx context.Context, client *kusto.Client, database string, tableName string) (*kustov1alpha1.KustoRetentionPolicy, error) {
	var policy *kustov1alpha1.KustoRetentionPolicy
	// check if a policy is defined on the table
	stmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true, SuppressWarning: true})).UnsafeAdd(".show table " + tableName + " policy retention ")
	iterator, err := client.Mgmt(ctx, database, stmt)
	if err != nil {
		log.Error().Err(err).Msg("failed to get table retention policy")
		return nil, err
	}
	defer iterator.Stop()
	rec := retentionRecord{}
	err = iterator.DoOnRowOrError(
		func(row *table.Row, inlineError *errors.Error) error {
			if row != nil {
				row.ToStruct(&rec)
				log.Debug().Msgf("got policy: %+v, policy: %s", rec, rec.Policy)
				if rec.Policy != "null" {
					policy = &kustov1alpha1.KustoRetentionPolicy{}
					err = json.Unmarshal([]byte(rec.Policy), policy)
					if err != nil {
						log.Error().Err(err).Msg("failed to unmarshal policy")
						return err
					}
					log.Debug().Msgf("got policy: %+v", policy)
				}
			} else {
				// ignore inline errors - not relevant for this use case
				log.Error().Msgf("got inline error: %s", inlineError.Error())
			}
			// log.Debug().Msgf("dbname: %s", dbName)
			return nil
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to iterate results")
		return nil, err
	}
	// if we found a policy on the table, return it
	if policy != nil {
		return policy, nil
	}
	log.Debug().Msg("no policy defined on table, checking database")

	// check if a policy is defined on the database
	dbstmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true, SuppressWarning: true})).UnsafeAdd(".show database " + database + " policy retention ")
	dbiterator, err := client.Mgmt(ctx, database, dbstmt)
	if err != nil {
		log.Error().Err(err).Msg("failed to get database retention policy")
		return nil, err
	}
	defer dbiterator.Stop()
	dbRec := dbRetentionRecord{}
	err = dbiterator.DoOnRowOrError(
		func(row *table.Row, inlineError *errors.Error) error {
			if row != nil {
				log.Debug().Msgf("got row: %+v", row)
				row.ToStruct(&dbRec)
				log.Debug().Msgf("got database policy: %+v, policy: %s", dbRec, dbRec.Policy)
				policy = &kustov1alpha1.KustoRetentionPolicy{}
				json.Unmarshal([]byte(dbRec.Policy), policy)
			} else {
				// ignore inline errors - not relevant for this use case
				log.Error().Msgf("got inline error: %s", inlineError.Error())
			}
			// log.Debug().Msgf("dbname: %s", dbName)
			return nil
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to iterate results")
		return nil, err
	}
	// if we found a policy on the table, return it
	return policy, nil

}

// SetTableRetentionPolicy sets the retention policy of a table
func SetTableRetentionPolicy(ctx context.Context, client *kusto.Client, database string, tableName string, policy *kustov1alpha1.KustoRetentionPolicy) (*kustov1alpha1.KustoRetentionPolicy, error) {
	policyStr, err := json.Marshal(policy)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal policy")
		return nil, err
	}
	stmtStr := fmt.Sprintf(".alter table %s policy retention ``` %s ```", tableName, policyStr)
	stmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true, SuppressWarning: true})).UnsafeAdd(stmtStr)
	iterator, err := client.Mgmt(ctx, database, stmt)
	if err != nil {
		log.Error().Err(err).Msg("failed to alter table retention policy")
		return nil, err
	}
	defer iterator.Stop()
	rec := retentionRecord{}
	var newPolicy *kustov1alpha1.KustoRetentionPolicy
	err = iterator.DoOnRowOrError(
		func(row *table.Row, inlineError *errors.Error) error {
			if row != nil {
				row.ToStruct(&rec)
				log.Debug().Msgf("got policy: %+v, policy: %s", rec, rec.Policy)
				if rec.Policy != "null" {
					newPolicy = &kustov1alpha1.KustoRetentionPolicy{}
					err = json.Unmarshal([]byte(rec.Policy), newPolicy)
					if err != nil {
						log.Error().Err(err).Msg("failed to unmarshal policy")
						return err
					}
					log.Debug().Msgf("got policy: %+v", newPolicy)
				}
			} else {
				// ignore inline errors - not relevant for this use case
				log.Error().Msgf("got inline error: %s", inlineError.Error())
			}
			// log.Debug().Msgf("dbname: %s", dbName)
			return nil
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to iterate results")
		return nil, err
	}
	if newPolicy.SoftDeletePeriod != policy.SoftDeletePeriod || newPolicy.Recoverability != policy.Recoverability {
		log.Error().Msgf("returned policy doesn't match requested policy, %s vs %s", newPolicy.SoftDeletePeriod, policy.SoftDeletePeriod)
		return nil, fmt.Errorf("returned policy doesn't match requested policy")
	}
	return newPolicy, nil
}
