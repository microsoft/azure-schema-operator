package types

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
type PolicyName string

const (
	Caching            PolicyName = "CachingPolicy"
	ExtentsMerge       PolicyName = "ExtentsMergePolicy"
	DataSharding       PolicyName = "DataShardingPolicy"
	Retention          PolicyName = "RetentionPolicy"
	StreamingIngestion PolicyName = "StreamingIngestionPolicy"
	IngestionBatching  PolicyName = "IngestionBatchingPolicy"
)

type Policy interface {
	GetName() PolicyName
	GetShortName() string
}

// RetentionPolicy defines a retention policy
type RetentionPolicy struct {
	SoftDeletePeriod string `json:"softDeletePeriod"`
	// +kubebuilder:validation:Enum:=Disabled;Enabled
	Recoverability string `json:"recoverability"`
}

func (p *RetentionPolicy) GetName() PolicyName {
	return Retention
}

func (p *RetentionPolicy) GetShortName() string {
	return "retention"
}

type Valuer struct {
	Value string `json:"value"`
}

// RetentionPolicy defines a retention policy
type CachingPolicy struct {
	DataHotSpan  Valuer `json:"dataHotSpan"`
	IndexHotSpan Valuer `json:"indexHotSpan"`
}

func (p *CachingPolicy) GetName() PolicyName {
	return Caching
}

func (p *CachingPolicy) GetShortName() string {
	return "caching"
}

type ExtentsMergePolicy struct {
	RowCountUpperBoundForMerge       int    `json:"RowCountUpperBoundForMerge"`
	OriginalSizeMBUpperBoundForMerge int    `json:"OriginalSizeMBUpperBoundForMerge"`
	MaxExtentsToMerge                int    `json:"MaxExtentsToMerge"`
	LoopPeriod                       string `json:"LoopPeriod"`
	MaxRangeInHours                  int    `json:"MaxRangeInHours"`
	AllowRebuild                     bool   `json:"AllowRebuild"`
	AllowMerge                       bool   `json:"AllowMerge"`
	Lookback                         struct {
		Kind         string      `json:"Kind"`
		CustomPeriod interface{} `json:"CustomPeriod"`
	} `json:"Lookback"`
}

func (p *ExtentsMergePolicy) GetName() PolicyName {
	return ExtentsMerge
}

func (p *ExtentsMergePolicy) GetShortName() string {
	return "merge"
}
