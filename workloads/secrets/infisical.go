package secrets

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewInfisicalChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Namespace
	cdk8s.NewApiObject(chart, jsii.String("infisical-namespace"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Namespace"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String(namespace),
		},
	})

	// infisical-secrets is created by: infra/scripts/create-bootstrap-secrets.sh
	// Keys: DB_PASSWORD, AUTH_SECRET, ENCRYPTION_KEY, DB_CONNECTION_URI, REDIS_PASSWORD

	// Infisical Standalone Helm Values
	values := map[string]any{
		"infisical": map[string]any{
			"kubeSecretRef": "infisical-secrets",
			"replicaCount":  1,
			"resources": map[string]any{
				"requests": map[string]any{"cpu": "200m", "memory": "512Mi"},
				"limits":   map[string]any{"memory": "1024Mi"},
			},
		},
		"postgresql": map[string]any{
			"enabled": false, // external StatefulSet below
			"useExistingPostgresSecret": map[string]any{
				"enabled": true,
				"existingConnectionStringSecret": map[string]any{
					"name": "infisical-secrets",
					"key":  "DB_CONNECTION_URI",
				},
			},
		},
		"redis": map[string]any{
			"enabled": true,
			// Chart builds REDIS_URL from redis.auth.password directly — no secret-ref support.
			// Both Redis and infisical use the chart default password so they match.
		},
		"ingress": map[string]any{
			"enabled":  false,
			"hostname": "infisical.madhan.app",
		},
		// Disable bundled nginx subchart — even with ingress disabled it renders a
		// ValidatingWebhookConfiguration that can block all pod creation cluster-wide.
		"ingress-nginx": map[string]any{"enabled": false},
	}

	// Deploy Infisical Application
	cdk8s.NewHelm(chart, jsii.String("infisical-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("infisical-standalone"),
		Repo:        jsii.String("https://dl.cloudsmith.io/public/infisical/helm-charts/helm/charts"),
		Version:     jsii.String("1.7.2"),
		ReleaseName: jsii.String("infisical"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// PostgreSQL Service — official image (Bitnami removed from Docker Hub)
	cdk8s.NewApiObject(chart, jsii.String("postgresql-service"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Service"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("postgresql"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"selector": map[string]any{"app": "postgresql"},
		"ports": []map[string]any{
			{"port": 5432, "targetPort": 5432, "name": "postgres"},
		},
	}))

	// PostgreSQL StatefulSet
	cdk8s.NewApiObject(chart, jsii.String("postgresql-statefulset"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("apps/v1"),
		Kind:       jsii.String("StatefulSet"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("postgresql"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"serviceName": "postgresql",
		"replicas":    1,
		"selector": map[string]any{
			"matchLabels": map[string]any{"app": "postgresql"},
		},
		"template": map[string]any{
			"metadata": map[string]any{
				"labels": map[string]any{"app": "postgresql"},
			},
			"spec": map[string]any{
				"containers": []map[string]any{
					{
						"name":  "postgresql",
						"image": "docker.io/library/postgres:17",
						"ports": []map[string]any{
							{"containerPort": 5432, "name": "postgres"},
						},
						"env": []map[string]any{
							{"name": "POSTGRES_USER", "value": "infisical"},
							{"name": "POSTGRES_DB", "value": "infisical"},
							{
								"name": "POSTGRES_PASSWORD",
								"valueFrom": map[string]any{
									"secretKeyRef": map[string]any{
										"name": "infisical-secrets",
										"key":  "DB_PASSWORD",
									},
								},
							},
							{"name": "PGDATA", "value": "/var/lib/postgresql/data/pgdata"},
						},
						"resources": map[string]any{
							"requests": map[string]any{"cpu": "100m", "memory": "256Mi"},
							"limits":   map[string]any{"memory": "512Mi"},
						},
						"volumeMounts": []map[string]any{
							{"name": "data", "mountPath": "/var/lib/postgresql/data"},
							{"name": "run", "mountPath": "/var/run/postgresql"},
						},
					},
				},
				"volumes": []map[string]any{
					{"name": "run", "emptyDir": map[string]any{}},
				},
			},
		},
		"volumeClaimTemplates": []map[string]any{
			{
				"metadata": map[string]any{"name": "data"},
				"spec": map[string]any{
					"accessModes":      []string{"ReadWriteOnce"},
					"storageClassName": "longhorn",
					"resources": map[string]any{
						"requests": map[string]any{"storage": "10Gi"},
					},
				},
			},
		},
	}))

	// HTTPRoute — infisical.madhan.app + infisical.local → infisical:8080
	cdk8s.NewApiObject(chart, jsii.String("infisical-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("infisical"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"infisical.madhan.app", "infisical.local"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "infisical-infisical-standalone-infisical", "port": 8080, "weight": 1},
				},
			},
		},
	}))

	// Infisical Operator
	// Bug in v0.10.25: leader-election RoleBinding has subjects[0].namespace="default" — operator
	// cannot acquire the lease and never reconciles. Overridden explicitly below.
	cdk8s.NewHelm(chart, jsii.String("infisical-operator-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("secrets-operator"),
		Repo:        jsii.String("https://dl.cloudsmith.io/public/infisical/helm-charts/helm/charts/"),
		Version:     jsii.String("0.10.25"),
		ReleaseName: jsii.String("infisical-operator"),
		Values: &map[string]any{
			"controllerManager": map[string]any{
				"manager": map[string]any{
					"resources": map[string]any{
						"limits":   map[string]any{"cpu": "500m", "memory": "128Mi"},
						"requests": map[string]any{"cpu": "10m", "memory": "64Mi"},
					},
				},
			},
		},
	})

	// Fix leader-election RoleBinding — correct subject namespace (chart bug workaround)
	leaderElectionBinding := cdk8s.NewApiObject(chart, jsii.String("infisical-opera-leader-election-rolebinding-fix"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("rbac.authorization.k8s.io/v1"),
		Kind:       jsii.String("RoleBinding"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("infisical-opera-leader-election-rolebinding"), // must match chart name
			Namespace: jsii.String(namespace),
			Labels: &map[string]*string{
				"app.kubernetes.io/instance": jsii.String("infisical-operator"),
				"app.kubernetes.io/name":     jsii.String("secrets-operator"),
			},
		},
	})
	leaderElectionBinding.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/roleRef"), map[string]any{
		"apiGroup": "rbac.authorization.k8s.io",
		"kind":     "Role",
		"name":     "infisical-opera-leader-election-role",
	}))
	leaderElectionBinding.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/subjects"), []map[string]any{
		{"kind": "ServiceAccount", "name": "infisical-opera-controller-manager", "namespace": namespace},
	}))

	// Fix manager ClusterRoleBinding — same chart bug: subjects[0].namespace="default"
	// Without this, operator cannot list secrets/CRDs at cluster scope and crashes immediately after acquiring the lease.
	managerBinding := cdk8s.NewApiObject(chart, jsii.String("infisical-opera-manager-rolebinding-fix"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("rbac.authorization.k8s.io/v1"),
		Kind:       jsii.String("ClusterRoleBinding"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("infisical-opera-manager-rolebinding"), // must match chart name
			Labels: &map[string]*string{
				"app.kubernetes.io/instance": jsii.String("infisical-operator"),
				"app.kubernetes.io/name":     jsii.String("secrets-operator"),
			},
		},
	})
	managerBinding.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/roleRef"), map[string]any{
		"apiGroup": "rbac.authorization.k8s.io",
		"kind":     "ClusterRole",
		"name":     "infisical-opera-manager-role",
	}))
	managerBinding.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/subjects"), []map[string]any{
		{"kind": "ServiceAccount", "name": "infisical-opera-controller-manager", "namespace": namespace},
	}))

	// Fix metrics-auth ClusterRoleBinding — same chart bug
	metricsBinding := cdk8s.NewApiObject(chart, jsii.String("infisical-opera-metrics-auth-rolebinding-fix"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("rbac.authorization.k8s.io/v1"),
		Kind:       jsii.String("ClusterRoleBinding"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("infisical-opera-metrics-auth-rolebinding"), // must match chart name
			Labels: &map[string]*string{
				"app.kubernetes.io/instance": jsii.String("infisical-operator"),
				"app.kubernetes.io/name":     jsii.String("secrets-operator"),
			},
		},
	})
	metricsBinding.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/roleRef"), map[string]any{
		"apiGroup": "rbac.authorization.k8s.io",
		"kind":     "ClusterRole",
		"name":     "infisical-opera-metrics-auth-role",
	}))
	metricsBinding.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/subjects"), []map[string]any{
		{"kind": "ServiceAccount", "name": "infisical-opera-controller-manager", "namespace": namespace},
	}))

	// Kubernetes Auth — ClusterRole: grants tokenreviews create for JWT verification
	cdk8s.NewApiObject(chart, jsii.String("infisical-token-reviewer-role"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("rbac.authorization.k8s.io/v1"),
		Kind:       jsii.String("ClusterRole"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("infisical-token-reviewer"),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/rules"), []map[string]any{
		{"apiGroups": []string{"authentication.k8s.io"}, "resources": []string{"tokenreviews"}, "verbs": []string{"create"}},
	}))

	// ClusterRoleBinding — binds infisical-token-reviewer to operator SA
	tokenReviewerBinding := cdk8s.NewApiObject(chart, jsii.String("infisical-token-reviewer-binding"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("rbac.authorization.k8s.io/v1"),
		Kind:       jsii.String("ClusterRoleBinding"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("infisical-token-reviewer"),
		},
	})
	tokenReviewerBinding.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/roleRef"), map[string]any{
		"apiGroup": "rbac.authorization.k8s.io",
		"kind":     "ClusterRole",
		"name":     "infisical-token-reviewer",
	}))
	tokenReviewerBinding.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/subjects"), []map[string]any{
		{"kind": "ServiceAccount", "name": "infisical-operator-controller-manager", "namespace": namespace},
	}))

	// InfisicalSecret CR — kubernetesAuth (operator uses its own SA JWT, no stored credentials)
	// TODO: replace identityId after creating Machine Identity in Infisical UI → Access Control →
	// Machine Identities → k8s-homelab → Kubernetes Auth (host: https://192.168.1.210:6443)
	cdk8s.NewApiObject(chart, jsii.String("infisical-bootstrap-secret-cr"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets.infisical.com/v1alpha1"),
		Kind:       jsii.String("InfisicalSecret"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("infisical-bootstrap-secret"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"hostAPI":        "https://infisical.madhan.app/api",
		"resyncInterval": 60,
		"authentication": map[string]any{
			"kubernetesAuth": map[string]any{
				"identityId": "REPLACE_WITH_IDENTITY_ID", // TODO: set after Infisical UI setup
				"serviceAccountRef": map[string]any{
					"name":      "infisical-operator-controller-manager",
					"namespace": namespace,
				},
			},
		},
		"managedSecretReference": map[string]any{
			"secretName":      "infisical-synced-secrets",
			"secretNamespace": namespace,
			"creationPolicy":  "Orphan",
		},
	}))

	return chart
}
