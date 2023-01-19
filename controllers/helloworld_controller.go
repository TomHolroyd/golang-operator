/*
Copyright 2023.

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

package controllers

import (
	"context"

	"github.com/tomholroyd96/api/v1alpha1"
	helloworldv1alpha1 "github.com/tomholroyd96/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// HelloWorldReconciler reconciles a HelloWorld object
type HelloWorldReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=helloworld.example.net,resources=helloworlds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=helloworld.example.net,resources=helloworlds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=helloworld.example.net,resources=helloworlds/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HelloWorld object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *HelloWorldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Verifying that hello world deployment exists")
	helloWorld := &helloworldv1alpha1.HelloWorld{}
	err := r.Get(ctx, req.NamespacedName, helloWorld)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Heloo world not found, deleting. Ignoring since object must be deleted")
			return ctrl.Result{}, err
		}
		logger.Error(err, "Failed to get hello world application")
		return ctrl.Result{}, err
	}
	logger.Info("Verify if the deployment already exists, if not create a new one")
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: helloWorld.Name, Namespace: helloWorld.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Deploying doesn't exist.. creating.")
		dep := r.deploymentForHelloWorld(helloWorld, ctx)
		err = r.Create(ctx, dep)
		if err != nil {
			logger.Error(err, "Error occured creating new hello world deployment")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get deployment")
		return ctrl.Result{}, nil
	}

	// TODO(user): your logic here
	logger.Info("Just returning nil")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloWorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&helloworldv1alpha1.HelloWorld{}).
		Complete(r)
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloWorldReconciler) deploymentForHelloWorld(helloWorld *v1alpha1.HelloWorld, ctx context.Context) *appsv1.Deployment {
	logger := log.FromContext(ctx)
	ls := labelsForHelloWorld(helloWorld.Name, helloWorld.Name)
	replicas := helloWorld.Spec.Size

	// Just reflect the command in the deployment.yaml
	// for the ReadinessProbe and LivenessProbe
	// command: ["sh", "-c", "curl -s http://localhost:8080"]
	mycommand := make([]string, 3)
	mycommand[0] = "/bin/sh"
	mycommand[1] = "-c"
	mycommand[2] = "./main"

	// Using the context to log information
	logger.Info("Logging: Creating a new Deployment", "Replicas", replicas)
	message := "Logging: (Name: " + helloWorld.Name + ") \n"
	logger.Info(message)
	message = "Logging: (Namespace: " + helloWorld.Namespace + ") \n"
	logger.Info(message)

	for key, value := range ls {
		message = "Logging: (Key: [" + key + "] Value: [" + value + "]) \n"
		logger.Info(message)
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helloWorld.Name,
			Namespace: helloWorld.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "tomholroyd96/gin-hello-world:v1.0.0",
						Name:  "hello-world",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "gin-port",
						}},
						Env: []corev1.EnvVar{
							{Name: "VUE_APP_API_URL_CATEGORIES",
								Value: "VUE_APP_API_URL_CATEGORIES_VALUE",
							}}, // End of Env listed values and Env definition
					}}, // Container
				}, // PodSec
			}, // PodTemplateSpec
		}, // Spec
	} // Deployment

	// Set TenancyFrontend instance as the owner and controller
	ctrl.SetControllerReference(helloWorld, dep, r.Scheme)
	return dep
}

func labelsForHelloWorld(name_app string, name_cr string) map[string]string {
	return map[string]string{"app": name_app, "tenancyfrontend_cr": name_cr}
}
