// +build !local

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

package k8s

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloudfoundry-community/go-cfenv"
	brokerHttp "github.com/trustedanalytics/kubernetes-broker/http"
	v1 "k8s.io/api/core/v1"
	"time"
)

type K8sCreatorConnector struct {
	ApiVersion       string
	Server           string
	Username         string
	Password         string
	Client           *http.Client
	OrgQuota         int
	KubernetesClient KubernetesClientCreator
}

type K8sCreatorConnectorCredentials struct {
	URL         string `envconfig:"connector_url" required:"true"`
	Username    string `envconfig:"connector_username" required:"true"`
	Password    string `envconfig:"connector_password" required:"true"`
	MaxOrgQuota int    `envconfig:"max_org_quota" required:"true"`
}

func NewK8sCreatorConnector(server, user, pass string, maxOrgQuota int) *K8sCreatorConnector {
	clientCreator, _, err := brokerHttp.GetHttpClientWithBasicAuth()
	if err != nil {
		logger.Panic("Can't get http client!", err)
	}

	return &K8sCreatorConnector{
		Server:           server,
		Username:         user,
		Password:         pass,
		Client:           clientCreator,
		OrgQuota:         maxOrgQuota,
		KubernetesClient: &KubernetesRestCreator{},
	}
}

func (k *K8sCreatorConnector) IsApiWorking(credential K8sClusterCredentials) bool {
	req_url := credential.Server + "/api/v1"
	statusCde, _, err := brokerHttp.RestGET(req_url, &brokerHttp.BasicAuth{credential.Username, credential.Password}, k.Client)

	if err != nil {
		logger.Error("[IsApiWorking] Error: ", err)
		return false
	}
	return statusCde == 200
}

func (k *K8sCreatorConnector) DeleteCluster(org string) error {
	status, _, err := brokerHttp.RestDELETE(k.Server+"/clusters/"+org, "", &brokerHttp.BasicAuth{k.Username, k.Password}, k.Client)
	if status != 204 {
		logger.Error("[DeleteCluster] Error - Cluster not exist! Org:", org)
		return err
	}
	return nil
}
func (k *K8sCreatorConnector) GetCluster(org string) (int, K8sClusterCredentials, error) {
	url := k.Server + "/clusters/" + org
	k8sCreatorPostClusterResponse := K8sClusterCredentials{}

	logger.Info("[GetCluster] GetCluster on url: ", url)
	status, resp, err := brokerHttp.RestGET(url, &brokerHttp.BasicAuth{k.Username, k.Password}, k.Client)

	if status != 200 {
		return status, K8sClusterCredentials{}, errors.New("Cluster not exist!")
	}

	err = json.Unmarshal(resp, &k8sCreatorPostClusterResponse)
	if err != nil {
		return status, K8sClusterCredentials{}, err
	}
	return status, k8sCreatorPostClusterResponse, nil
}
func (k *K8sCreatorConnector) PostCluster(org string) (int, error) {
	err := k.checkIfClustersQuotaNotExeeded()
	if err != nil {
		return -1, err
	}

	url := k.Server + "/clusters/" + org
	logger.Info("[PostCluster] PostCluster on url: ", url)
	status, _, err := brokerHttp.RestPUT(url, "", &brokerHttp.BasicAuth{k.Username, k.Password}, k.Client)

	if err != nil {
		return -1, err
	}
	return status, nil
}

func (k *K8sCreatorConnector) CreateSecretForPrivateTapRepo(creds K8sClusterCredentials) error {
	c, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return err
	}

	secret := v1.Secret{}
	secret.Name = "private-tap-repo-secret"
	secret.Type = v1.SecretTypeDockercfg
	secret.Data = map[string][]byte{}

	secretValues := map[string]string{}
	secretValues["username"] = cfenv.CurrentEnv()["KUBE_REPO_USER"]
	secretValues["password"] = cfenv.CurrentEnv()["KUBE_REPO_PASS"]
	secretValues["email"] = cfenv.CurrentEnv()["KUBE_REPO_MAIL"]
	secretValues["auth"] = base64.StdEncoding.EncodeToString([]byte(secretValues["username"] + ":" + secretValues["password"]))

	secretContent := map[string]map[string]string{}
	secretContent[cfenv.CurrentEnv()["KUBE_REPO_URL"]] = secretValues
	secret.Data[".dockercfg"], err = json.Marshal(secretContent)

	if err != nil {
		return err
	}

	_, err = c.CoreV1().Secrets(v1.NamespaceDefault).Create(&secret)
	if err != nil {
		return err
	}
	return nil
}

func (k *K8sCreatorConnector) checkIfClustersQuotaNotExeeded() error {
	clusters, err := k.GetClusters()
	if err != nil {
		return err
	}

	if len(clusters) >= k.OrgQuota {
		return errors.New(fmt.Sprintf("Clusters quota exceeded! Max allowed level is: %d", k.OrgQuota))
	} else {
		return nil
	}
}

func (k *K8sCreatorConnector) GetClusters() ([]K8sClusterCredentials, error) {
	k8sCreatorGetClustersResponse := []K8sClusterCredentials{}

	_, resp, err := brokerHttp.RestGET(k.Server+"/clusters", &brokerHttp.BasicAuth{k.Username, k.Password}, k.Client)
	logger.Debug("RESP: ", string(resp))
	err = json.Unmarshal(resp, &k8sCreatorGetClustersResponse)
	if err != nil {
		logger.Error("[GetClusters] Error: ", err)
		return []K8sClusterCredentials{}, err
	}
	return k8sCreatorGetClustersResponse, nil
}

func (k *K8sCreatorConnector) GetOrCreateCluster(org string) (K8sClusterCredentials, error) {
	wasCreated := false
	for {
		status, kresp, err := k.GetCluster(org)

		if status == 200 {
			if k.IsApiWorking(kresp) {
				logger.Warning("[GetOrCreateCluster] Cluster already created for org:", org)
				if wasCreated == true {
					err = k.CreateSecretForPrivateTapRepo(kresp)
					if err != nil {
						logger.Error("[GetOrCreateCluster] ERROR: Unable to create secret with credentials for private TAP repo!", err)
					}
				}
				return kresp, nil
			}
		} else if status == 404 {
			if !wasCreated {
				logger.Info("[GetOrCreateCluster] Creating cluster for org:", org)
				status, err = k.PostCluster(org)
				if err != nil {
					logger.Error("[GetOrCreateCluster] ERROR: PostCluster", err)
					return K8sClusterCredentials{}, err
				} else if status == 409 {
					logger.Error("PostCluster: Unexpected cluster conflict!", err)
					return K8sClusterCredentials{}, errors.New("UnExpected Cluster conflict")
				}
				wasCreated = true
			} else {
				return K8sClusterCredentials{}, errors.New("After creating CLuster bad response received")
			}
		} else if status == 204 {
			logger.Info("[GetOrCreateCluster] Waiting for cluster to finish creating for org:", org)
		} else if err != nil {
			logger.Error("[GetOrCreateCluster] ERROR: GetCluster! We will not fetch/create requested cluster!", err)
			return K8sClusterCredentials{}, err
		}
		time.Sleep(30 * time.Second)
	}
}
