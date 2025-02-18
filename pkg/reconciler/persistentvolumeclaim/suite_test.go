/*
Copyright The CloudNativePG Contributors

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

package persistentvolumeclaim

import (
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cloudnative-pg/cloudnative-pg/pkg/specs"
	"github.com/cloudnative-pg/cloudnative-pg/pkg/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpecs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Specification properties")
}

func makePVC(clusterName, serial string, role utils.PVCRole, isResizing bool) corev1.PersistentVolumeClaim {
	annotations := map[string]string{
		specs.ClusterSerialAnnotationName: serial,
		StatusAnnotationName:              StatusReady,
	}

	var conditions []corev1.PersistentVolumeClaimCondition
	if isResizing {
		conditions = append(conditions, corev1.PersistentVolumeClaimCondition{
			Type:   corev1.PersistentVolumeClaimResizing,
			Status: corev1.ConditionTrue,
		})
	}

	labels := map[string]string{
		utils.PvcRoleLabelName: string(role),
	}

	if role == "" {
		labels = nil
	}

	return corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        clusterName + "-" + serial,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.PersistentVolumeClaimSpec{},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase:      corev1.ClaimBound,
			Conditions: conditions,
		},
	}
}

func makePod(clusterName, serial string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName + "-" + serial,
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: clusterName + "-" + serial,
						},
					},
				},
			},
		},
	}
}

func makeJob(clusterName, serial string) batchv1.Job {
	return batchv1.Job{
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: clusterName + "-" + serial,
								},
							},
						},
					},
				},
			},
		},
	}
}
