package schemaversions

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import "strconv"

// NameForConfigMap returns a new name for a versioned `ConfigMap` of specified `version`
func NameForConfigMap(name string, version int32) string {
	return name + "-" + strconv.Itoa(int(version))
}
