/*
Copyright 2020 The KEDA Authors

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

package keda

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	mf "github.com/manifestival/manifestival"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	kedav1alpha1 "github.com/kedacore/keda-olm-operator/apis/keda/v1alpha1"
)

var _ = Describe("Deploying KedaController manifest", func() {
	const (
		olmOperatorName      = "keda-olm-operator"
		operatorName         = "keda-operator"
		kedaManifestFilepath = "../../config/samples/keda_v1alpha1_kedacontroller.yaml"
		timeout              = time.Second * 60
		interval             = time.Millisecond * 250
		namespace            = "keda"
	)

	var (
		ctx      = context.Background()
		err      error
		scheme   *runtime.Scheme
		manifest mf.Manifest
	)

	BeforeEach(func() {
		Eventually(func() error {
			_, err = getObject(ctx, "Pod", olmOperatorName, namespace, k8sClient)
			return err
		}, timeout, interval).Should(Succeed())

	})

	var _ = Describe("Changing namespace", func() {

		const ()

		BeforeEach(func() {
			scheme = k8sManager.GetScheme()
			manifest, err = createManifest(kedaManifestFilepath, k8sClient)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			manifest, err = changeAttribute(manifest, "namespace", namespace, scheme, "")
			Expect(err).To(BeNil())

			Expect(manifest.Delete()).Should(Succeed())

			Eventually(func() error {
				_, err = getObject(ctx, "Pod", operatorName, namespace, k8sClient)
				return err
			}, timeout, interval).ShouldNot(Succeed())
		})

		Context("When deploying in \"keda\" namespace", func() {
			It("Should deploy KedaController", func() {

				Expect(manifest.Apply()).Should(Succeed())

				Eventually(func() error {
					_, err = getObject(ctx, "Pod", operatorName, namespace, k8sClient)
					return err
				}, timeout, interval).Should(Succeed())
			})
		})

		Context("When deploying not in \"keda\" namespace", func() {
			const changedNamespace = "default"

			It("Should not deploy KedaController", func() {

				manifest, err = changeAttribute(manifest, "namespace", changedNamespace, scheme, "")
				Expect(err).To(BeNil())

				Expect(manifest.Apply()).Should(Succeed())

				Eventually(func() error {
					_, err = getObject(ctx, "Pod", operatorName, namespace, k8sClient)
					return err
				}, timeout, interval).ShouldNot(Succeed())
			})
		})
	})
})

var _ = Describe("Testing functionality", func() {

	var _ = Describe("Changing operator parameters", func() {

		const (
			deploymentName       = "keda-operator"
			containerName        = "keda-operator"
			logLevelPrefix       = "--zap-log-level="
			kind                 = "KedaController"
			name                 = "keda"
			namespace            = "keda"
			kedaManifestFilepath = "../../config/samples/keda_v1alpha1_kedacontroller.yaml"
		)

		var (
			ctx      = context.Background()
			timeout  = time.Second * 60
			interval = time.Millisecond * 250
			scheme   *runtime.Scheme
			manifest mf.Manifest
			err      error
			arg      string
			dep      = &appsv1.Deployment{}
		)

		waitForTestCaseLogLevel := func(logLevel string, deploymentAnnotation string) error {
			u, err := getObject(ctx, "Deployment", deploymentName, namespace, k8sClient)
			if err != nil {
				return err
			}
			err = scheme.Convert(u, dep, nil)
			if err != nil {
				return err
			}
			arg, err = getDepArg(dep, logLevelPrefix, containerName)
			if err != nil {
				return err
			}
			testcase, ok := u.GetAnnotations()["testCase"]
			if ok && deploymentAnnotation == testcase && arg == logLevel {
				return nil
			}
			return fmt.Errorf("log level was not set to %s, log level was %s", logLevel, arg)

		}

		BeforeEach(func() {
			By("Applying a manifest that defaults operator logLevel back to info")
			scheme = k8sManager.GetScheme()
			manifest, err = createManifest(kedaManifestFilepath, k8sClient)
			Expect(err).To(BeNil())
			manifest, err = changeAttribute(manifest, "logLevel", "info", scheme, "default")
			Expect(err).To(BeNil())
			Expect(manifest.Apply()).Should(Succeed())
			By("Waiting for the operator deployment to reflect the changes")
			Eventually(func() error { return waitForTestCaseLogLevel("info", "default") }, timeout, interval).Should(Succeed())
		})

		Context("When changing \"--zap-log-level\"", func() {
			variants := []struct {
				initialLogLevel string
				actualLogLevel  string
			}{
				{
					initialLogLevel: "debug",
					actualLogLevel:  "debug",
				},
				{
					initialLogLevel: "info",
					actualLogLevel:  "info",
				},
				{
					initialLogLevel: "error",
					actualLogLevel:  "error",
				},
			}

			for _, variant := range variants {
				caseName := fmt.Sprintf("Should change it, initialLoglevel='%s', actualLoglevel='%s'", variant.initialLogLevel, variant.actualLogLevel)
				It(caseName,
					func() {
						By(fmt.Sprintf("Setting operator loglevel to %s in kedaController manifest", variant.initialLogLevel))
						manifest, err = changeAttribute(manifest, "logLevel", variant.initialLogLevel, scheme, caseName)
						Expect(err).To(BeNil())
						err = manifest.Apply()
						Expect(err).To(BeNil())
						By("Waiting for the operator deployment to reflect the changes")
						Eventually(func() error { return waitForTestCaseLogLevel(variant.initialLogLevel, caseName) }, timeout, interval).Should(Succeed())

					})
			}

			variants = []struct {
				initialLogLevel string
				actualLogLevel  string
			}{
				{
					initialLogLevel: "",
					actualLogLevel:  "info",
				},
				{
					initialLogLevel: "foo",
					actualLogLevel:  "info",
				},
			}

			for _, variant := range variants {
				caseName := fmt.Sprintf("Should not change it, initialLoglevel='%s', actualLoglevel='%s'", variant.initialLogLevel, variant.actualLogLevel)
				It(caseName,
					func() {
						By(fmt.Sprintf("Setting operator loglevel to %s in kedaController manifest", variant.initialLogLevel))
						manifest, err = changeAttribute(manifest, "logLevel", variant.initialLogLevel, scheme, caseName)
						Expect(err).To(BeNil())
						err = manifest.Apply()
						Expect(err).To(BeNil())
						By("Waiting for the operator deployment to reflect the changes")
						Eventually(func() error { return waitForTestCaseLogLevel(variant.actualLogLevel, caseName) }, timeout, interval).Should(Succeed())

					})
			}
		})
	})

	var _ = Describe("Changing webhook parameters", func() {

		const (
			deploymentName       = "keda-admission"
			containerName        = "keda-admission-webhooks"
			logLevelPrefix       = "--zap-log-level="
			kind                 = "KedaController"
			name                 = "keda"
			namespace            = "keda"
			kedaManifestFilepath = "../../config/samples/keda_v1alpha1_kedacontroller.yaml"
		)

		var (
			ctx      = context.Background()
			timeout  = time.Second * 20
			interval = time.Millisecond * 250
			scheme   *runtime.Scheme
			manifest mf.Manifest
			err      error
			arg      string
			dep      = &appsv1.Deployment{}
		)

		waitForLogLevelToBe := func(logLevel string, deploymentAnnotation string) error {
			u, err := getObject(ctx, "Deployment", deploymentName, namespace, k8sClient)
			if err != nil {
				return err
			}
			err = scheme.Convert(u, dep, nil)
			if err != nil {
				return err
			}
			arg, err = getDepArg(dep, logLevelPrefix, containerName)
			if err != nil {
				return err
			}
			testcase, ok := u.GetAnnotations()["testCase"]
			if ok && deploymentAnnotation == testcase && arg == logLevel {
				return nil
			}
			return fmt.Errorf("log level was not set to %s, log level was %s", logLevel, arg)

		}

		BeforeEach(func() {
			By("Applying a manifest that defaults admission logLevel back to info")
			scheme = k8sManager.GetScheme()
			manifest, err = createManifest(kedaManifestFilepath, k8sClient)
			Expect(err).To(BeNil())
			manifest, err = changeAttribute(manifest, "logLevel-admission", "info", scheme, "default")
			Expect(err).To(BeNil())
			Expect(manifest.Apply()).Should(Succeed())
			By("Waiting for the admission deployment to reflect the changes")
			Eventually(func() error { return waitForLogLevelToBe("info", "default") }, timeout, interval).Should(Succeed())
		})

		Context("When changing \"--zap-log-level\"", func() {
			variants := []struct {
				initialLogLevel string
				actualLogLevel  string
			}{
				{
					initialLogLevel: "debug",
					actualLogLevel:  "debug",
				},
				{
					initialLogLevel: "info",
					actualLogLevel:  "info",
				},
				{
					initialLogLevel: "error",
					actualLogLevel:  "error",
				},
			}

			for _, variant := range variants {
				caseName := fmt.Sprintf("Should change it, initialLoglevel='%s', actualLoglevel='%s'", variant.initialLogLevel, variant.actualLogLevel)

				It(caseName, func() {
					By(fmt.Sprintf("Setting admission loglevel to %s in kedaController manifest", variant.initialLogLevel))
					manifest, err = changeAttribute(manifest, "logLevel-admission", variant.initialLogLevel, scheme, caseName)
					Expect(err).To(BeNil())
					err = manifest.Apply()
					Expect(err).To(BeNil())
					By("Waiting for the admission deployment to reflect the changes")
					Eventually(func() error { return waitForLogLevelToBe(variant.initialLogLevel, caseName) }, timeout, interval).Should(Succeed())

				})
			}

			variants = []struct {
				initialLogLevel string
				actualLogLevel  string
			}{
				// the default in the sample kedacontroller manifest is "info", so "info" in this list can also mean "make sure it doesn't change"
				{
					initialLogLevel: "",
					actualLogLevel:  "info",
				},
				{
					initialLogLevel: "foo",
					actualLogLevel:  "info",
				},
			}

			for _, variant := range variants {
				caseName := fmt.Sprintf("Should not change it, initialLoglevel='%s', actualLoglevel='%s'", variant.initialLogLevel, variant.actualLogLevel)

				It(caseName, func() {
					By(fmt.Sprintf("Setting admission loglevel to %s in kedaController manifest", variant.initialLogLevel))
					manifest, err = changeAttribute(manifest, "logLevel-admission", variant.initialLogLevel, scheme, caseName)
					Expect(err).To(BeNil())
					err = manifest.Apply()
					Expect(err).To(BeNil())
					By("Waiting for the admission deployment to reflect the changes")
					Eventually(func() error { return waitForLogLevelToBe(variant.actualLogLevel, caseName) }, timeout, interval).Should(Succeed())

				})
			}

		})
	})

})

var _ = Describe("Testing audit flags", func() {
	const (
		metricsServerName    = "keda-metrics-apiserver"
		kind                 = "KedaController"
		name                 = "keda"
		namespace            = "keda"
		kedaManifestFilepath = "../../config/samples/keda_v1alpha1_kedacontroller.yaml"
	)

	var (
		ctx      = context.Background()
		timeout  = time.Second * 60
		interval = time.Millisecond * 250
		scheme   *runtime.Scheme
		manifest mf.Manifest
		err      error
		dep      = &appsv1.Deployment{}
	)

	When("Manipulating parameters", func() {
		BeforeEach(func() {
			scheme = k8sManager.GetScheme()
			dep = &appsv1.Deployment{}
			manifest, err = createManifest(kedaManifestFilepath, k8sClient)
			Expect(err).To(BeNil())
		})

		Context("to add audit configuration flags", func() {
			vars := []struct {
				argument string
				prefix   string
				value    string
			}{
				{
					argument: "auditLogFormat",
					prefix:   "--audit-log-format=",
					value:    "json",
				},
				{
					argument: "auditMaxAge=",
					prefix:   "--audit-log-maxage=",
					value:    "1",
				},
				{
					argument: "auditMaxBackup",
					prefix:   "--audit-log-maxbackup=",
					value:    "2",
				},
				{
					argument: "auditLogMaxSize",
					prefix:   "--audit-log-maxsize=",
					value:    "3",
				},
			}
			for _, variant := range vars {

				It(fmt.Sprintf("adds '%s' with value '%s'", variant.argument, variant.value), func() {
					manifest, err := changeAttribute(manifest, variant.argument, variant.value, scheme, "")
					Expect(manifest.Apply()).To(Succeed())
					Eventually(func() error {
						_, err = getObject(ctx, "Deployment", metricsServerName, namespace, k8sClient)
						return err
					}, timeout, interval).Should(Succeed())
					u, err := getObject(ctx, "Deployment", metricsServerName, namespace, k8sClient)
					Expect(err).To(BeNil())

					Expect(scheme.Convert(u, dep, nil)).To(Succeed())
					arg, err := getDepArg(dep, variant.prefix, metricsServerName)
					Expect(err).To(BeNil())
					Expect(arg).To(Equal(variant.value))
				})
			}

		})
	})
})

func getDepArg(dep *appsv1.Deployment, prefix string, containerName string) (string, error) {
	for _, container := range dep.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			for _, arg := range container.Args {
				if strings.HasPrefix(arg, prefix) {
					return strings.TrimPrefix(arg, prefix), nil
				}
			}
			return "", errors.New("Could not find an argument with given prefix: " + prefix)
		}
	}
	return "", errors.New("Could not find a container: " + containerName)
}

func changeAttribute(manifest mf.Manifest, attr string, value string, scheme *runtime.Scheme, annotation string) (mf.Manifest, error) {
	transformer := func(u *unstructured.Unstructured) error {
		kedaControllerInstance := &kedav1alpha1.KedaController{}
		if err := scheme.Convert(u, kedaControllerInstance, nil); err != nil {
			return err
		}

		// Annotations might be nil, so we need to make sure we account for that
		if kedaControllerInstance.Spec.Operator.DeploymentAnnotations == nil {
			kedaControllerInstance.Spec.Operator.DeploymentAnnotations = make(map[string]string)
		}

		if kedaControllerInstance.Spec.AdmissionWebhooks.DeploymentAnnotations == nil {
			kedaControllerInstance.Spec.AdmissionWebhooks.DeploymentAnnotations = make(map[string]string)
		}
		if kedaControllerInstance.Spec.MetricsServer.DeploymentAnnotations == nil {
			kedaControllerInstance.Spec.MetricsServer.DeploymentAnnotations = make(map[string]string)
		}

		// When we push through an attribute change, we also set an annotation matching the test case
		// so we can be sure the controller reacted to our kedaControlelrInstance updates and we have "our"
		// changes, not just some changes that might be from a previous test case
		kedaControllerInstance.Spec.Operator.DeploymentAnnotations["testCase"] = annotation
		kedaControllerInstance.Spec.AdmissionWebhooks.DeploymentAnnotations["testCase"] = annotation
		kedaControllerInstance.Spec.MetricsServer.DeploymentAnnotations["testCase"] = annotation

		switch attr {
		case "namespace":
			kedaControllerInstance.Namespace = value
		case "logLevel":
			kedaControllerInstance.Spec.Operator.LogLevel = value
		// TODO(jkyros): should probably a separate arg or separate function, this
		// breaks pattern with the rest of these cases but multiple operands have
		// the same field and we wanted to explicitly test admission loglevel separately
		case "logLevel-admission":
			kedaControllerInstance.Spec.AdmissionWebhooks.LogLevel = value
		// metricsServer audit arguments
		case "auditLogFormat":
			kedaControllerInstance.Spec.MetricsServer.AuditConfig.LogFormat = value
		case "auditMaxAge":
			kedaControllerInstance.Spec.MetricsServer.AuditConfig.AuditLifetime.MaxAge = value
		case "auditMaxBackup":
			kedaControllerInstance.Spec.MetricsServer.AuditConfig.AuditLifetime.MaxBackup = value
		case "auditLogMaxSize":
			kedaControllerInstance.Spec.MetricsServer.AuditConfig.AuditLifetime.MaxSize = value
		default:
			return errors.New("Not a valid attribute")
		}
		return scheme.Convert(kedaControllerInstance, u, nil)
	}

	return manifest.Transform(transformer)
}
