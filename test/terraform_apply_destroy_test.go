package test

import (
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

func getDefaultTerraformOptions(t *testing.T, suffix string) (string, *terraform.Options, error) {

	tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, "..", ".")

	namespace := "test-ns-" + suffix + "-" + strings.ToLower(random.UniqueId())

	terraformOptions := &terraform.Options{
		TerraformDir:       tempTestFolder,
		Vars:               map[string]interface{}{},
		MaxRetries:         5,
		TimeBetweenRetries: 5 * time.Minute,
		NoColor:            true,
		Logger:             logger.TestingT,
	}

	terraformOptions.Vars["name"] = "test-name"
	terraformOptions.Vars["namespace"] = namespace

	return namespace, terraformOptions, nil
}

func TestApplyAndDestroyWithDefaultValues(t *testing.T) {
	t.Parallel()

	namespace, options, err := getDefaultTerraformOptions(t, "single-cnt-with-map")
	assert.NoError(t, err)

	k8sOptions := k8s.NewKubectlOptions("", "", namespace)
	k8s.CreateNamespace(t, k8sOptions, namespace)
	// website::tag::5::Make sure to delete the namespace at the end of the test
	defer k8s.DeleteNamespace(t, k8sOptions, namespace)

	kubeResourcePath := "./resources.yml"
	defer k8s.KubectlDelete(t, k8sOptions, kubeResourcePath)
	k8s.KubectlApply(t, k8sOptions, kubeResourcePath)

	options.Vars["image"] = map[string]interface{}{"test-container": "training/webapp:latest"}

	options.Vars["ports"] = map[string]interface{}{
		"test-container": map[string]interface{}{
			"5000": map[string]interface{}{
				"protocol": "TCP",
			},
		},
	}

	options.Vars["readiness_probes"] = map[string]interface{}{
		"test-container": map[string]interface{}{
			"tcp_socket": map[string]interface{}{
				"port": 5000,
			},
			"type": "tcp_socket",
		},
	}

	options.Vars["liveness_probes"] = map[string]interface{}{
		"test-container": map[string]interface{}{
			"tcp_socket": map[string]interface{}{
				"port": 5000,
			},
			"type": "tcp_socket",
		},
	}

	defer terraform.Destroy(t, options)
	_, err = terraform.InitAndApplyE(t, options)
	assert.NoError(t, err)
}

func TestApplyAndDestroyWithSingleContainer(t *testing.T) {
	t.Parallel()

	namespace, options, err := getDefaultTerraformOptions(t, "sgl-cnt-without-map")
	assert.NoError(t, err)

	k8sOptions := k8s.NewKubectlOptions("", "", namespace)
	k8s.CreateNamespace(t, k8sOptions, namespace)
	// website::tag::5::Make sure to delete the namespace at the end of the test
	defer k8s.DeleteNamespace(t, k8sOptions, namespace)

	kubeResourcePath := "./resources.yml"
	defer k8s.KubectlDelete(t, k8sOptions, kubeResourcePath)
	k8s.KubectlApply(t, k8sOptions, kubeResourcePath)

	options.Vars["image"] = "\"training/webapp:latest\""

	options.Vars["ports"] = map[string]interface{}{
		"5000": map[string]interface{}{
			"protocol": "TCP",
		},
	}

	options.Vars["readiness_probes"] = map[string]interface{}{
		"tcp_socket": map[string]interface{}{
			"port": 5000,
		},
		"type": "tcp_socket",
	}

	options.Vars["liveness_probes"] = map[string]interface{}{
		"tcp_socket": map[string]interface{}{
			"port": 5000,
		},
		"type": "tcp_socket",
	}

	options.Vars["environment_variables_from_secret"] = map[string]interface{}{
		"SUPER_SECRET": map[string]interface{}{
			"secret_name": "test-secret",
			"secret_key":  "username",
		},
	}

	options.Vars["environment_variables"] = map[string]interface{}{
		"SUPER_VARIABLE": "super-value",
	}

	options.Vars["volumes_mounts_from_config_map"] = map[string]interface{}{
		"test-configmap": map[string]interface{}{
			"mount_path": "/data/myconfigmap",
			"sub_path":   "",
		},
	}

	options.Vars["volumes_mounts_from_secret"] = map[string]interface{}{
		"test-secret": map[string]interface{}{
			"mount_path": "/data/mysecret",
			"sub_path":   "",
		},
	}

	defer terraform.Destroy(t, options)
	_, err = terraform.InitAndApplyE(t, options)
	assert.NoError(t, err)
}

func TestApplyAndDestroyWithPlentyOfValues(t *testing.T) {
	t.Parallel()

	namespace, options, err := getDefaultTerraformOptions(t, "multi-cnt-plenty-vals")
	assert.NoError(t, err)

	k8sOptions := k8s.NewKubectlOptions("", "", namespace)
	k8s.CreateNamespace(t, k8sOptions, namespace)
	// website::tag::5::Make sure to delete the namespace at the end of the test
	defer k8s.DeleteNamespace(t, k8sOptions, namespace)

	kubeResourcePath := "./resources.yml"
	defer k8s.KubectlDelete(t, k8sOptions, kubeResourcePath)
	k8s.KubectlApply(t, k8sOptions, kubeResourcePath)

	options.Vars["image"] = map[string]interface{}{
		"test-container":   "training/webapp:latest",
		"test-container-2": "nginxdemos/hello",
	}

	options.Vars["ports"] = map[string]interface{}{
		"test-container": map[string]interface{}{
			"5000": map[string]interface{}{
				"protocol": "TCP",
			},
			"6000": map[string]interface{}{
				"protocol": "TCP",
				"ingress": map[string]interface{}{
					"foo.example.com": "/",
					"bar.example.com": "/",
				},
				"default_ingress_annotations": "traefik",
				"cert_manager_issuer":         "letsencrypt-prod",
				"path":                        "/api/v1",
				"ingress_annotations": map[string]interface{}{
					"foo.annotations.io": "bar",
				},
			},
		},
		"test-container-2": map[string]interface{}{
			"80": map[string]interface{}{
				"protocol": "TCP",
			},
		},
	}

	options.Vars["environment_variables_from_secret"] = map[string]interface{}{
		"test-container-2": map[string]interface{}{
			"SUPER_SECRET": map[string]interface{}{
				"secret_name": "test-secret",
				"secret_key":  "username",
			},
		},
	}

	options.Vars["environment_variables"] = map[string]interface{}{
		"test-container": map[string]interface{}{
			"SUPER_VARIABLE": "super-value",
		},
	}

	options.Vars["readiness_probes"] = map[string]interface{}{
		"test-container": map[string]interface{}{
			"tcp_socket": map[string]interface{}{
				"port": 5000,
			},
			"type": "tcp_socket",
		},
		"test-container-2": map[string]interface{}{
			"tcp_socket": map[string]interface{}{
				"port": 80,
			},
			"type": "tcp_socket",
		},
	}

	options.Vars["liveness_probes"] = map[string]interface{}{
		"test-container": map[string]interface{}{
			"tcp_socket": map[string]interface{}{
				"port": 5000,
			},
			"type": "tcp_socket",
		},
		"test-container-2": map[string]interface{}{
			"tcp_socket": map[string]interface{}{
				"port": 80,
			},
			"type": "tcp_socket",
		},
	}

	options.Vars["volumes_mounts_from_config_map"] = map[string]interface{}{
		"test-container": map[string]interface{}{
			"test-configmap": map[string]interface{}{
				"mount_path": "/data/myconfigmap",
				"sub_path":   "",
			},
		},
	}

	options.Vars["volumes_mounts_from_secret"] = map[string]interface{}{
		"test-container-2": map[string]interface{}{
			"test-secret": map[string]interface{}{
				"mount_path": "/data/mysecret",
				"sub_path":   "",
			},
		},
	}

	options.Vars["hpa"] = map[string]interface{}{
		"enabled":        true,
		"target_cpu":    50,
		"target_memory": 60,
		"min_replicas":  1,
		"max_replicas":   2,
	}

	defer terraform.Destroy(t, options)
	_, err = terraform.InitAndApplyE(t, options)
	assert.NoError(t, err)
}
