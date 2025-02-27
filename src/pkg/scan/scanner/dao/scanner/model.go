// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scanner

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Registration represents a named configuration for invoking a scanner via its adapter.
// UUID will be used to track the scanner.Endpoint as unique ID
type Registration struct {
	// Basic information
	// int64 ID is kept for being aligned with previous DB schema
	ID          int64  `orm:"pk;auto;column(id)" json:"-"`
	UUID        string `orm:"unique;column(uuid)" json:"uuid"`
	Name        string `orm:"unique;column(name);size(128)" json:"name"`
	Description string `orm:"column(description);null;size(1024)" json:"description"`
	URL         string `orm:"column(url);unique;size(512)" json:"url"`
	Disabled    bool   `orm:"column(disabled);default(true)" json:"disabled"`
	IsDefault   bool   `orm:"column(is_default);default(false)" json:"is_default"`
	Health      bool   `orm:"-" json:"health"`

	// Authentication settings
	// "None","Basic" and "Bearer" can be supported
	Auth             string `orm:"column(auth);size(16)" json:"auth"`
	AccessCredential string `orm:"column(access_cred);null;size(512)" json:"access_credential,omitempty"`

	// Http connection settings
	SkipCertVerify bool `orm:"column(skip_cert_verify);default(false)" json:"skip_certVerify"`

	// Adapter settings
	Adapter string `orm:"column(adapter);size(128)" json:"adapter"`
	Vendor  string `orm:"column(vendor);size(128)" json:"vendor"`
	Version string `orm:"column(version);size(32)" json:"version"`

	// Timestamps
	CreateTime time.Time `orm:"column(create_time);auto_now_add;type(datetime)" json:"create_time"`
	UpdateTime time.Time `orm:"column(update_time);auto_now;type(datetime)" json:"update_time"`
}

// TableName for Endpoint
func (r *Registration) TableName() string {
	return "scanner_registration"
}

// FromJSON parses json data
func (r *Registration) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), r)
}

// ToJSON marshals endpoint to JSON data
func (r *Registration) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Validate endpoint
func (r *Registration) Validate(checkUUID bool) error {
	if checkUUID && len(r.UUID) == 0 {
		return errors.New("malformed endpoint")
	}

	if len(r.Name) == 0 {
		return errors.New("missing registration name")
	}

	err := checkURL(r.URL)
	if err != nil {
		return errors.Wrap(err, "scanner registration validate")
	}

	if len(r.Adapter) == 0 ||
		len(r.Vendor) == 0 ||
		len(r.Version) == 0 {
		return errors.Errorf("missing adapter settings in registration %s:%s", r.Name, r.URL)
	}

	return nil
}

// Check the registration URL with url package
func checkURL(u string) error {
	if len(strings.TrimSpace(u)) == 0 {
		return errors.New("empty url")
	}

	uri, err := url.Parse(u)
	if err == nil {
		if uri.Scheme != "http" && uri.Scheme != "https" {
			err = errors.New("invalid scheme")
		}
	}

	return err
}
