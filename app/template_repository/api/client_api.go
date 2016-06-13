/**
 * Copyright (c) 2016 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/trustedanalytics/kubernetes-broker/catalog"
	brokerHttp "github.com/trustedanalytics/kubernetes-broker/http"
	"github.com/trustedanalytics/kubernetes-broker/logger"
)

type TemplateRepository interface {
	GetCatalog() (catalog.ServicesMetadata, error)
	GenerateParsedTemplate(request GenerateParsedTemplateRequest) (catalog.Template, error)
}

type TemplateRepositoryConnector struct {
	Address  string
	Username string
	Password string
	Client   *http.Client
}

var logger = logger_wrapper.InitLogger("api")

func NewTemplateRepositoryBasicAuth(address, username, password string) (*TemplateRepositoryConnector, error) {
	client, _, err := brokerHttp.GetHttpClientWithBasicAuth()
	if err != nil {
		return nil, err
	}
	return &TemplateRepositoryConnector{address, username, password, client}, nil
}

func NewTemplateRepositoryCa(address, username, password, certPemFile, keyPemFile, caPemFile string) (*TemplateRepositoryConnector, error) {
	client, _, err := brokerHttp.GetHttpClientWithCertAndCaFromFile(certPemFile, keyPemFile, caPemFile)
	if err != nil {
		return nil, err
	}
	return &TemplateRepositoryConnector{address, username, password, client}, nil
}

func (t *TemplateRepositoryConnector) GetCatalog() (catalog.ServicesMetadata, error) {
	url := t.Address + "/catalog"
	status, body, err := brokerHttp.RestGET(url, &brokerHttp.BasicAuth{t.Username, t.Password}, t.Client)

	services := catalog.ServicesMetadata{}
	err = json.Unmarshal(body, &services)
	if err != nil {
		logger.Error("GetCatalog error:", err)
		return services, err
	}
	if status != http.StatusOK {
		return services, errors.New("Bad response status: " + strconv.Itoa(status) + ". Body: " + string(body))
	}
	return services, nil
}

func (t *TemplateRepositoryConnector) GenerateParsedTemplate(request GenerateParsedTemplateRequest) (catalog.Template, error) {
	template := catalog.Template{}

	url := t.Address + "/catalog/parsed"
	body, err := json.Marshal(request)
	if err != nil {
		logger.Error("GenerateParsedTemplate marshall request error:", err)
		return template, err
	}

	status, body, err := brokerHttp.RestPOST(url, string(body), &brokerHttp.BasicAuth{t.Username, t.Password}, t.Client)
	err = json.Unmarshal(body, &template)
	if err != nil {
		logger.Error("GenerateParsedTemplate unmarshall response error:", err)
		return template, err
	}
	if status != http.StatusOK {
		return template, errors.New("Bad response status: " + strconv.Itoa(status) + ". Body: " + string(body))
	}
	return template, nil
}
