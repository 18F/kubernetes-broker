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

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"

	"github.com/trustedanalytics/kubernetes-broker/catalog"
	"github.com/trustedanalytics/kubernetes-broker/state"
)

type KubernetesApi interface {
	FabricateService(creds K8sClusterCredentials, space, cf_service_id, parameters string, ss state.StateService,
		component *catalog.KubernetesComponent) (FabricateResult, error)
	CheckKubernetesServiceHealthByServiceInstanceId(creds K8sClusterCredentials, space, instance_id string) (bool, error)
	DeleteAllByServiceId(creds K8sClusterCredentials, service_id string) error
	DeleteAllPersistentVolumeClaims(creds K8sClusterCredentials) error
	GetAllPersistentVolumes(creds K8sClusterCredentials) ([]apiv1.PersistentVolume, error)
	GetAllPodsEnvsByServiceId(creds K8sClusterCredentials, space, service_id string) ([]PodEnvs, error)
	GetService(creds K8sClusterCredentials, org, serviceId string) ([]apiv1.Service, error)
	GetServices(creds K8sClusterCredentials, org string) ([]apiv1.Service, error)
	GetQuota(creds K8sClusterCredentials, space string) (*apiv1.ResourceQuotaList, error)
	GetClusterWorkers(creds K8sClusterCredentials) ([]string, error)
	GetPodsStateByServiceId(creds K8sClusterCredentials, service_id string) ([]PodStatus, error)
	GetPodsStateForAllServices(creds K8sClusterCredentials) (map[string][]PodStatus, error)
	ListDeployments(creds K8sClusterCredentials) (*appsv1beta1.DeploymentList, error)
	GetSecret(creds K8sClusterCredentials, key string) (*apiv1.Secret, error)
	CreateSecret(creds K8sClusterCredentials, secret apiv1.Secret) error
	DeleteSecret(creds K8sClusterCredentials, key string) error
	UpdateSecret(creds K8sClusterCredentials, secret apiv1.Secret) error
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

	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return result, err
	}

	extraEnvironments := []apiv1.EnvVar{{Name: "TAP_K8S", Value: "true"}}
	if parameters != "" {
		extraUserParam := apiv1.EnvVar{}
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
		_, err = client.CoreV1().Secrets(metav1.NamespaceDefault).Create(sc)
		if err != nil {
			ss.ReportProgress(cf_service_id, "FAILED", err)
			return result, err
		}
	}

	ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIMS", nil)
	for idx, claim := range component.PersistentVolumeClaims {
		ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIM"+strconv.Itoa(idx), nil)
		_, err = client.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault).Create(claim)
		if err != nil {
			ss.ReportProgress(cf_service_id, "FAILED", err)
			return result, err
		}
	}

	ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_STATEFULSETS", nil)
	for idx, statefulSet := range component.StatefulSets {
		ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_STATEFULSET"+strconv.Itoa(idx), nil)
		for i, container := range statefulSet.Spec.Template.Spec.Containers {
			statefulSet.Spec.Template.Spec.Containers[i].Env = append(container.Env, extraEnvironments...)
		}

		_, err = client.AppsV1beta1().StatefulSets(apiv1.NamespaceDefault).Create(statefulSet)
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

		_, err = client.AppsV1beta1().Deployments(apiv1.NamespaceDefault).Create(deployment)
		if err != nil {
			ss.ReportProgress(cf_service_id, "FAILED", err)
			return result, err
		}
	}

	ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_SVCS", nil)
	for idx, svc := range component.Services {
		ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_SVC"+strconv.Itoa(idx), nil)
		_, err = client.CoreV1().Services(apiv1.NamespaceDefault).Create(svc)
		if err != nil {
			ss.ReportProgress(cf_service_id, "FAILED", err)
			return result, err
		}
	}

	ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_ACCS", nil)
	for idx, acc := range component.ServiceAccounts {
		ss.ReportProgress(cf_service_id, "IN_PROGRESS_CREATING_ACC"+strconv.Itoa(idx), nil)
		_, err = client.CoreV1().ServiceAccounts(apiv1.NamespaceDefault).Create(acc)
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
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return err
	}

	for _, jobHook := range jobs {
		if jobHook.Type == jobType {
			_, err = client.BatchV1().Jobs(apiv1.NamespaceDefault).Create(&jobHook.Job)
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
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		logger.Error("getKubernetesClientAndExtensionClient error:", err)
		return err
	}

	selector, err := getSelectorForManagedByLabel()
	if err != nil {
		return err
	}

	jobs, err := client.BatchV1().Jobs(apiv1.NamespaceDefault).List(metav1.ListOptions{
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
			_, err = client.CoreV1().ConfigMaps(apiv1.NamespaceDefault).Create(getConfigMapFromLogs(job, logs))
			if err != nil {
				ss.NotifyCatalog(serviceId, fmt.Sprintf("Can't save Jobs credentials! Job name: %s, serviceId: %s. Logs: %v", job.Name, serviceId, logs), err)
			} else {
				ss.NotifyCatalog(serviceId, fmt.Sprintf("Job with name: %s and serviceId: %s SAVED SUCCESSFULLY in ConfMap!", job.Name, serviceId), err)
			}
		}
		err = client.BatchV1().Jobs(apiv1.NamespaceDefault).Delete(job.Name, &metav1.DeleteOptions{})
		if err != nil {
			ss.NotifyCatalog(serviceId, fmt.Sprintf("Delete Job ERROR! Job name: %s, serviceId: %s", job.Name, serviceId), err)
		} else {
			ss.NotifyCatalog(serviceId, fmt.Sprintf("Job name: %s and serviceId: %s DELETED SUCCESSFULLY!", job.Name, serviceId), err)
		}
	}
	return nil
}

func getPodsLogs(client kubernetes.Interface, selector string) (map[string]string, error) {
	result := map[string]string{}
	pods, err := client.CoreV1().Pods(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		byteBody, err := client.CoreV1().Pods(apiv1.NamespaceDefault).GetLogs(pod.Name, &apiv1.PodLogOptions{}).Do().Raw()
		if err != nil {
			return nil, err
		}
		result[pod.Name] = string(byteBody)
	}
	return result, nil
}

func getConfigMapFromLogs(job batchv1.Job, logs map[string]string) *apiv1.ConfigMap {
	return &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: job.Name, Labels: job.Labels},
		Data:       logs,
	}
}

func (k *K8Fabricator) CheckKubernetesServiceHealthByServiceInstanceId(creds K8sClusterCredentials, space, instance_id string) (bool, error) {
	logger.Info("[CheckKubernetesServiceHealthByServiceInstanceId] serviceId:", instance_id)
	// http://kubernetes.io/v1.1/docs/user-guide/liveness/README.html

	client, selector, err := k.getKubernetesClientWithServiceIdSelector(creds, instance_id)
	if err != nil {
		return false, err
	}

	sets, err := client.AppsV1beta1().StatefulSets(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return false, err
	}

	// if we have statefulsets as part of this deployment then we check their readiness first and bail if not
	if len(sets.Items) > 0 {
		logger.Debug("[CheckKubernetesServiceHealthByServiceInstanceId] Sets:", sets)

		for _, set := range sets.Items {
			logger.Debug("[CheckKubernetesServiceHealthByServiceInstanceId] Set Replicas Want/Have:", set.Spec.ServiceName, *set.Spec.Replicas, set.Status.Replicas)
			if *set.Spec.Replicas != set.Status.Replicas {
				logger.Debug("[CheckKubernetesServiceHealthByServiceInstanceId] Not all replicas are up.")
				return false, fmt.Errorf("Not all replicas are up. Want %s, Have %s", set.Spec.Replicas, set.Status.Replicas)
			}
		}
	}

	pods, err := client.CoreV1().Pods(apiv1.NamespaceDefault).List(metav1.ListOptions{
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
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return err
	}
	selector, err := getSelectorForServiceIdLabel(service_id)
	if err != nil {
		return err
	}

	accs, err := client.CoreV1().ServiceAccounts(apiv1.NamespaceDefault).List(metav1.ListOptions{
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
		err = client.CoreV1().ServiceAccounts(apiv1.NamespaceDefault).Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			logger.Error("[DeleteAllByServiceId] Delete service account failed:", err)
			return err
		}
	}

	svcs, err := client.CoreV1().Services(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[DeleteAllByServiceId] List services failed:", err)
		return err
	}

	for _, i := range svcs.Items {
		name = i.ObjectMeta.Name
		logger.Debug("[DeleteAllByServiceId] Delete service:", name)
		err = client.CoreV1().Services(apiv1.NamespaceDefault).Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			logger.Error("[DeleteAllByServiceId] Delete service failed:", err)
			return err
		}
	}

	statefulSets, err := client.AppsV1beta1().StatefulSets(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[DeleteAllByServiceId] List statefulsets failed:", err)
		return err
	}
	for _, statefulSet := range statefulSets.Items {
		name = statefulSet.ObjectMeta.Name
		logger.Debug("[DeleteAllByServiceId] Delete statefulset:", name)

		foreground := metav1.DeletePropagationForeground
		err = client.AppsV1beta1().StatefulSets(apiv1.NamespaceDefault).Delete(name, &metav1.DeleteOptions{
			PropagationPolicy: &foreground,
		})
		if err != nil {
			logger.Error("[DeleteAllByServiceId] Delete statefulset failed:", err)
			return err
		}
	}

	dpls, err := client.AppsV1beta1().Deployments(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[DeleteAllByServiceId] List deployments failed:", err)
		return err
	}
	for _, i := range dpls.Items {
		name = i.ObjectMeta.Name
		logger.Debug("[DeleteAllByServiceId] Delete deployment:", name)

		foreground := metav1.DeletePropagationForeground
		err = client.AppsV1beta1().Deployments(apiv1.NamespaceDefault).Delete(name, &metav1.DeleteOptions{
			PropagationPolicy: &foreground,
		})
		if err != nil {
			logger.Error("[DeleteAllByServiceId] Delete deployment failed:", err)
			return err
		}
	}

	secrets, err := client.CoreV1().Secrets(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[DeleteAllByServiceId] List secrets failed:", err)
		return err
	}

	for _, i := range secrets.Items {
		name = i.ObjectMeta.Name
		logger.Debug("[DeleteAllByServiceId] Delete secret:", name)
		err = client.CoreV1().Secrets(apiv1.NamespaceDefault).Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			logger.Error("[DeleteAllByServiceId] Delete secret failed:", err)
			return err
		}
	}

	pvcs, err := client.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[DeleteAllByServiceId] List PersistentVolumeClaims failed:", err)
		return err
	}

	for _, i := range pvcs.Items {
		name = i.ObjectMeta.Name
		logger.Debug("[DeleteAllByServiceId] Delete PersistentVolumeClaims:", name)
		err = client.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault).Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			logger.Error("[DeleteAllByServiceId] Delete PersistentVolumeClaims failed:", err)
			return err
		}
	}

	return nil
}

func (k *K8Fabricator) DeleteAllPersistentVolumeClaims(creds K8sClusterCredentials) error {
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return err
	}

	pvList, err := client.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: labels.NewSelector().String(),
	})
	if err != nil {
		logger.Error("[DeleteAllPersistentVolumeClaims] List PersistentVolumeClaims failed:", err)
		return err
	}

	var errorFound bool = false
	for _, i := range pvList.Items {
		name := i.ObjectMeta.Name
		logger.Debug("[DeleteAllPersistentVolumeClaims] Delete PersistentVolumeClaims:", name)
		err = client.CoreV1().PersistentVolumeClaims(apiv1.NamespaceDefault).Delete(name, &metav1.DeleteOptions{})
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

func (k *K8Fabricator) GetAllPersistentVolumes(creds K8sClusterCredentials) ([]apiv1.PersistentVolume, error) {
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return nil, err
	}

	pvList, err := client.CoreV1().PersistentVolumes().List(metav1.ListOptions{
		LabelSelector: labels.NewSelector().String(),
	})
	if err != nil {
		logger.Error("[GetAllPersistentVolumes] List PersistentVolume failed:", err)
		return nil, err
	}
	return pvList.Items, nil
}

func (k *K8Fabricator) GetService(creds K8sClusterCredentials, org, serviceId string) ([]apiv1.Service, error) {
	logger.Info("[GetService] orgId:", org)
	response := []apiv1.Service{}

	client, selector, err := k.getKubernetesClientWithServiceIdSelector(creds, serviceId)
	if err != nil {
		return response, err
	}

	serviceList, err := client.CoreV1().Services(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[GetService] ListServices failed:", err)
		return response, err
	}

	return serviceList.Items, nil
}

func (k *K8Fabricator) GetServices(creds K8sClusterCredentials, org string) ([]apiv1.Service, error) {
	logger.Info("[GetServices] orgId:", org)
	response := []apiv1.Service{}

	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		logger.Error("[GetServices] GetNewClient error", err)
		return response, err
	}
	selector, err := getSelectorForManagedByLabel()
	if err != nil {
		logger.Error("[GetServices] GetSelectorForManagedByLabel error", err)
		return response, err
	}

	serviceList, err := client.CoreV1().Services(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error("[GetServices] ListServices failed:", err)
		return response, err
	}
	return serviceList.Items, nil
}

func (k *K8Fabricator) GetQuota(creds K8sClusterCredentials, space string) (*apiv1.ResourceQuotaList, error) {
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return nil, err
	}

	return client.CoreV1().ResourceQuotas(apiv1.NamespaceDefault).List(metav1.ListOptions{})
}

func (k *K8Fabricator) GetClusterWorkers(creds K8sClusterCredentials) ([]string, error) {
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return []string{}, err
	}

	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
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

func (k *K8Fabricator) ListDeployments(creds K8sClusterCredentials) (*appsv1beta1.DeploymentList, error) {
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return nil, err
	}
	selector, err := getSelectorForManagedByLabel()
	if err != nil {
		return nil, err
	}

	return client.AppsV1beta1().Deployments(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
}

type PodStatus struct {
	PodName       string
	ServiceId     string
	Status        apiv1.PodPhase
	StatusMessage string
}

func (k *K8Fabricator) GetPodsStateByServiceId(creds K8sClusterCredentials, service_id string) ([]PodStatus, error) {
	result := []PodStatus{}

	client, selector, err := k.getKubernetesClientWithServiceIdSelector(creds, service_id)
	if err != nil {
		return result, err
	}

	pods, err := client.CoreV1().Pods(apiv1.NamespaceDefault).List(metav1.ListOptions{
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

	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return result, err
	}
	selector, err := getSelectorForManagedByLabel()
	if err != nil {
		return result, err
	}

	pods, err := client.CoreV1().Pods(apiv1.NamespaceDefault).List(metav1.ListOptions{
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
	Ports []apiv1.ServicePort
}

func (k *K8Fabricator) GetSecret(creds K8sClusterCredentials, key string) (*apiv1.Secret, error) {
	secret := &apiv1.Secret{}
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return secret, err
	}
	result, err := client.CoreV1().Secrets(apiv1.NamespaceDefault).Get(key, metav1.GetOptions{})
	if err != nil {
		return secret, err
	}
	return result, nil
}

func (k *K8Fabricator) CreateSecret(creds K8sClusterCredentials, secret apiv1.Secret) error {
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return err
	}
	_, err = client.CoreV1().Secrets(apiv1.NamespaceDefault).Create(&secret)
	if err != nil {
		return err
	}
	return nil
}

func (k *K8Fabricator) DeleteSecret(creds K8sClusterCredentials, key string) error {
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return err
	}
	err = client.CoreV1().Secrets(apiv1.NamespaceDefault).Delete(key, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (k *K8Fabricator) UpdateSecret(creds K8sClusterCredentials, secret apiv1.Secret) error {
	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return err
	}
	_, err = client.CoreV1().Secrets(apiv1.NamespaceDefault).Update(&secret)
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

	client, err := k.KubernetesClient.GetNewClient(creds)
	if err != nil {
		return result, err
	}
	selector, err := getSelectorForServiceIdLabel(service_id)
	if err != nil {
		return result, err
	}

	deployments, err := client.AppsV1beta1().Deployments(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return result, err
	}
	sets, err := client.AppsV1beta1().StatefulSets(apiv1.NamespaceDefault).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return result, err
	}
	if len(deployments.Items) == 0 && len(sets.Items) == 0{
		return result, fmt.Errorf("No deployments or statefulsets associated with the service: %s", service_id)
	}

	secrets, err := client.CoreV1().Secrets(apiv1.NamespaceDefault).List(metav1.ListOptions{
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

	for _, set := range sets.Items {
		pod := PodEnvs{}
		pod.DeploymentName = set.Name
		pod.Containers = []ContainerSimple{}

		for _, container := range set.Spec.Template.Spec.Containers {
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

func findSecretValue(secrets *apiv1.SecretList, selector *apiv1.SecretKeySelector) string {
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

func (k *K8Fabricator) getKubernetesClientWithServiceIdSelector(creds K8sClusterCredentials, serviceId string) (kubernetes.Interface, string, error) {
	selector, err := getSelectorForServiceIdLabel(serviceId)
	if err != nil {
		return nil, "", err
	}

	client, err := k.KubernetesClient.GetNewClient(creds)
	return client, selector, err
}

func getSelectorForServiceIdLabel(serviceId string) (string, error) {
	selector := labels.NewSelector()
	managedByReq, err := labels.NewRequirement(managedByLabel, selection.Equals, []string{"TAP"})
	if err != nil {
		return "", err
	}
	serviceIdReq, err := labels.NewRequirement(serviceIdLabel, selection.Equals, []string{serviceId})
	if err != nil {
		return "", err
	}
	return selector.Add(*managedByReq, *serviceIdReq).String(), nil
}

func getSelectorForManagedByLabel() (string, error) {
	selector := labels.NewSelector()
	managedByReq, err := labels.NewRequirement(managedByLabel, selection.Equals, []string{"TAP"})
	if err != nil {
		return "", err
	}
	return selector.Add(*managedByReq).String(), nil
}
