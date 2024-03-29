package schemaregistry

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

// UserAgent returns the UserAgent string to use when sending http.Requests.
func UserAgent() string {
	return "Azure-SDK-For-Go/" + Version() + " schemaregistry/2021-10"
}

// Version returns the semantic version (see http://semver.org) of the client.
func Version() string {
	return "v1.0.0"
}
