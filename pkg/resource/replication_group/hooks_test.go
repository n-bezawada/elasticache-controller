// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package replication_group

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/elasticache-controller/apis/v1alpha1"
)

func newDurabilityResource(durability *string, annotations map[string]string) *resource {
	return &resource{
		ko: &svcapitypes.ReplicationGroup{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: annotations,
			},
			Spec: svcapitypes.ReplicationGroupSpec{
				Durability: durability,
			},
		},
	}
}

func Test_lastRequestedDurability(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		want        *string
	}{
		{
			name:        "no annotations",
			annotations: nil,
			want:        nil,
		},
		{
			name:        "annotation not set",
			annotations: map[string]string{"some-other-annotation": "value"},
			want:        nil,
		},
		{
			name:        "annotation set",
			annotations: map[string]string{AnnotationLastRequestedDurability: "sync"},
			want:        aws.String("sync"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desired := newDurabilityResource(nil, tt.annotations)
			got := lastRequestedDurability(desired)
			if (got == nil) != (tt.want == nil) {
				t.Fatalf("lastRequestedDurability() = %v, want %v", got, tt.want)
			}
			if got != nil && *got != *tt.want {
				t.Errorf("lastRequestedDurability() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func Test_durabilityRequiresUpdate(t *testing.T) {
	tests := []struct {
		name              string
		desiredDurability *string
		annotations       map[string]string
		want              bool
	}{
		{
			name:              "no last-requested annotation, no desired durability",
			desiredDurability: nil,
			annotations:       nil,
			want:              false,
		},
		{
			name:              "no last-requested annotation, desired durability set",
			desiredDurability: aws.String("sync"),
			annotations:       nil,
			want:              true,
		},
		{
			name:              "last-requested matches desired",
			desiredDurability: aws.String("sync"),
			annotations:       map[string]string{AnnotationLastRequestedDurability: "sync"},
			want:              false,
		},
		{
			name:              "last-requested differs from desired",
			desiredDurability: aws.String("sync"),
			annotations:       map[string]string{AnnotationLastRequestedDurability: "async"},
			want:              true,
		},
		{
			// AWS offers no way to unset durability, so a nil desired value means the
			// field is unmanaged and no request should be made.
			name:              "last-requested set, desired durability cleared",
			desiredDurability: nil,
			annotations:       map[string]string{AnnotationLastRequestedDurability: "async"},
			want:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desired := newDurabilityResource(tt.desiredDurability, tt.annotations)
			if got := durabilityRequiresUpdate(desired); got != tt.want {
				t.Errorf("durabilityRequiresUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setLastRequestedDurability(t *testing.T) {
	tests := []struct {
		name              string
		durability        *string
		wantAnnotationSet bool
	}{
		{
			name:              "durability set",
			durability:        aws.String("sync"),
			wantAnnotationSet: true,
		},
		{
			name:              "durability nil",
			durability:        nil,
			wantAnnotationSet: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newDurabilityResource(tt.durability, nil)
			annotations := map[string]string{}
			rm := &resourceManager{}
			rm.setLastRequestedDurability(r, annotations)

			val, ok := annotations[AnnotationLastRequestedDurability]
			if ok != tt.wantAnnotationSet {
				t.Fatalf("annotation present = %v, want %v", ok, tt.wantAnnotationSet)
			}
			if ok && val != *tt.durability {
				t.Errorf("annotation value = %v, want %v", val, *tt.durability)
			}
		})
	}
}
