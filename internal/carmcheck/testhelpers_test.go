// Copyright 2026 kropath Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package carmcheck

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var iamRoleSelectorGVK = schema.GroupVersionKind{
	Group:   "services.k8s.aws",
	Version: "v1alpha1",
	Kind:    "IAMRoleSelector",
}

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	sch := runtime.NewScheme()
	if err := corev1.AddToScheme(sch); err != nil {
		t.Fatalf("AddToScheme corev1: %v", err)
	}
	if err := appsv1.AddToScheme(sch); err != nil {
		t.Fatalf("AddToScheme appsv1: %v", err)
	}
	// Register the ACK IAMRoleSelector CRD as an unstructured type so the
	// fake client can List/Get it without kropath-aws-controller vendoring
	// ACK's IAM API types.
	sch.AddKnownTypeWithName(iamRoleSelectorGVK, &unstructured.Unstructured{})
	sch.AddKnownTypeWithName(iamRoleSelectorListGVK, &unstructured.UnstructuredList{})
	metav1.AddToGroupVersion(sch, schema.GroupVersion{Group: "services.k8s.aws", Version: "v1alpha1"})
	return sch
}

func testClient(t *testing.T, objs ...runtime.Object) *fake.ClientBuilder {
	t.Helper()
	return fake.NewClientBuilder().WithScheme(testScheme(t)).WithRuntimeObjects(objs...)
}

// fakeClientWithoutIAMRoleSelector builds a client from a scheme that does
// not know the IAMRoleSelector GVK at all, reproducing what a real cluster's
// RESTMapper returns when the CRD is not installed.
func fakeClientWithoutIAMRoleSelector(t *testing.T, _ *runtime.Scheme) client.Client {
	t.Helper()
	sch := runtime.NewScheme()
	if err := corev1.AddToScheme(sch); err != nil {
		t.Fatalf("AddToScheme corev1: %v", err)
	}
	if err := appsv1.AddToScheme(sch); err != nil {
		t.Fatalf("AddToScheme appsv1: %v", err)
	}
	return fake.NewClientBuilder().WithScheme(sch).Build()
}

func namespace(name string, annotations, labels map[string]string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: annotations, Labels: labels},
	}
}

func testDeployment(namespace, name string, args []string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name: "controller",
						Args: args,
					}},
				},
			},
		},
	}
}

func testConfigMap(namespace, name string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Data:       data,
	}
}

func testIAMRoleSelector(name string, spec map[string]interface{}) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(iamRoleSelectorGVK)
	u.SetName(name)
	if err := unstructured.SetNestedMap(u.Object, spec, "spec"); err != nil {
		panic(err)
	}
	return u
}
