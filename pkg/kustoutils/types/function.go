package types
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"fmt"
	"strings"
)

type KustoFunction struct {
	Name       string `json:"name"`
	DocString  string `json:"docString"`
	Folder     string `json:"folder"`
	Parameters string `json:"parameters"`
	Body       string `json:"body"`
}

// GetFunctionQuery returns a query to show function defined
func (f *KustoFunction) GetFunctionQuery() string {
	return fmt.Sprintf(".show function %s", f.Name)
}

// SetFunctionQuery return a query to create or alter the function
func (f *KustoFunction) SetFunctionQuery() string {
	var sb strings.Builder
	sb.WriteString(".create-or-alter function ")

	if f.DocString != "" || f.Folder != "" {
		sb.WriteString("with (")
		if f.DocString != "" {
			sb.WriteString(fmt.Sprintf("docstring = '%s' ", f.DocString))
		}
		if f.Folder != "" {
			if f.DocString != "" {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("folder = '%s' ", f.Folder))
		}
		sb.WriteString(") ")
	}
	sb.WriteString(f.Name)
	sb.WriteString(" ")
	sb.WriteString(f.Parameters)
	sb.WriteString("  ")
	sb.WriteString(f.Body)
	return sb.String()

}

// Equals returns true if the two functions are equal
func (f *KustoFunction) Equals(other *KustoFunction) bool {
	return f.Name == other.Name && f.Parameters == other.Parameters && f.Body == other.Body && f.DocString == other.DocString && f.Folder == other.Folder
}
