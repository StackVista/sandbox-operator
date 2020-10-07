/*


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
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv1 "github.com/stackvista/sandbox-operator/apis/devops/v1"
)

// SandboxReconciler reconciles a Sandbox object
type SandboxReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=devops.stackstate.com,resources=sandboxes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.stackstate.com,resources=sandboxes/status,verbs=get;update;patch

func (r *SandboxReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("sandbox", req.NamespacedName)

	sandbox := &devopsv1.Sandbox{}
	err := r.Get(ctx, req.NamespacedName, sandbox)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Sandbox resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Sandbox")
		return ctrl.Result{}, err
	}

	namespaceName := r.constructNamespaceName(sandbox)

	ns, err := r.findNamespace(ctx, namespaceName)
	if err != nil {
		return ctrl.Result{}, err
	}

	if ns == nil {
		log.Info("Provisioning Namespace for Sandbox")
		newNs := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaceName,
				Labels: map[string]string{
					"sandboxer/created-by": sandbox.Spec.User,
				},
			},
		}

		if err := ctrl.SetControllerReference(sandbox, newNs, r.Scheme); err != nil {
			log.Error(err, "Error setting controller reference to sandbox")
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, newNs, &client.CreateOptions{}); err != nil {
			log.Error(err, "Error creating new Namespace")
			return ctrl.Result{}, err
		}

		sandbox.Status.NamespaceStatus = &newNs.Status
		if err := r.Status().Update(ctx, sandbox); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if sandbox.Status.NamespaceStatus != nil && sandbox.Status.NamespaceStatus.Phase == v1.NamespaceActive {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("Namespace for sandbox already exists")
	}

	return ctrl.Result{}, nil
}

func (r *SandboxReconciler) constructNamespaceName(sandbox *devopsv1.Sandbox) string {
	name := "sandbox"
	if !strings.HasPrefix(sandbox.Name, sandbox.Spec.User) {
		name = fmt.Sprintf("%s-%s", name, sandbox.Spec.User)
	}

	name = fmt.Sprintf("%s-%s", name, sandbox.Name)

	return name
}

func (r *SandboxReconciler) findNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	namespaces := &corev1.NamespaceList{}
	if err := r.List(ctx, namespaces, &client.ListOptions{}); err != nil {
		return nil, err
	}

	for _, ns := range namespaces.Items {
		if ns.Name == name {
			return ns.DeepCopy(), nil
		}
	}

	return nil, nil
}

func (r *SandboxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.Sandbox{}).
		Complete(r)
}
