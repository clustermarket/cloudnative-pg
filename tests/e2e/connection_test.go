/*
This file is part of Cloud Native PostgreSQL.

Copyright (C) 2019-2021 EnterpriseDB Corporation.
*/

package e2e

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/EnterpriseDB/cloud-native-postgresql/tests"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Set of tests in which we check that we're able to connect to the -rw,
// -ro and -r services, using both the application user and the superuser one
var _ = Describe("Connection via services", func() {
	// We test custom db name and user
	const appDBName = "appdb"
	const appDBUser = "appuser"

	AssertServices := func(namespace string,
		clusterName string,
		appDBName string,
		appDBUser string,
		appPassword string,
		superuserPassword string,
		env *tests.TestingEnvironment) {

		// we use a pod in the cluster to have a psql client ready and
		// internal access to the k8s cluster
		podName := clusterName + "-1"
		pod := &corev1.Pod{}
		namespacedName := types.NamespacedName{
			Namespace: namespace,
			Name:      podName,
		}
		err := env.Client.Get(env.Ctx, namespacedName, pod)
		Expect(err).ToNot(HaveOccurred())

		// We test -rw, -ro and -r services with the app user and the superuser
		rwService := fmt.Sprintf("%v-rw.%v.svc", clusterName, namespace)
		rService := fmt.Sprintf("%v-r.%v.svc", clusterName, namespace)
		roService := fmt.Sprintf("%v-ro.%v.svc", clusterName, namespace)
		services := []string{rwService, roService, rService}
		for _, service := range services {
			AssertConnection(service, "postgres", appDBName, superuserPassword, *pod, 10, env)
			AssertConnection(service, appDBUser, appDBName, appPassword, *pod, 10, env)
		}

		AssertWritesToReplicaFails(pod, roService, appDBName, appDBUser, appPassword)
		AssertWritesToPrimarySucceeds(pod, rwService, appDBName, appDBUser, appPassword)
	}

	Context("Auto-generated passwords", func() {
		const namespace = "cluster-autogenerated-secrets-e2e"
		const sampleFile = fixturesDir + "/secrets/cluster-auto-generated.yaml"
		const clusterName = "postgresql-auto-generated"
		JustAfterEach(func() {
			if CurrentSpecReport().Failed() {
				env.DumpClusterEnv(namespace, clusterName,
					"out/"+CurrentSpecReport().LeafNodeText+".log")
			}
		})
		AfterEach(func() {
			err := env.DeleteNamespace(namespace)
			Expect(err).ToNot(HaveOccurred())
		})
		// If we don't specify secrets, the operator should autogenerate them.
		// We check that we're able to use them
		It("can connect with auto-generated passwords", func() {
			// Create a cluster in a namespace we'll delete after the test
			err := env.CreateNamespace(namespace)
			Expect(err).ToNot(HaveOccurred())
			AssertCreateCluster(namespace, clusterName, sampleFile, env)

			// Get the superuser password from the -superuser secret
			superuserSecretName := clusterName + "-superuser"
			superuserSecret := &corev1.Secret{}
			superuserSecretNamespacedName := types.NamespacedName{
				Namespace: namespace,
				Name:      superuserSecretName,
			}
			err = env.Client.Get(env.Ctx, superuserSecretNamespacedName, superuserSecret)
			Expect(err).ToNot(HaveOccurred())
			generatedSuperuserPassword := string(superuserSecret.Data["password"])

			// Get the app user password from the -app secret
			appSecretName := clusterName + "-app"
			appSecret := &corev1.Secret{}
			appSecretNamespacedName := types.NamespacedName{
				Namespace: namespace,
				Name:      appSecretName,
			}
			err = env.Client.Get(env.Ctx, appSecretNamespacedName, appSecret)
			Expect(err).ToNot(HaveOccurred())
			generatedAppUserPassword := string(appSecret.Data["password"])

			AssertServices(namespace, clusterName, appDBName, appDBUser,
				generatedAppUserPassword, generatedSuperuserPassword, env)
		})
	})

	Context("User-defined secrets", func() {
		const namespace = "cluster-user-supplied-secrets-e2e"
		const sampleFile = fixturesDir + "/secrets/cluster-user-supplied.yaml"
		const clusterName = "postgresql-user-supplied"
		JustAfterEach(func() {
			if CurrentSpecReport().Failed() {
				env.DumpClusterEnv(namespace, clusterName,
					"out/"+CurrentSpecReport().LeafNodeText+".log")
			}
		})
		AfterEach(func() {
			err := env.DeleteNamespace(namespace)
			Expect(err).ToNot(HaveOccurred())
		})
		// If we have specified secrets, we test that we're able to use them
		// to connect
		It("can connect with user-supplied passwords", func() {
			const suppliedSuperuserPassword = "v3ry54f3" // NOSONAR
			const suppliedAppUserPassword = "4ls054f3"   // NOSONAR

			// Create a cluster in a namespace we'll delete after the test
			err := env.CreateNamespace(namespace)
			Expect(err).ToNot(HaveOccurred())
			AssertCreateCluster(namespace, clusterName, sampleFile, env)
			AssertServices(namespace, clusterName, appDBName, appDBUser,
				suppliedAppUserPassword, suppliedSuperuserPassword, env)
		})
	})
})
