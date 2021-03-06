/*
* Copyright 2019 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package inputs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/newrelic/nri-flex/internal/load"
)

func TestNetDial(t *testing.T) {
	load.Refresh()

	config := load.Config{
		APIs: []load.API{
			{
				Name: "failure",
				Commands: []load.Command{
					load.Command{
						Dial: "fake12311290.com:9989",
					},
				},
			},
		},
	}

	var jsonOut interface{}
	expectedOutput := []byte{}
	if strings.Contains(runtime.Version(), "go1.13") {
		expectedOutput, _ = ioutil.ReadFile("../../test/payloadsExpected/portTestSingle-go113.json")
	} else {
		expectedOutput, _ = ioutil.ReadFile("../../test/payloadsExpected/portTestSingle.json")
	}

	json.Unmarshal(expectedOutput, &jsonOut)
	expectedDatastore := jsonOut.([]interface{})

	dataStore := []interface{}{}
	dataSample := map[string]interface{}{}
	processType := ""
	NetDialWithTimeout(&dataStore, config.APIs[0].Commands[0], &dataSample, config.APIs[0], &processType)

	if len(expectedDatastore) != len(dataStore) {
		t.Errorf("Incorrect number of samples generated expected: %d, got: %d", len(expectedDatastore), len(dataStore))
		t.Errorf("%v", (dataStore))
	}

	for _, sample := range expectedDatastore {
		switch sample := sample.(type) {
		case map[string]interface{}:
			for _, rSample := range dataStore {
				switch recSample := rSample.(type) {
				case map[string]interface{}:
					for key := range sample {
						if recSample[key] != nil {
							if key == "err" {
								allowedErrors := []string{"dial tcp: lookup fake12311290.com(.*?): no such host", "context deadline exceeded", "dial tcp: i/o timeout"}
								foundError := false

								for _, allowedError := range allowedErrors {
									p := regexp.MustCompile(allowedError)
									if p.MatchString(fmt.Sprintf("%v", recSample[key])) {
										foundError = true
										break
									}
								}
								if !foundError {
									t.Errorf("expected one of these errors %v", allowedErrors)
								}
							} else if fmt.Sprintf("%v", sample[key]) != fmt.Sprintf("%v", recSample[key]) {
								t.Errorf("%v want %v, got %v", key, sample[key], recSample[key])
							}
						}
					}
				}
			}
		}
	}
}
