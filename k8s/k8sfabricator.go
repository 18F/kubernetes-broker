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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/sets"

	"github.com/trustedanalytics/kubernetes-broker/catalog"
	"github.com/trustedanalytics/kubernetes-broker/state"
)

type KubernetesApi interface {
	FabricateService(creds K8sClusterCredentials, space, cf_service_id, parameters string, ss state.StateService,
		component *catalog.KubernetesComponent) (FabricateResult, error)
	CheckKubernetesServiceHealthByServiceInstanceId(creds K8sClusterCredentials, space, instance_id string) (bool, error)
	DeleteAllByServiceId(creds K8sClusterCredentials, service_id string) error
	DeleteAllPersistentVolumeClaims(creds K8sClusterCredentials) error
	GetAllPersistentVolumes(creds K8sClusterCredentials) ([]api.PersistentVolume, error)
	GetAllPodsEnvsByServiceId(creds K8sClusterCredentials, space, service_id string) ([]PodEnvs, error)
	GetService(creds K8sClusterCredentials, org, serviceId string) ([]api.Service, error)
	GetServices(creds K8sClusterCredentials, org string) ([]api.Service, error)
	GetQuota(creds K8sClusterCredentials, space string) (*api.ResourceQuotaList, error)
	GetClusterWorkers(creds K8sClusterCredentials) ([]string, error)
	GetPodsStateByServiceId(creds K8sClusterCredentials, service_id string) ([]PodStatus, error)
	GetPodsStateForAllServices(creds K8sClusterCredentials) (map[string][]PodStatus, error)
	ListDeployments(creds K8sClusterCredentials) (*extensions.DeploymentList, error)
	GetSecret(creds K8sClusterCredentials, key string) (*api.Secret, error)
	CreateSecret(creds K8sClusterCredentials, secret api.Secret) error
	DeleteSecret(creds K8sClusterCredentials, key string) error
	UpdateSecret(creds K8sClusterCredentials, secret api.Secret) error
	ProcessJobsResult(creds K8sClusterCredentials, ss state.StateService) error
	CreateJobsByType(creds K8sClusterCredentials, jobs []*catalog.JobHook, serviceId string, jobType catalog.JobType, ss state.StateService) error
}

type K8Fabricator struct {
	KubernetesClient KubernetesClientCreator
}

type FabricateResult struct {
	Url string
	Env map[string]string
}

type K8sServiceInfo struct {
	ServiceId string   `json:"serviceId"`
	Org       string   `json:"org"`
	Space     string   `json:"space"`
	Name      string   `json:"name"`
	TapPublic bool     `json:"tapPublic"`
	Uri       []string `json:"uri"`
}

const serviceIdLabel string = "service_id"
const managedByLabel string = "managed_by"

func NewK8Fabricator() *K8Fabricator {
	return &K8Fabricator{KubernetesClient: &KubernetesRestCreator{}}
}

func (k *K8Fabricator) FabricateService(creds K8sClusterCredentials, space, cf_service_id, parameters string,
	ss state.StateService, component *catalog.KubernetesComponent) (FabricateResult, error) {
	result := FabricateResult{"", map[string]string{}}

	client, extensionsClient, err := k.getKubernetesClientAndExtensionClient(creds)
	if err != nil {
		return result, err
	}

	extraEnvironments := []api.EnvVar{{Name: "TAP_K8S", Value: "true"}}
	if parameters != "" {
		extraUserParam := api.EnvVar{}
		err := json.Unmarshal([]byte(parameters), &extraUserParam)
		if err != nil {
			logger.Error("[FabricateService] Unmarshalling extra user parameters error!", err)
			return result, err
		}

		if extraUserParam.Name != "" {
			// kubernetes env name validation:
			// "must be a C identifier (matching regex [A-Za-z_][A-Za-z0-9_]*): e.g. \"my_name\" or \"MyName\"","
			extraUserParam.Name = extraUserParam.Name + "_" + space
			extraUserParam.Name = strings.Replace(extraUserParam.Name, "_", "__", -1) //name_1 --> name__1__SpaceGUID
			extraUserParam.Name = strings.Replace(extraUserParam.Name, "-", "_", -1)  //name-1 --> name_1__SpaceGUID

			extraEnvironments = append(extraEnvironments, extraUserParam)
		}
		logger.Debug("[FabricateService] Extra parameters value:", extraEnvironments)
	}

	ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_SECRETS", nil)
	for idx, sc := range component.Secrets {
		ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_SECRET"+strconv.Itoa(idx), nil)
		_, err = client.Secrets(api.NamespaceDefault).Create(sc)
		if err != nil {
			ss.ReportProgress(cf_service_id, "FAILED", err)
			return result, err
		}
	}

	ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIMS", nil)
	for idx, claim := range component.PersistentVolumeClaims {
		ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIM"+strconv.Itoa(idx), nil)
		_, err = client.PersistentVolumeClaims(api.NamespaceDefault).Create(claim)
		if err != nil {
			ss.ReportProgress(cf_service_id, "FAILED", err)
			return result, err
		}
	}

	ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_DEPLOYMENTS", nil)
	for idx, deployment := range component.Deployments {
		ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_DEPLOYMENT"+strconv.Itoa(idx), nil)
		for i, container := range deployment.Spec.Template.Spec.Containers {
			deployment.Spec.Template.Spec.Containers[i].Env = append(container.Env, extraEnvironments...)
		}

		_, err = extensionsClient.Deployments(api.NamespaceDefault).Create(deployment)
		if err != nil {
			ss.ReportProgress(cf_service_id, "FAILED", err)
			return result, err
		}
	}

	ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_SVCS", nil)
	for idx, svc := range component.Services {
		ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_SVC"+strconv.Itoa(idx), nil)
		_, err = client.Services(api.NamespaceDefault).Create(svc)
		if err != nil {
			ss.ReportProgress(cf_service_id, "FAILED", err)
			return result, err
		}
	}

	ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_ACCS", nil)
	for idx, acc := range component.ServiceAccounts {
		ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_ACC"+strconv.Itoa(idx), nil)
		_, err = client.ServiceAccounts(api.NamespaceDefault).Create(acc)
		if err != nil {
			ss.ReportProgress(cf_service_id, "FAILED", err)
			return result, err
		}
	}

	ss.ReportProgress(cf_service_id, "IN_PROGRESS_FAB_OK", nil)
	return result, nil
}

func (k *K8Fabricator) CreateJobsByType(creds K8sClusterCredentials, jobs []*catalog.JobHook, serviceId string,
	jobType catalog.JobType, ss state.StateService) error {
	c, err := k.KubernetesClient.GetNewExtensionsClient(creds)
	if err != nil {
		return err
	}

	for _, jobHook := range jobs {
		if jobHook.Type == jobType {
			_, err = c.Jobs(api.NamespaceDefault).Create(&jobHook.Job)
			if err != nil {
				ss.NotifyCatalog(serviceId, fmt.Sprintf("Create job error! Job type: %s", jobHook.Type), err)
				return err
			}
			ss.NotifyCatalog(serviceId, fmt.Sprintf("Job created! Job type: %s", jobHook.Type), nil)
		}
	}
	return nil
}

func (k *K8Fabricator) ProcessJobsResult(creds K8sClusterCredentials, ss state.StateService) error {
	logger.Debug("Processing jobs...")
	client, extensionsClient, err := k.getKubernetesClientAndExtensionClient(creds)
	if err != nil {
		logger.Error("getKubernetesClientAndExtensionClient error:", err)
		return err
	}

	selector, err := getSelectorForManagedByLabel()
	if err != nil {
		return err
	}

	jobs, err := extensionsClient.Jobs(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("get Jobs error:", err)
		return err
	}

jobs:
	for _, job := range jobs.Items {
		logger.Info("Processing job: ", job.Name)
		serviceId := job.Labels[serviceIdLabel]
		secretSelector, err := getSelectorForServiceIdLabel(serviceId)

		if job.Status.Active > 0 {
			logger.Info(fmt.Sprintf("Job with name: %s and serviceId: %s is still running. Results will be collected on next attempt", job.Name, serviceId))
			continue jobs
		}

		logs, err := getPodsLogs(client, secretSelector)
		if err != nil {
			ss.NotifyCatalog(serviceId, fmt.Sprintf("Can't get Job logs from pod! Job name: %s, serviceId: %s", job.Name, serviceId), err)
			continue jobs
		}

		if job.Status.Failed > 0 {
			ss.NotifyCatalog(serviceId, fmt.Sprintf("Job with name: %s and serviceId: %s FAILED! Pod logs:", job.Name, serviceId), err)
		}
		if job.Status.Succeeded > 0 && job.Annotations["createConfigMap"] == "true" {
			_, err = client.ConfigMaps(api.NamespaceDefault).Create(getConfigMapFromLogs(job, logs))
			if err != nil {
				ss.NotifyCatalog(serviceId, fmt.Sprintf("Can't save Jobs credentials! Job name: %s, serviceId: %s. Logs: %v", job.Name, serviceId, logs), err)
			} else {
				ss.NotifyCatalog(serviceId, fmt.Sprintf("Job with name: %s and serviceId: %s SAVED SUCCESSFULLY in ConfMap!", job.Name, serviceId), err)
			}
		}
		err = extensionsClient.Jobs(api.NamespaceDefault).Delete(job.Name, &api.DeleteOptions{})
		if err != nil {
			ss.NotifyCatalog(serviceId, fmt.Sprintf("Delete Job ERROR! Job name: %s, serviceId: %s", job.Name, serviceId), err)
		} else {
			ss.NotifyCatalog(serviceId, fmt.Sprintf("Job name: %s and serviceId: %s DELETED SUCCESSFULLY!", job.Name, serviceId), err)
		}
	}
	return nil
}

func getPodsLogs(client KubernetesClient, selector labels.Selector) (map[string]string, error) {
	result := map[string]string{}
	pods, err := client.Pods(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		byteBody, err := client.Pods(api.NamespaceDefault).GetLogs(pod.Name, &api.PodLogOptions{}).Do().Raw()
		if err != nil {
			return nil, err
		}
		result[pod.Name] = string(byteBody)
	}
	return result, nil
}

func getConfigMapFromLogs(job batch.Job, logs map[string]string) *api.ConfigMap {
	return &api.ConfigMap{
		ObjectMeta: api.ObjectMeta{Name: job.Name, Labels: job.Labels},
		Data:       logs,
	}
}

func (k *K8Fabricator) CheckKubernetesServiceHealthByServiceInstanceId(creds K8sClusterCredentials, space, instance_id string) (bool, error) {
	logger.Info("[CheckKubernetesServiceHealthByServiceInstanceId] serviceId:", instance_id)
	// http://kubernetes.io/v1.1/docs/user-guide/liveness/README.html

	c, selector, err := k.getKubernetesClientWithServiceIdSelector(creds, instance_id)
	if err != nil {
		return false, err
	}

	pods, err := c.Pods(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[CheckKubernetesServiceHealthByServiceInstanceId] Getting pods failed:", err)
		return false, err
	}
	logger.Debug("[CheckKubernetesServiceHealthByServiceInstanceId] PODS:", pods)

	// for each pod check if healthy
	// if all healthy return true
	// else return false
	pending := []string{}
	for _, pod := range pods.Items {
		for _, status := range pod.Status.ContainerStatuses {
			if !status.Ready {
				pending = append(pending, status.Name)
			}
		}
	}
	if len(pending) > 0 {
		return false, fmt.Errorf("Pod(s) not ready: %s", strings.Join(pending, ", "))
	}

	return true, nil
}

func (k *K8Fabricator) DeleteAllByServiceId(creds K8sClusterCredentials, service_id string) error {
	logger.Info("[DeleteAllByServiceId] serviceId:", service_id)
	c, extensionClient, err := k.getKubernetesClientAndExtensionClient(creds)
	if err != nil {
		return err
	}
	selector, err := getSelectorForServiceIdLabel(service_id)
	if err != nil {
		return err
	}

	accs, err := c.ServiceAccounts(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[DeleteAllByServiceId] List service accounts failed:", err)
		return err
	}
	var name string
	for _, i := range accs.Items {
		name = i.ObjectMeta.Name
		logger.Debug("[DeleteAllByServiceId] Delete service account:", name)
		err = c.ServiceAccounts(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("[DeleteAllByServiceId] Delete service account failed:", err)
			return err
		}
	}

	svcs, err := c.Services(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[DeleteAllByServiceId] List services failed:", err)
		return err
	}

	for _, i := range svcs.Items {
		name = i.ObjectMeta.Name
		logger.Debug("[DeleteAllByServiceId] Delete service:", name)
		err = c.Services(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("[DeleteAllByServiceId] Delete service failed:", err)
			return err
		}
	}

	if err = NewDeploymentControllerManager(extensionClient).DeleteAll(selector); err != nil {
		logger.Error("[DeleteAllByServiceId] Delete deployment failed:", err)
		return err
	}

	// TODO: Restore after k8s garbage collection is released
	// for _, i := range dpls.Items {
	// 	name = i.ObjectMeta.Name
	// 	logger.Debug("[DeleteAllByServiceId] Delete deployment:", name)

	// 	orphan := false
	// 	opts := api.DeleteOptions{OrphanDependents: &orphan}
	// 	err = c.Deployments(api.NamespaceDefault).Delete(name, &opts)
	// 	if err != nil {
	// 		logger.Error("[DeleteAllByServiceId] Delete deployment failed:", err)
	// 		return err
	// 	}
	// }

	// TODO: Delete after k8s garbage collection is released
	rss, err := extensionClient.ReplicaSets(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[DeleteAllByServiceId] List replica sets failed:", err)
		return err
	}

	for _, i := range rss.Items {
		name = i.ObjectMeta.Name
		logger.Debug("[DeleteAllByServiceId] Delete replica set:", name)

		dpl, err := extensionClient.ReplicaSets(api.NamespaceDefault).Get(name)
		if err != nil {
			return err
		}

		dpl.Spec.Replicas = 0
		_, err = extensionClient.ReplicaSets(api.NamespaceDefault).Update(dpl)
		if err != nil {
			return err
		}

		err = extensionClient.ReplicaSets(api.NamespaceDefault).Delete(name, &api.DeleteOptions{})
		if err != nil {
			logger.Error("[DeleteAllByServiceId] Delete replica set failed:", err)
			return err
		}
	}

	secrets, err := c.Secrets(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[DeleteAllByServiceId] List secrets failed:", err)
		return err
	}

	for _, i := range secrets.Items {
		name = i.ObjectMeta.Name
		logger.Debug("[DeleteAllByServiceId] Delete secret:", name)
		err = c.Secrets(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("[DeleteAllByServiceId] Delete secret failed:", err)
			return err
		}
	}

	pvcs, err := c.PersistentVolumeClaims(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[DeleteAllByServiceId] List PersistentVolumeClaims failed:", err)
		return err
	}

	for _, i := range pvcs.Items {
		name = i.ObjectMeta.Name
		logger.Debug("[DeleteAllByServiceId] Delete PersistentVolumeClaims:", name)
		err = c.PersistentVolumeClaims(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("[DeleteAllByServiceId] Delete PersistentVolumeClaims failed:", err)
			return err
		}
	}

	return nil
}

func (k *K8Fabricator) DeleteAllPersistentVolumeClaims(creds K8sClusterCredentials) error {

	c, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return err
	}

	pvList, err := c.PersistentVolumeClaims(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: labels.NewSelector(),
	})
	if err != nil {
		logger.Error("[DeleteAllPersistentVolumeClaims] List PersistentVolumeClaims failed:", err)
		return err
	}

	var errorFound bool = false
	for _, i := range pvList.Items {
		name := i.ObjectMeta.Name
		logger.Debug("[DeleteAllPersistentVolumeClaims] Delete PersistentVolumeClaims:", name)
		err = c.PersistentVolumeClaims(api.NamespaceDefault).Delete(name)
		if err != nil {
			logger.Error("[DeleteAllPersistentVolumeClaims] Delete PersistentVolumeClaims: "+name+" failed!", err)
			errorFound = true
		}
	}

	if errorFound {
		return errors.New("Error on deleting PersistentVolumeClaims!")
	} else {
		return nil
	}
}

func (k *K8Fabricator) GetAllPersistentVolumes(creds K8sClusterCredentials) ([]api.PersistentVolume, error) {

	c, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return nil, err
	}

	pvList, err := c.PersistentVolumes().List(api.ListOptions{
		LabelSelector: labels.NewSelector(),
	})
	if err != nil {
		logger.Error("[GetAllPersistentVolumes] List PersistentVolume failed:", err)
		return nil, err
	}
	return pvList.Items, nil
}

func (k *K8Fabricator) GetService(creds K8sClusterCredentials, org, serviceId string) ([]api.Service, error) {
	logger.Info("[GetService] orgId:", org)
	response := []api.Service{}

	c, selector, err := k.getKubernetesClientWithServiceIdSelector(creds, serviceId)
	if err != nil {
		return response, err
	}

	serviceList, err := c.Services(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[GetService] ListServices failed:", err)
		return response, err
	}

	return serviceList.Items, nil
}

func (k *K8Fabricator) GetServices(creds K8sClusterCredentials, org string) ([]api.Service, error) {
	logger.Info("[GetServices] orgId:", org)
	response := []api.Service{}

	c, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		logger.Error("[GetServices] GetNewClient error", err)
		return response, err
	}
	selector, err := getSelectorForManagedByLabel()
	if err != nil {
		logger.Error("[GetServices] GetSelectorForManagedByLabel error", err)
		return response, err
	}

	serviceList, err := c.Services(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[GetServices] ListServices failed:", err)
		return response, err
	}
	return serviceList.Items, nil
}

func (k *K8Fabricator) GetQuota(creds K8sClusterCredentials, space string) (*api.ResourceQuotaList, error) {
	c, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return nil, err
	}

	return c.ResourceQuotas(api.NamespaceDefault).List(api.ListOptions{})
}

func (k *K8Fabricator) GetClusterWorkers(creds K8sClusterCredentials) ([]string, error) {
	c, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return []string{}, err
	}

	nodes, err := c.Nodes().List(api.ListOptions{})
	if err != nil {
		logger.Error("[GetClusterWorkers] GetNodes error:", err)
		return []string{}, err
	}

	workers := []string{}
	for _, i := range nodes.Items {
		workers = append(workers, i.Spec.ExternalID)
	}
	return workers, nil
}

func (k *K8Fabricator) ListDeployments(creds K8sClusterCredentials) (*extensions.DeploymentList, error) {
	c, err := k.KubernetesClient.GetNewExtensionsClient(creds)
	if err != nil {
		return nil, err
	}
	selector, err := getSelectorForManagedByLabel()
	if err != nil {
		return nil, err
	}

	return c.Deployments(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
}

type PodStatus struct {
	PodName       string
	ServiceId     string
	Status        api.PodPhase
	StatusMessage string
}

func (k *K8Fabricator) GetPodsStateByServiceId(creds K8sClusterCredentials, service_id string) ([]PodStatus, error) {
	result := []PodStatus{}

	c, selector, err := k.getKubernetesClientWithServiceIdSelector(creds, service_id)
	if err != nil {
		return result, err
	}

	pods, err := c.Pods(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return result, err
	}

	for _, pod := range pods.Items {
		podStatus := PodStatus{
			pod.Name, service_id, pod.Status.Phase, pod.Status.Message,
		}
		result = append(result, podStatus)
	}
	return result, nil
}

func (k *K8Fabricator) GetPodsStateForAllServices(creds K8sClusterCredentials) (map[string][]PodStatus, error) {
	result := map[string][]PodStatus{}

	c, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return result, err
	}
	selector, err := getSelectorForManagedByLabel()
	if err != nil {
		return result, err
	}

	pods, err := c.Pods(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return result, err
	}

	for _, pod := range pods.Items {
		service_id := pod.Labels[serviceIdLabel]
		if service_id != "" {
			podStatus := PodStatus{
				pod.Name, service_id, pod.Status.Phase, pod.Status.Message,
			}
			result[service_id] = append(result[service_id], podStatus)
		}
	}
	return result, nil
}

type ServiceCredential struct {
	Name  string
	Host  string
	Ports []api.ServicePort
}

func (k *K8Fabricator) GetSecret(creds K8sClusterCredentials, key string) (*api.Secret, error) {
	secret := &api.Secret{}
	c, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return secret, err
	}
	result, err := c.Secrets(api.NamespaceDefault).Get(key)
	if err != nil {
		return secret, err
	}
	return result, nil
}

func (k *K8Fabricator) CreateSecret(creds K8sClusterCredentials, secret api.Secret) error {
	c, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return err
	}
	_, err = c.Secrets(api.NamespaceDefault).Create(&secret)
	if err != nil {
		return err
	}
	return nil
}

func (k *K8Fabricator) DeleteSecret(creds K8sClusterCredentials, key string) error {
	c, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return err
	}
	err = c.Secrets(api.NamespaceDefault).Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (k *K8Fabricator) UpdateSecret(creds K8sClusterCredentials, secret api.Secret) error {
	c, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return err
	}
	_, err = c.Secrets(api.NamespaceDefault).Update(&secret)
	if err != nil {
		return err
	}
	return nil
}

type PodEnvs struct {
	DeploymentName string
	Containers     []ContainerSimple
}

type ContainerSimple struct {
	Name string
	Envs map[string]string
}

func (k *K8Fabricator) GetAllPodsEnvsByServiceId(creds K8sClusterCredentials, space, service_id string) ([]PodEnvs, error) {
	logger.Info("[GetAllPodsEnvsByServiceId] serviceId:", service_id)
	result := []PodEnvs{}

	c, extensionClient, err := k.getKubernetesClientAndExtensionClient(creds)
	if err != nil {
		return result, err
	}
	selector, err := getSelectorForServiceIdLabel(service_id)
	if err != nil {
		return result, err
	}

	deployments, err := extensionClient.Deployments(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return result, err
	}
	if len(deployments.Items) == 0 {
		return result, fmt.Errorf("No deployments associated with the service: %s", service_id)
	}

	secrets, err := c.Secrets(api.NamespaceDefault).List(api.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[GetAllPodsEnvsByServiceId] List secrets failed:", err)
		return result, err
	}

	for _, deployment := range deployments.Items {
		pod := PodEnvs{}
		pod.DeploymentName = deployment.Name
		pod.Containers = []ContainerSimple{}

		for _, container := range deployment.Spec.Template.Spec.Containers {
			simpleContainer := ContainerSimple{}
			simpleContainer.Name = container.Name
			simpleContainer.Envs = map[string]string{}

			for _, env := range container.Env {
				if env.Value == "" && env.ValueFrom.SecretKeyRef != nil {
					logger.Debug("Empty env value, searching env variable in secrets")
					simpleContainer.Envs[env.Name] = findSecretValue(secrets, env.ValueFrom.SecretKeyRef)
				} else {
					simpleContainer.Envs[env.Name] = env.Value
				}

			}
			pod.Containers = append(pod.Containers, simpleContainer)
		}
		result = append(result, pod)
	}
	return result, nil
}

func findSecretValue(secrets *api.SecretList, selector *api.SecretKeySelector) string {
	for _, secret := range secrets.Items {
		if secret.Name != selector.Name {
			continue
		}
		for key, value := range secret.Data {
			if key == selector.Key {
				return string((value))
			}
		}
	}
	logger.Info("Secret key not found: ", selector.Key)
	return ""
}

func (k *K8Fabricator) getKubernetesClientWithServiceIdSelector(creds K8sClusterCredentials, serviceId string) (KubernetesClient, labels.Selector, error) {
	selector, err := getSelectorForServiceIdLabel(serviceId)
	if err != nil {
		return nil, selector, err
	}

	c, err := k.KubernetesClient.GetNewClient(creds)
	return c, selector, err
}

func (k *K8Fabricator) getKubernetesClientAndExtensionClient(creds K8sClusterCredentials) (KubernetesClient, ExtensionsInterface, error) {
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return client, nil, err
	}

	extensionsClient, err := k.KubernetesClient.GetNewExtensionsClient(creds)
	return client, extensionsClient, err
}

func getSelectorForServiceIdLabel(serviceId string) (labels.Selector, error) {
	selector := labels.NewSelector()
	managedByReq, err := labels.NewRequirement(managedByLabel, labels.EqualsOperator, sets.NewString("TAP"))
	if err != nil {
		return selector, err
	}
	serviceIdReq, err := labels.NewRequirement(serviceIdLabel, labels.EqualsOperator, sets.NewString(serviceId))
	if err != nil {
		return selector, err
	}
	return selector.Add(*managedByReq, *serviceIdReq), nil
}

func getSelectorForManagedByLabel() (labels.Selector, error) {
	selector := labels.NewSelector()
	managedByReq, err := labels.NewRequirement(managedByLabel, labels.EqualsOperator, sets.NewString("TAP"))
	if err != nil {
		return selector, err
	}
	return selector.Add(*managedByReq), nil
}
