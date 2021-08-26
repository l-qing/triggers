/*
Copyright 2021 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"os"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCustomObject(t *testing.T) {
	err := os.Setenv("METRICS_PROMETHEUS_PORT", "9000")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Setenv("SYSTEM_NAMESPACE", "tekton-pipelines")
	if err != nil {
		t.Fatal(err)
	}

	config := *MakeConfig()
	metadata := map[string]interface{}{
		"creationTimestamp": nil,
		"labels": map[string]interface{}{
			"app.kubernetes.io/managed-by": "EventListener",
			"app.kubernetes.io/part-of":    "Triggers",
			"eventlistener":                eventListenerName,
		},
		"namespace": namespace,
		"ownerReferences": []interface{}{
			map[string]interface{}{
				"apiVersion":         "triggers.tekton.dev/v1beta1",
				"blockOwnerDeletion": true,
				"controller":         true,
				"kind":               "EventListener",
				"name":               eventListenerName,
				"uid":                "",
			},
		},
	}
	args := []interface{}{
		"--el-name=" + eventListenerName,
		"--el-namespace=" + namespace,
		"--port=" + strconv.Itoa(eventListenerContainerPort),
		"--readtimeout=" + strconv.FormatInt(DefaultReadTimeout, 10),
		"--writetimeout=" + strconv.FormatInt(DefaultWriteTimeout, 10),
		"--idletimeout=" + strconv.FormatInt(DefaultIdleTimeout, 10),
		"--timeouthandler=" + strconv.FormatInt(DefaultTimeOutHandler, 10),
		"--is-multi-ns=" + strconv.FormatBool(false),
		"--payload-validation=" + strconv.FormatBool(true),
	}

	env := []interface{}{
		map[string]interface{}{
			"name":  "SYSTEM_NAMESPACE",
			"value": namespace,
		},
		map[string]interface{}{
			"name":  "CONFIG_OBSERVABILITY_NAME",
			"value": "config-observability",
		},
		map[string]interface{}{
			"name":  "METRICS_DOMAIN",
			"value": TriggersMetricsDomain,
		},
		map[string]interface{}{
			"name":  "METRICS_PROMETHEUS_PORT",
			"value": "9000",
		},
	}

	tests := []struct {
		name string
		el   *v1beta1.EventListener
		want *unstructured.Unstructured
	}{{
		name: "vanilla",
		el: makeEL(func(el *v1beta1.EventListener) {
			el.Spec.Resources.CustomResource = &v1beta1.CustomResource{
				RawExtension: runtime.RawExtension{
					Raw: []byte(`{
						"apiVersion": "serving.knative.dev/v1",
						"kind": "Service"
					}`),
				},
			}
		}),
		want: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "serving.knative.dev/v1",
				"kind":       "Service",
				"metadata":   metadata,
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"creationTimestamp": nil,
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "event-listener",
									"image": DefaultImage,
									"args":  args,
									"env":   env,
									"ports": []interface{}{
										map[string]interface{}{
											"containerPort": int64(8080),
											"protocol":      "TCP",
										},
									},
									"resources": map[string]interface{}{},
									"volumeMounts": []interface{}{
										map[string]interface{}{
											"mountPath": "/etc/config-logging",
											"name":      "config-logging",
											"readOnly":  true,
										},
									},
									"readinessProbe": map[string]interface{}{
										"httpGet": map[string]interface{}{
											"path":   "/live",
											"port":   int64(0),
											"scheme": "HTTP",
										},
										"successThreshold": int64(1),
									},
								},
							},
							"volumes": []interface{}{
								map[string]interface{}{
									"configMap": map[string]interface{}{
										"name": EventListenerConfigMapName,
									},
									"name": "config-logging",
								},
							},
						},
					},
				},
			},
		},
	}, {
		name: "with env vars",
		el: makeEL(func(el *v1beta1.EventListener) {
			el.Spec.Resources.CustomResource = &v1beta1.CustomResource{
				RawExtension: runtime.RawExtension{
					Raw: []byte(`{
						"apiVersion": "serving.knative.dev/v1",
						"kind": "Service",
						"spec": {
							"template": {
								"spec": {
									"containers": [{
										"env": [{
											"name": "FOO",
											"value": "bar"
										}]
									}]
								}
							}
						}
					}`),
				},
			}
		}),
		want: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "serving.knative.dev/v1",
				"kind":       "Service",
				"metadata":   metadata,
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"creationTimestamp": nil,
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "event-listener",
									"image": DefaultImage,
									"args":  args,
									"env": append([]interface{}{
										map[string]interface{}{
											"name":  "FOO",
											"value": "bar",
										},
									}, env...),
									"ports": []interface{}{
										map[string]interface{}{
											"containerPort": int64(8080),
											"protocol":      "TCP",
										},
									},
									"resources": map[string]interface{}{},
									"volumeMounts": []interface{}{
										map[string]interface{}{
											"mountPath": "/etc/config-logging",
											"name":      "config-logging",
											"readOnly":  true,
										},
									},
									"readinessProbe": map[string]interface{}{
										"httpGet": map[string]interface{}{
											"path":   "/live",
											"port":   int64(0),
											"scheme": "HTTP",
										},
										"successThreshold": int64(1),
									},
								},
							},
							"volumes": []interface{}{
								map[string]interface{}{
									"configMap": map[string]interface{}{
										"name": EventListenerConfigMapName,
									},
									"name": "config-logging",
								},
							},
						},
					},
				},
			},
		}}, {
		name: "with resources",
		el: makeEL(func(el *v1beta1.EventListener) {
			el.Spec.Resources.CustomResource = &v1beta1.CustomResource{
				RawExtension: runtime.RawExtension{
					Raw: []byte(`{
							"apiVersion": "serving.knative.dev/v1",
							"kind": "Service",
							"spec": {
								"template": {
									"spec": {
										"containers": [{
											"resources": {
												"limits": {
													"cpu": "101m"
												}
											}
										}]
									}
								}
							}
						}`),
				},
			}
		}),
		want: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "serving.knative.dev/v1",
				"kind":       "Service",
				"metadata":   metadata,
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"creationTimestamp": nil,
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "event-listener",
									"image": DefaultImage,
									"args":  args,
									"env":   env,
									"ports": []interface{}{
										map[string]interface{}{
											"containerPort": int64(8080),
											"protocol":      "TCP",
										},
									},
									"resources": map[string]interface{}{
										"limits": map[string]interface{}{
											"cpu": "101m",
										},
									},
									"volumeMounts": []interface{}{
										map[string]interface{}{
											"mountPath": "/etc/config-logging",
											"name":      "config-logging",
											"readOnly":  true,
										},
									},
									"readinessProbe": map[string]interface{}{
										"httpGet": map[string]interface{}{
											"path":   "/live",
											"port":   int64(0),
											"scheme": "HTTP",
										},
										"successThreshold": int64(1),
									},
								},
							},
							"volumes": []interface{}{
								map[string]interface{}{
									"configMap": map[string]interface{}{
										"name": EventListenerConfigMapName,
									},
									"name": "config-logging",
								},
							},
						},
					},
				},
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := MakeCustomObject(tt.el, config)
			if err != nil {
				t.Fatalf("MakeCustomObject() = %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("MakeCustomObject() did not return expected. -want, +got: %s", diff)
			}
		})
	}
}

func TestCustomObjectError(t *testing.T) {
	err := os.Setenv("METRICS_PROMETHEUS_PORT", "9000")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Setenv("SYSTEM_NAMESPACE", "tekton-pipelines")
	if err != nil {
		t.Fatal(err)
	}

	config := *MakeConfig()

	got, _, err := MakeCustomObject(makeEL(func(el *v1beta1.EventListener) {
		el.Spec.Resources.CustomResource = &v1beta1.CustomResource{
			RawExtension: runtime.RawExtension{
				Raw: []byte(`garbage`),
			},
		}
	}), config)
	if err == nil {
		t.Fatalf("MakeCustomObject() = %v, wanted error", got)
	}
}