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
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/trustedanalytics/kubernetes-broker/catalog"
	"github.com/trustedanalytics/kubernetes-broker/state"
	tst "github.com/trustedanalytics/kubernetes-broker/test"
)

func prepareMocksAndRouter(t *testing.T) (fabricator *K8Fabricator, mockStateService *state.MockStateService,
	mockKubernetesRest *KubernetesTestCreator) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStateService = state.NewMockStateService(mockCtrl)

	mockKubernetesRest = &KubernetesTestCreator{}
	fabricator = &K8Fabricator{mockKubernetesRest}
	return
}

const serviceId = "mockKubernetesRest"
const orgHost = "orgHost"
const space = "spaceTest"

var testCreds K8sClusterCredentials = K8sClusterCredentials{Server: orgHost}

func TestFabricateService(t *testing.T) {
	fabricator, mockStateService, mockKubernetesRest := prepareMocksAndRouter(t)

	blueprint := &catalog.KubernetesComponent{
		Deployments: []*appsv1beta1.Deployment{&appsv1beta1.Deployment{Spec: appsv1beta1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{{}},
			}}}},
		},
		Services:               []*apiv1.Service{&apiv1.Service{}},
		ServiceAccounts:        []*apiv1.ServiceAccount{&apiv1.ServiceAccount{}},
		Secrets:                []*apiv1.Secret{&apiv1.Secret{}},
		PersistentVolumeClaims: []*apiv1.PersistentVolumeClaim{&apiv1.PersistentVolumeClaim{}},
	}

	secretResponse := &apiv1.SecretList{
		Items: []apiv1.Secret{{}},
	}
	pvmResponse := &apiv1.PersistentVolumeClaimList{
		Items: []apiv1.PersistentVolumeClaim{{}},
	}
	deploymentResponse := &appsv1beta1.DeploymentList{
		Items: []appsv1beta1.Deployment{{}},
	}
	serviceResponse := &apiv1.ServiceList{
		Items: []apiv1.Service{{}},
	}
	serviceAccountResponse := &apiv1.ServiceAccountList{
		Items: []apiv1.ServiceAccount{{}},
	}

	Convey("Test FabricateService", t, func() {
		Convey("Should returns proper response", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction(secretResponse, pvmResponse, serviceResponse, serviceAccountResponse, deploymentResponse)
			gomock.InOrder(
				mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SECRETS", nil),
				mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SECRET0", nil),
				mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIMS", nil),
				mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIM0", nil),
				mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_DEPLOYMENTS", nil),
				mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_DEPLOYMENT0", nil),
				mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SVCS", nil),
				mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SVC0", nil),
				mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_ACCS", nil),
				mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_ACC0", nil),
				mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_FAB_OK", nil),
			)
			result, err := fabricator.FabricateService(testCreds, space, serviceId, `{"name": "param"}`, mockStateService, blueprint)

			So(err, ShouldBeNil)
			So(result.Url, ShouldEqual, "")
		})

		// Convey("Should returns error on Create Secret fail ", func() {
		// 	mockKubernetesRest.LoadErrorResponse("Secret")
		// 	mockKubernetesRest.LoadSimpleResponsesWithSameAction()

		// 	gomock.InOrder(
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SECRETS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SECRET0", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "FAILED", gomock.Any()),
		// 	)
		// 	_, err := fabricator.FabricateService(testCreds, space, serviceId, "", mockStateService, blueprint)

		// 	So(err, ShouldNotBeNil)
		// })

		// Convey("Should returns error on Create Deployments fail ", func() {
		// 	mockKubernetesRest.LoadErrorResponse("Deployment")
		// 	mockKubernetesRest.LoadSimpleResponsesWithSameAction(secretResponse, pvmResponse)
		// 	gomock.InOrder(
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SECRETS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SECRET0", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIMS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIM0", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_DEPLOYMENTS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, gomock.Any(), nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "FAILED", gomock.Any()),
		// 	)
		// 	_, err := fabricator.FabricateService(testCreds, space, serviceId, "", mockStateService, blueprint)

		// 	So(err, ShouldNotBeNil)
		// })

		// Convey("Should returns error on Create Service fail ", func() {
		// 	mockKubernetesRest.LoadErrorResponse("Service")
		// 	mockKubernetesRest.LoadSimpleResponsesWithSameAction(secretResponse, pvmResponse, deploymentResponse)
		// 	gomock.InOrder(
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SECRETS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SECRET0", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIMS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIM0", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_DEPLOYMENTS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, gomock.Any(), nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SVCS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, gomock.Any(), nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "FAILED", gomock.Any()),
		// 	)
		// 	_, err := fabricator.FabricateService(testCreds, space, serviceId, "", mockStateService, blueprint)

		// 	So(err, ShouldNotBeNil)
		// })

		// Convey("Should returns error on Create AccountService fail ", func() {
		// 	mockKubernetesRest.LoadErrorResponse("ServiceAccount")
		// 	mockKubernetesRest.LoadSimpleResponsesWithSameAction(secretResponse, pvmResponse, serviceResponse, deploymentResponse)
		// 	gomock.InOrder(
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SECRETS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SECRET0", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIMS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_PERSIST_VOL_CLAIM0", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_DEPLOYMENTS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, gomock.Any(), nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_SVCS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, gomock.Any(), nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "IN_PROGRESS_CREATING_ACCS", nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, gomock.Any(), nil),
		// 		mockStateService.EXPECT().ReportProgress(serviceId, "FAILED", gomock.Any()),
		// 	)
		// 	_, err := fabricator.FabricateService(testCreds, space, serviceId, "", mockStateService, blueprint)

		// 	So(err, ShouldNotBeNil)
		// })
		Convey("Should returns error when extra paramaters are wrong", func() {
			_, err := fabricator.FabricateService(testCreds, space, serviceId, `BAD_PARAMETER`, mockStateService, blueprint)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestCheckKubernetesServiceHealthByServiceInstanceId(t *testing.T) {
	fabricator, _, mockKubernetesRest := prepareMocksAndRouter(t)

	Convey("Test CheckKubernetesServiceHealthByServiceInstanceId", t, func() {
		Convey("Should returns proper response", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction()

			response, err := fabricator.CheckKubernetesServiceHealthByServiceInstanceId(testCreds, space, serviceId)
			So(err, ShouldBeNil)
			So(response, ShouldBeTrue)
		})

		Convey("Should error on pending containers", func() {
			pods := apiv1.PodList{
				Items: []apiv1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Labels: map[string]string{
							"managed_by":   "TAP",
							serviceIdLabel: serviceId,
						},
					},
					Status: apiv1.PodStatus{
						ContainerStatuses: []apiv1.ContainerStatus{
							apiv1.ContainerStatus{Name: "running", Ready: true},
							apiv1.ContainerStatus{Name: "waiting", Ready: false},
						},
					}},
				},
			}
			mockKubernetesRest.LoadSimpleResponsesWithSameAction(&pods)

			response, err := fabricator.CheckKubernetesServiceHealthByServiceInstanceId(testCreds, space, serviceId)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "waiting")
			So(response, ShouldBeFalse)
		})

		// TODO: Fails on nil pointer exception on returning error in FakePods.List
		/*Convey("Should returns error on Get pods fail", func() {
			mockKubernetesRest.Init(getErrorResponseForSpecificResource("PodList"))

			response, err := fabricator.CheckKubernetesServiceHealthByServiceInstanceId(testCreds, space, serviceId)

			So(err, ShouldNotBeNil)
			So(response, ShouldBeFalse)
		})*/
	})
}

func TestDeleteAllByServiceId(t *testing.T) {
	fabricator, _, mockKubernetesRest := prepareMocksAndRouter(t)

	Convey("Test DeleteAllByServiceId", t, func() {
		Convey("Should returns proper response", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction()

			err := fabricator.DeleteAllByServiceId(testCreds, serviceId)
			So(err, ShouldBeNil)
		})

		Convey("Should returns error on List ServiceAccounts fail", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction()
			mockKubernetesRest.LoadErrorResponse("ServiceAccount")

			err := fabricator.DeleteAllByServiceId(testCreds, serviceId)
			So(err, ShouldNotBeNil)
		})

		Convey("Should returns error on List Services fail", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction()
			mockKubernetesRest.LoadErrorResponse("Service")

			err := fabricator.DeleteAllByServiceId(testCreds, serviceId)
			So(err, ShouldNotBeNil)
		})

		Convey("Should returns error on List Secret fail", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction()
			mockKubernetesRest.LoadErrorResponse("Secret")

			err := fabricator.DeleteAllByServiceId(testCreds, serviceId)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestGetAllPodsEnvsByServiceId(t *testing.T) {
	fabricator, _, mockKubernetesRest := prepareMocksAndRouter(t)
	deployments := appsv1beta1.DeploymentList{
		Items: []appsv1beta1.Deployment{{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Labels: map[string]string{
					"managed_by":   "TAP",
					serviceIdLabel: serviceId,
				},
			},
			Spec: appsv1beta1.DeploymentSpec{
				Template: apiv1.PodTemplateSpec{
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{{
							Env: []apiv1.EnvVar{{
								Name: "secret-env",
								ValueFrom: &apiv1.EnvVarSource{
									SecretKeyRef: &apiv1.SecretKeySelector{
										LocalObjectReference: apiv1.LocalObjectReference{Name: "secret-name"},
										Key:                  "secret-key",
									},
								}},
							}},
						},
					},
				},
			}},
		},
	}

	Convey("Test GetAllPodsEnvsByServiceId", t, func() {
		Convey("Should return error when no items in response", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction()

			_, err := fabricator.GetAllPodsEnvsByServiceId(testCreds, space, serviceId)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "No deployments associated with the service: "+serviceId)
		})

		Convey("Should return data from matching secret", func() {
			secrets := apiv1.SecretList{
				Items: []apiv1.Secret{
					apiv1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "secret-name",
							Namespace: "default",
							Labels: map[string]string{
								"managed_by":   "TAP",
								serviceIdLabel: serviceId,
							},
						},
						Data: map[string][]byte{
							"secret-key": []byte("secret-value"),
						},
					},
				},
			}

			mockKubernetesRest.LoadSimpleResponsesWithSameAction(&deployments, &secrets)

			envs, err := fabricator.GetAllPodsEnvsByServiceId(testCreds, space, serviceId)
			So(err, ShouldBeNil)
			So(envs, ShouldHaveLength, 1)
			So(envs[0].Containers, ShouldHaveLength, 1)
			So(envs[0].Containers[0].Envs["secret-env"], ShouldEqual, "secret-value")
		})

		Convey("Should return empty string when no matching secret", func() {
			secrets := apiv1.SecretList{
				Items: []apiv1.Secret{
					apiv1.Secret{
						ObjectMeta: metav1.ObjectMeta{Name: "another-secret-name"},
						Data: map[string][]byte{
							"secret-key": []byte("secret-value"),
						},
					},
				},
			}

			mockKubernetesRest.LoadSimpleResponsesWithSameAction(&deployments, &secrets)

			envs, err := fabricator.GetAllPodsEnvsByServiceId(testCreds, space, serviceId)
			So(err, ShouldBeNil)
			So(envs, ShouldHaveLength, 1)
			So(envs[0].Containers, ShouldHaveLength, 1)
			So(envs[0].Containers[0].Envs["secret"], ShouldEqual, "")
		})
	})
}

func TestGetSecret(t *testing.T) {
	fabricator, _, mockKubernetesRest := prepareMocksAndRouter(t)

	secret := tst.GetTestSecret()

	Convey("Test GetSecret", t, func() {
		Convey("Should returns proper response", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction(&secret)
			result, err := fabricator.GetSecret(testCreds, tst.TestSecretName)

			So(err, ShouldBeNil)
			So(result.Name, ShouldEqual, tst.TestSecretName)
			So(result.Data[tst.TestSecretName], ShouldResemble, tst.GetTestSecretData())
		})

		Convey("Should returns error on SecretsGet fail", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction()
			mockKubernetesRest.LoadErrorResponse("Secret")
			_, err := fabricator.GetSecret(testCreds, tst.TestSecretName)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestCreateSecret(t *testing.T) {
	fabricator, _, mockKubernetesRest := prepareMocksAndRouter(t)

	secret := tst.GetTestSecret()

	Convey("Test CreateSecret", t, func() {
		Convey("Should returns proper response", func() {
			secret.ObjectMeta.Namespace = ""
			mockKubernetesRest.LoadSimpleResponsesWithSameAction(&secret)
			err := fabricator.CreateSecret(testCreds, secret)

			So(err, ShouldBeNil)
		})

		Convey("Should returns error on SecretsCreate fail", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction()
			mockKubernetesRest.LoadErrorResponse("Secret")
			err := fabricator.CreateSecret(testCreds, secret)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestUpdateSecret(t *testing.T) {
	fabricator, _, mockKubernetesRest := prepareMocksAndRouter(t)

	secret := tst.GetTestSecret()

	Convey("Test UpdateSecret", t, func() {
		Convey("Should returns proper response", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction(&secret)
			err := fabricator.UpdateSecret(testCreds, secret)

			So(err, ShouldBeNil)
		})

		Convey("Should returns error on SecretsGet fail", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction()
			mockKubernetesRest.LoadErrorResponse("Secret")
			err := fabricator.UpdateSecret(testCreds, secret)

			So(err, ShouldNotBeNil)
		})
	})
}

func TestDeleteSecret(t *testing.T) {
	fabricator, _, mockKubernetesRest := prepareMocksAndRouter(t)

	secret := tst.GetTestSecret()

	Convey("Test DeleteSecret", t, func() {
		Convey("Should returns proper response", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction(&secret)
			err := fabricator.DeleteSecret(testCreds, tst.TestSecretName)

			So(err, ShouldBeNil)
		})

		Convey("Should returns error on SecretsGet fail", func() {
			mockKubernetesRest.LoadSimpleResponsesWithSameAction()
			mockKubernetesRest.LoadErrorResponse("Secret")
			err := fabricator.DeleteSecret(testCreds, tst.TestSecretName)

			So(err, ShouldNotBeNil)
		})
	})
}
