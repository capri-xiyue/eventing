/*
Copyright 2019 The Knative Authors

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

package v1alpha1

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
)

var (
	validEmptyFilter         = &TriggerFilter{}
	validSourceAndTypeFilter = &TriggerFilter{
		DeprecatedSourceAndType: &TriggerFilterSourceAndType{
			Type:   "other_type",
			Source: "other_source",
		},
	}
	validAttributesFilter = &TriggerFilter{
		Attributes: &TriggerFilterAttributes{
			"type":   "other_type",
			"source": "other_source",
		},
	}
	validSubscriber = &apisv1alpha1.Destination{
		Ref: &corev1.ObjectReference{
			Name:       "subscriber_test",
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
	}
	invalidSubscriber = &apisv1alpha1.Destination{
		Ref: &corev1.ObjectReference{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
	}
	validDependencyAnnotation   = "{\"kind\":\"CronJobSource\",\"name\":\"test-cronjob-source\",\"apiVersion\":\"sources.eventing.knative.dev/v1alpha1\"}"
	invalidDependencyAnnotation = "invalid dependency annotation"
	dependencyAnnotationPath    = fmt.Sprintf("metadata.annotations[%s]", DependencyAnnotation)
)

func TestTriggerValidation(t *testing.T) {
	tests := []struct {
		name string
		t    *Trigger
		want *apis.FieldError
	}{{
		name: "invalid trigger spec",
		t:    &Trigger{Spec: TriggerSpec{}},
		want: &apis.FieldError{
			Paths:   []string{"spec.broker", "spec.filter", "spec.subscriber"},
			Message: "missing field(s)",
		},
	}, {
		name: "invalid dependency annotation, not a corev1.ObjectReference",
		t: &Trigger{
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{
					DependencyAnnotation: invalidDependencyAnnotation,
				}},
			Spec: TriggerSpec{
				Broker:     "test_broker",
				Filter:     validEmptyFilter,
				Subscriber: validSubscriber,
			}},
		want: &apis.FieldError{
			Paths:   []string{dependencyAnnotationPath},
			Message: "The provided annotation was not a corev1.ObjectReference: \"invalid dependency annotation\"",
			Details: "invalid character 'i' looking for beginning of value",
		},
	}, {
		name: "invalid dependency annotation, trigger namespace is not equal to dependency namespace)",
		t: &Trigger{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "test-ns-1",
				Annotations: map[string]string{
					DependencyAnnotation: "{\"kind\":\"CronJobSource\",\"namespace\":\"test-ns-2\", \"name\":\"test-cronjob-source\",\"apiVersion\":\"sources.eventing.knative.dev/v1alpha1\"}",
				}},
			Spec: TriggerSpec{
				Broker:     "test_broker",
				Filter:     validEmptyFilter,
				Subscriber: validSubscriber,
			}},
		want: &apis.FieldError{
			Paths:   []string{dependencyAnnotationPath + "." + "namespace"},
			Message: "Namespace must be empty or equal to the trigger namespace \"test-ns-1\"",
		},
	},
		{
			name: "invalid dependency annotation, missing kind)",
			t: &Trigger{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "test-ns",
					Annotations: map[string]string{
						DependencyAnnotation: "{\"name\":\"test-cronjob-source\",\"apiVersion\":\"sources.eventing.knative.dev/v1alpha1\"}",
					}},
				Spec: TriggerSpec{
					Broker:     "test_broker",
					Filter:     validEmptyFilter,
					Subscriber: validSubscriber,
				}},
			want: &apis.FieldError{
				Paths:   []string{dependencyAnnotationPath + "." + "kind"},
				Message: "missing field(s)",
			},
		}, {
			name: "invalid dependency annotation, missing name",
			t: &Trigger{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "test-ns",
					Annotations: map[string]string{
						DependencyAnnotation: "{\"kind\":\"CronJobSource\",\"apiVersion\":\"sources.eventing.knative.dev/v1alpha1\"}",
					}},
				Spec: TriggerSpec{
					Broker:     "test_broker",
					Filter:     validEmptyFilter,
					Subscriber: validSubscriber,
				}},
			want: &apis.FieldError{
				Paths:   []string{dependencyAnnotationPath + "." + "name"},
				Message: "missing field(s)",
			},
		}, {
			name: "invalid dependency annotation, missing apiVersion",
			t: &Trigger{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "test-ns",
					Annotations: map[string]string{
						DependencyAnnotation: "{\"kind\":\"CronJobSource\",\"name\":\"test-cronjob-source\"}",
					}},
				Spec: TriggerSpec{
					Broker:     "test_broker",
					Filter:     validEmptyFilter,
					Subscriber: validSubscriber,
				}},
			want: &apis.FieldError{
				Paths:   []string{dependencyAnnotationPath + "." + "apiVersion"},
				Message: "missing field(s)",
			},
		}, {
			name: "invalid dependency annotation, missing kind, name, apiVersion",
			t: &Trigger{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "test-ns",
					Annotations: map[string]string{
						DependencyAnnotation: "{}",
					}},
				Spec: TriggerSpec{
					Broker:     "test_broker",
					Filter:     validEmptyFilter,
					Subscriber: validSubscriber,
				}},
			want: &apis.FieldError{
				Paths: []string{
					dependencyAnnotationPath + "." + "kind",
					dependencyAnnotationPath + "." + "name",
					dependencyAnnotationPath + "." + "apiVersion"},
				Message: "missing field(s)",
			},
		},
		{
			name: "invalid trigger spec, invalid dependency annotation(missing kind, name, apiVersion)",
			t: &Trigger{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "test-ns",
					Annotations: map[string]string{
						DependencyAnnotation: "{}",
					}},
				Spec: TriggerSpec{}},
			want: &apis.FieldError{
				Paths: []string{
					"spec.broker", "spec.filter", "spec.subscriber",
					dependencyAnnotationPath + "." + "kind",
					dependencyAnnotationPath + "." + "name",
					dependencyAnnotationPath + "." + "apiVersion"},
				Message: "missing field(s)",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.t.Validate(context.TODO())
			if diff := cmp.Diff(test.want.Error(), got.Error()); diff != "" {
				t.Errorf("Trigger.Validate (-want, +got) = %v", diff)
			}
		})
	}
}

func TestTriggerSpecValidation(t *testing.T) {
	tests := []struct {
		name string
		ts   *TriggerSpec
		want *apis.FieldError
	}{{
		name: "invalid trigger spec",
		ts:   &TriggerSpec{},
		want: func() *apis.FieldError {
			fe := apis.ErrMissingField("broker", "filter", "subscriber")
			return fe
		}(),
	}, {
		name: "missing broker",
		ts: &TriggerSpec{
			Broker:     "",
			Filter:     validSourceAndTypeFilter,
			Subscriber: validSubscriber,
		},
		want: func() *apis.FieldError {
			fe := apis.ErrMissingField("broker")
			return fe
		}(),
	}, {
		name: "missing filter",
		ts: &TriggerSpec{
			Broker:     "test_broker",
			Subscriber: validSubscriber,
		},
		want: func() *apis.FieldError {
			fe := apis.ErrMissingField("filter")
			return fe
		}(),
	}, {
		name: "missing attributes keys",
		ts: &TriggerSpec{
			Broker: "test_broker",
			Filter: &TriggerFilter{
				Attributes: &TriggerFilterAttributes{},
			},
			Subscriber: validSubscriber,
		},
		want: &apis.FieldError{
			Message: "At least one filtered attribute must be specified",
			Paths:   []string{"filter.attributes"},
		},
	}, {
		name: "invalid attribute name, start with number",
		ts: &TriggerSpec{
			Broker: "test_broker",
			Filter: &TriggerFilter{
				Attributes: &TriggerFilterAttributes{
					"0invalid": "my-value",
				},
			},
			Subscriber: validSubscriber,
		},
		want: &apis.FieldError{
			Message: "Invalid attribute name: \"0invalid\"",
			Paths:   []string{"filter.attributes"},
		},
	}, {
		name: "invalid attribute name, capital letters",
		ts: &TriggerSpec{
			Broker: "test_broker",
			Filter: &TriggerFilter{
				Attributes: &TriggerFilterAttributes{
					"invALID": "my-value",
				},
			},
			Subscriber: validSubscriber,
		},
		want: &apis.FieldError{
			Message: "Invalid attribute name: \"invALID\"",
			Paths:   []string{"filter.attributes"},
		},
	}, {
		name: "multiple oneof sourceAndType and attributes",
		ts: &TriggerSpec{
			Broker: "test_broker",
			Filter: &TriggerFilter{
				DeprecatedSourceAndType: &TriggerFilterSourceAndType{
					Type:   "other_type",
					Source: "other_source",
				},
				Attributes: &TriggerFilterAttributes{
					"type": "other_type",
				},
			},
			Subscriber: validSubscriber,
		},
		want: func() *apis.FieldError {
			fe := apis.ErrMultipleOneOf("filter.sourceAndType", "filter.attributes")
			return fe
		}(),
	}, {
		name: "missing subscriber",
		ts: &TriggerSpec{
			Broker: "test_broker",
			Filter: validSourceAndTypeFilter,
		},
		want: func() *apis.FieldError {
			fe := apis.ErrMissingField("subscriber")
			return fe
		}(),
	}, {
		name: "missing subscriber.ref.name",
		ts: &TriggerSpec{
			Broker:     "test_broker",
			Filter:     validSourceAndTypeFilter,
			Subscriber: invalidSubscriber,
		},
		want: func() *apis.FieldError {
			fe := apis.ErrMissingField("subscriber.ref.name")
			return fe
		}(),
	}, {
		name: "missing broker",
		ts: &TriggerSpec{
			Broker:     "",
			Filter:     validSourceAndTypeFilter,
			Subscriber: validSubscriber,
		},
		want: func() *apis.FieldError {
			fe := apis.ErrMissingField("broker")
			return fe
		}(),
	}, {
		name: "valid empty filter",
		ts: &TriggerSpec{
			Broker:     "test_broker",
			Filter:     validEmptyFilter,
			Subscriber: validSubscriber,
		},
		want: &apis.FieldError{},
	}, {
		name: "valid SourceAndType filter",
		ts: &TriggerSpec{
			Broker:     "test_broker",
			Filter:     validSourceAndTypeFilter,
			Subscriber: validSubscriber,
		},
		want: &apis.FieldError{},
	}, {
		name: "valid Attributes filter",
		ts: &TriggerSpec{
			Broker:     "test_broker",
			Filter:     validAttributesFilter,
			Subscriber: validSubscriber,
		},
		want: &apis.FieldError{},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.ts.Validate(context.TODO())
			if diff := cmp.Diff(test.want.Error(), got.Error()); diff != "" {
				t.Errorf("%s: Validate TriggerSpec (-want, +got) = %v", test.name, diff)
			}
		})
	}
}

func TestTriggerImmutableFields(t *testing.T) {
	tests := []struct {
		name     string
		current  apis.Immutable
		original apis.Immutable
		want     *apis.FieldError
	}{{
		name: "good (no change)",
		current: &Trigger{
			Spec: TriggerSpec{
				Broker: "broker",
			},
		},
		original: &Trigger{
			Spec: TriggerSpec{
				Broker: "broker",
			},
		},
		want: nil,
	}, {
		name: "new nil is ok",
		current: &Trigger{
			Spec: TriggerSpec{
				Broker: "broker",
			},
		},
		original: nil,
		want:     nil,
	}, {
		name: "invalid type",
		current: &Trigger{
			Spec: TriggerSpec{
				Broker: "broker",
			},
		},
		original: &Broker{},
		want: &apis.FieldError{
			Message: "The provided original was not a Trigger",
		},
	}, {
		name: "good (filter change)",
		current: &Trigger{
			Spec: TriggerSpec{
				Broker: "broker",
			},
		},
		original: &Trigger{
			Spec: TriggerSpec{
				Broker: "broker",
				Filter: validSourceAndTypeFilter,
			},
		},
		want: nil,
	}, {
		name: "bad (broker change)",
		current: &Trigger{
			Spec: TriggerSpec{
				Broker: "broker",
			},
		},
		original: &Trigger{
			Spec: TriggerSpec{
				Broker: "original_broker",
			},
		},
		want: &apis.FieldError{
			Message: "Immutable fields changed (-old +new)",
			Paths:   []string{"spec", "broker"},
			Details: `{string}:
	-: "original_broker"
	+: "broker"
`,
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.current.CheckImmutableFields(context.TODO(), test.original)
			if diff := cmp.Diff(test.want.Error(), got.Error()); diff != "" {
				t.Errorf("CheckImmutableFields (-want, +got) = %v", diff)
			}
		})
	}
}


func TestTriggerSpecValidationTest(t *testing.T) {
	tests := []struct {
		name string
		ts   *TriggerSpec
		want *apis.FieldError
	}{{
		name: "missing subscriber.ref.name",
		ts: &TriggerSpec{
			Broker:     "test_broker",
			Filter:     validSourceAndTypeFilter,
			Subscriber: invalidSubscriber,
		},
		want: func() *apis.FieldError {
			fe := apis.ErrMissingField("subscriber.ref.name")
			return fe
		}(),
	},}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.ts.Validate(context.TODO())
			if diff := cmp.Diff(test.want.Error(), got.Error()); diff != "" {
				t.Errorf("%s: Validate TriggerSpec (-want, +got) = %v", test.name, diff)
			}
		})
	}
}
