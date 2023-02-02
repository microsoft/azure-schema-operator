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
func GetTableCachingPolicy(ctx context.Context, client *kusto.Client, database string, tableName string) (*types.CachingPolicy, error) {
	policy := &types.CachingPolicy{}
	err := GetTablePolicy(ctx, client, database, tableName, policy)
	if err != nil {
		log.Error().Err(err).Msg("failed to get table caching policy")
		return nil, err
	}
	return policy, nil
}

// GetTableRetentionPolicy returns the retention policy of a table
// it furst checks if a policy is defined on the table, if not it checks if a policy is defined on the database.
func GetTableRetentionPolicy(ctx context.Context, client *kusto.Client, database string, tableName string) (*types.RetentionPolicy, error) {
	policy := &types.RetentionPolicy{}
	err := GetTablePolicy(ctx, client, database, tableName, policy)
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
	stmtStr := fmt.Sprintf(".alter table %s policy retention ``` %s ```", tableName, policyStr)
	stmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true, SuppressWarning: true})).UnsafeAdd(stmtStr)
	iterator, err := client.Mgmt(ctx, database, stmt)
	if err != nil {
		log.Error().Err(err).Msg("failed to alter table retention policy")
		return nil, err
	}
	defer iterator.Stop()
	rec := retentionRecord{}
	newPolicy := &types.RetentionPolicy{}
	err = iterator.DoOnRowOrError(
		func(row *table.Row, inlineError *errors.Error) error {
			if row != nil {
				row.ToStruct(&rec)
				log.Debug().Msgf("got policy: %+v, policy: %s", rec, rec.Policy)
				if rec.Policy != "null" {
					newPolicy = &types.RetentionPolicy{}
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

// SetTableCachingPolicy sets the retention policy of a table
func SetTableCachingPolicy(ctx context.Context, client *kusto.Client, database string, tableName string, policy *types.CachingPolicy) (*types.CachingPolicy, error) {
	policyStr, err := json.Marshal(policy)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal policy")
		return nil, err
	}
	stmtStr := fmt.Sprintf(".alter table %s policy %s ``` %s ```", tableName, policy.GetShortName(), policyStr)
	stmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true, SuppressWarning: true})).UnsafeAdd(stmtStr)
	iterator, err := client.Mgmt(ctx, database, stmt)
	if err != nil {
		log.Error().Err(err).Msgf("failed to alter table %s policy", policy.GetShortName())
		return nil, err
	}
	defer iterator.Stop()
	rec := retentionRecord{}
	newPolicy := &types.CachingPolicy{}
	err = iterator.DoOnRowOrError(
		func(row *table.Row, inlineError *errors.Error) error {
			if row != nil {
				row.ToStruct(&rec)
				log.Debug().Msgf("got policy: %+v, policy: %s", rec, rec.Policy)
				if rec.Policy != "null" {
					newPolicy = &types.CachingPolicy{}
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
	if newPolicy.DataHotSpan != policy.DataHotSpan || newPolicy.IndexHotSpan != policy.IndexHotSpan {
		log.Error().Msgf("returned policy doesn't match requested policy, %s vs %s", newPolicy.DataHotSpan, policy.IndexHotSpan)
		return nil, fmt.Errorf("returned policy doesn't match requested policy")
	}
	return newPolicy, nil
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
				log.Debug().Msgf("got policy: %+v, policy: %s", rec, rec.Policy)
				if rec.Policy != "null" {
					err = json.Unmarshal([]byte(rec.Policy), policy)
					if err != nil {
						log.Error().Err(err).Msg("failed to unmarshal policy")
						return err
					}
					log.Debug().Msgf("got policy: %+v", policy)
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
	return GetDatabasePolicy(database, policy, client, ctx)

}

// GetDatabasePolicy returns a requested policy of a table
func GetDatabasePolicy(database string, policy types.Policy, client *kusto.Client, ctx context.Context) error {
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
