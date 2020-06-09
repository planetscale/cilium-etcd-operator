// Copyright 2018-2019 Authors of Cilium
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

package cilium_etcd_cluster

import (
	"github.com/cilium/cilium-etcd-operator/pkg/defaults"

	"github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CiliumEtcdCluster returns a Cilium ETCD cluster on the given namespace
// for the given etcd version with for the given size.
func CiliumEtcdCluster(namespace, repository, version string, size int, etcdEnv []v1.EnvVar, nodeSelector map[string]string, busyboxImage string) *v1beta2.EtcdCluster {
	var etcdNodeSelector map[string]string
	if len(nodeSelector) != 0 {
		etcdNodeSelector = nodeSelector
	}
	ciliumEtcdCluster := &v1beta2.EtcdCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaults.ClusterName,
			Namespace: namespace,
			Labels:    defaults.CiliumLabelsApp,
		},
		Spec: v1beta2.ClusterSpec{
			Size:       size,
			Repository: repository,
			Version:    version,
			TLS: &v1beta2.TLSPolicy{
				Static: &v1beta2.StaticTLS{
					Member: &v1beta2.MemberSecret{
						PeerSecret:   defaults.CiliumEtcdPeerTLS,
						ServerSecret: defaults.CiliumEtcdServerTLS,
					},
					OperatorSecret: defaults.CiliumEtcdClientTLS,
				},
			},
			Pod: &v1beta2.PodPolicy{
				EtcdEnv:      etcdEnv,
				Labels:       defaults.CiliumLabelsApp,
				BusyboxImage: busyboxImage,
				NodeSelector: etcdNodeSelector,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(1<<30, resource.BinarySI),
					},
				},
				Affinity: &v1.Affinity{
					NodeAffinity: &v1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
							NodeSelectorTerms: []v1.NodeSelectorTerm{
								MatchExpressions: []v1.NodeSelectorRequirement{
									Key:      "aws.amazon.com/lifecycle",
									Operator: v1.NodeSelectorOpNotIn,
									Values:   []string{"spot"},
								},
								MatchExpressions: []v1.NodeSelectorRequirement{
									Key:      "cloud.google.com/gke-preemptible",
									Operator: v1.NodeSelectorOpNotIn,
									Values:   []string{"true"},
								},
							},
						},
					},
					PodAntiAffinity: &v1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
							// Try to spread the across Nodes if possible.
							{
								Weight: 2,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"etcd_cluster": defaults.ClusterName,
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
							// Try to spread across zones if possible.
							{
								// Weight zone spreading as less important than node spreading.
								Weight: 1,
								PodAffinityTerm: v1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"etcd_cluster": defaults.ClusterName,
										},
									},
									TopologyKey: "failure-domain.beta.kubernetes.io/zone",
								},
							},
						},
					},
				},
			},
		},
	}

	return ciliumEtcdCluster
}
