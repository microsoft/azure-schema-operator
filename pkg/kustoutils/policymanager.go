package kustoutils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/errors"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils/types"
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

// GetTableCachingPolicy returns the caching policy of a table
// it furst checks if a policy is defined on the table, if not it checks if a policy is defined on the database.
func GetTableCachingPolicy(ctx context.Context, client *kusto.Client, database string, tableName string) (string, error) {
	policy := &types.CachingPolicy{}
	var err error
	if tableName != "" {
		err = GetTablePolicy(ctx, client, database, tableName, policy)
	} else {
		err = GetDatabasePolicy(ctx, client, database, policy)
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get table caching policy")
		return "", err
	}
	policyStr := ConvertTimeFormat(*&policy.DataHotSpan.Value)
	return policyStr, nil
}

// GetTableRetentionPolicy returns the retention policy of a table
// it furst checks if a policy is defined on the table, if not it checks if a policy is defined on the database.
func GetTableRetentionPolicy(ctx context.Context, client *kusto.Client, database string, tableName string) (*types.RetentionPolicy, error) {
	policy := &types.RetentionPolicy{}
	var err error
	if tableName != "" {
		log.Debug().Msgf("getting retention policy for table %s", tableName)
		err = GetTablePolicy(ctx, client, database, tableName, policy)
	} else {
		err = GetDatabasePolicy(ctx, client, database, policy)
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get table retention policy")
		return nil, err
	}
	return policy, nil
}

// SetTableRetentionPolicy sets the retention policy of a table
func SetTableRetentionPolicy(ctx context.Context, client *kusto.Client, database string, tableName string, policy *types.RetentionPolicy) (*types.RetentionPolicy, error) {
	policyStr, err := json.Marshal(policy)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal policy")
		return nil, err
	}
	var stmtStr string
	if tableName != "" {
		stmtStr = fmt.Sprintf(".alter table %s policy retention ``` %s ```", tableName, policyStr)
	} else {
		stmtStr = fmt.Sprintf(".alter database %s policy retention ``` %s ```", database, policyStr)
	}
	
	stmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true, SuppressWarning: true})).UnsafeAdd(stmtStr)
	iterator, err := client.Mgmt(ctx, database, stmt)
	if err != nil {
		log.Error().Err(err).Msg("failed to alter retention policy")
		return nil, err
	}
	defer iterator.Stop()
	rec := retentionRecord{}
	newPolicy := &types.RetentionPolicy{}
	err = iterator.DoOnRowOrError(
		func(row *table.Row, inlineError *errors.Error) error {
			if row != nil {
				row.ToStruct(&rec)
				log.Debug().Msgf("SetTableRetentionPolicy: got policy: %+v, policy: %s", rec, rec.Policy)
				if rec.Policy != "null" {
					newPolicy = &types.RetentionPolicy{}
					err = json.Unmarshal([]byte(rec.Policy), newPolicy)
					if err != nil {
						log.Error().Err(err).Msg("failed to unmarshal policy")
						return err
					}
					log.Debug().Msgf("SetTableRetentionPolicy: got policy: %+v", newPolicy)
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

// SetTableCachingPolicy sets the retention policy of a table
func SetTableCachingPolicy(ctx context.Context, client *kusto.Client, database string, tableName string, policy string) (string, error) {
	var stmtStr string
	if tableName != "" {
		stmtStr = fmt.Sprintf(".alter table %s policy caching hot = %s", tableName, policy)
	} else {
		stmtStr = fmt.Sprintf(".alter database %s policy caching hot = %s", database, policy)
	}
	stmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true, SuppressWarning: true})).UnsafeAdd(stmtStr)
	iterator, err := client.Mgmt(ctx, database, stmt)
	if err != nil {
		log.Error().Err(err).Msg("failed to alter table caching policy")
		return "", err
	}
	defer iterator.Stop()
	rec := retentionRecord{}
	newPolicy := &types.CachingPolicy{}
	err = iterator.DoOnRowOrError(
		func(row *table.Row, inlineError *errors.Error) error {
			if row != nil {
				row.ToStruct(&rec)
				log.Debug().Msgf("SetTableCachingPolicy: got policy: %+v, policy: %s", rec, rec.Policy)
				if rec.Policy != "null" {
					newPolicy = &types.CachingPolicy{}
					err = json.Unmarshal([]byte(rec.Policy), newPolicy)
					if err != nil {
						log.Error().Err(err).Msg("failed to unmarshal policy")
						return err
					}
					log.Debug().Msgf("SetTableCachingPolicy: got policy: %+v", newPolicy)
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
		return "", err
	}
	newPolicyStr := ConvertTimeFormat(newPolicy.DataHotSpan.Value)

	if newPolicyStr != policy {
		log.Error().Msgf("returned policy doesn't match requested policy, %s vs %s", newPolicy.DataHotSpan, policy)
		return "", fmt.Errorf("returned policy doesn't match requested policy")
	}
	return newPolicyStr, nil
}

// GetTablePolicy returns a requested policy of a table
// it furst checks if a policy is defined on the table, if not it checks if a policy is defined on the database.
func GetTablePolicy(ctx context.Context, client *kusto.Client, database string, tableName string, policy types.Policy) error {
	found := false
	// check if a policy is defined on the table
	stmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true, SuppressWarning: true})).UnsafeAdd(".show table " + tableName + " policy " + policy.GetShortName())
	iterator, err := client.Mgmt(ctx, database, stmt)
	if err != nil {
		log.Error().Err(err).Msg("failed to get table policy")
		return err
	}
	defer iterator.Stop()
	rec := retentionRecord{}
	err = iterator.DoOnRowOrError(
		func(row *table.Row, inlineError *errors.Error) error {
			if row != nil {
				row.ToStruct(&rec)
				log.Debug().Msgf("GetTablePolicy: got policy: %+v, policy: %s", rec, rec.Policy)
				if rec.Policy != "null" {
					err = json.Unmarshal([]byte(rec.Policy), policy)
					if err != nil {
						log.Error().Err(err).Msg("failed to unmarshal policy")
						return err
					}
					log.Debug().Msgf("GetTablePolicy: got policy: %+v", policy)
					found = true
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
		return err
	}
	// if we found a policy on the table, return it
	if found {
		return nil
	}
	log.Debug().Msg("no policy defined on table, checking database")

	// check if a policy is defined on the database
	// ignore inline errors - not relevant for this use case
	// log.Debug().Msgf("dbname: %s", dbName)
	// if we found a policy on the table, return it
	return GetDatabasePolicy(ctx,client,database, policy)

}

// GetDatabasePolicy returns a requested policy of a table
func GetDatabasePolicy(ctx context.Context,client *kusto.Client,database string, policy types.Policy) error {
	dbstmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true, SuppressWarning: true})).UnsafeAdd(".show database " + database + " policy " + policy.GetShortName())
	dbiterator, err := client.Mgmt(ctx, database, dbstmt)
	if err != nil {
		log.Error().Err(err).Msg("failed to get database policy")
		return err
	}
	defer dbiterator.Stop()
	dbRec := dbRetentionRecord{}
	err = dbiterator.DoOnRowOrError(
		func(row *table.Row, inlineError *errors.Error) error {
			if row != nil {
				log.Debug().Msgf("got row: %+v", row)
				row.ToStruct(&dbRec)
				log.Debug().Msgf("got database policy: %+v, policy: %s", dbRec, dbRec.Policy)
				json.Unmarshal([]byte(dbRec.Policy), policy)
			} else {

				log.Error().Msgf("got inline error: %s", inlineError.Error())
			}

			return nil
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to iterate results")
		return err
	}

	return nil
}
