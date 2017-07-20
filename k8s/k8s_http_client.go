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
	"net/http"
	"strconv"

	"github.com/cloudfoundry-community/go-cfenv"
	apiv1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	testcore "k8s.io/client-go/testing"

	brokerHttp "github.com/trustedanalytics/kubernetes-broker/http"
	"github.com/trustedanalytics/kubernetes-broker/logger"
)

type KubernetesClientCreator interface {
	GetNewClient(creds K8sClusterCredentials) (kubernetes.Interface, error)
}

type KubernetesRestCreator struct{}

type KubernetesTestCreator struct {
	testClient *fake.Clientset
}

var logger = logger_wrapper.InitLogger("k8s")

func (k *KubernetesRestCreator) GetNewClient(creds K8sClusterCredentials) (kubernetes.Interface, error) {
	var err error
	var config *rest.Config
	if creds.Server == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = getKubernetesConfig(creds)
	}
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func getKubernetesConfig(creds K8sClusterCredentials) (*rest.Config, error) {
	sslActive, parseError := strconv.ParseBool(cfenv.CurrentEnv()["KUBE_SSL_ACTIVE"])
	if parseError != nil {
		logger.Error("KUBE_SSL_ACTIVE env probably not set!")
		return nil, parseError
	}

	var transport *http.Transport
	var err error

	if sslActive {
		_, transport, err = brokerHttp.GetHttpClientWithCertAndCa(creds.AdminCert, creds.AdminKey, creds.CaCert)
	} else {
		_, transport, err = brokerHttp.GetHttpClientWithBasicAuth()
	}

	if err != nil {
		return nil, err
	}

	return &rest.Config{
		Host:      creds.Server,
		Username:  creds.Username,
		Password:  creds.Password,
		Transport: transport,
	}, nil
}

func (k *KubernetesTestCreator) GetNewClient(creds K8sClusterCredentials) (kubernetes.Interface, error) {
	return k.testClient, nil
}

/*
	Objects will be returned in provided order
	All objects should do same action e.g. list/update/create
*/
func (k *KubernetesTestCreator) LoadSimpleResponsesWithSameAction(responseObjects ...runtime.Object) {
	k.testClient = fake.NewSimpleClientset(responseObjects...)
}

func (k *KubernetesTestCreator) LoadErrorResponse(resourceName string) {
	k.testClient.PrependReactor("*", "*", func(action testcore.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, k8sErrors.NewForbidden(apiv1.Resource(resourceName), "", nil)
	})
}
