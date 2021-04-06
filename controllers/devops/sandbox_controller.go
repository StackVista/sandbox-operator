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

package devops

import (
	"context"
	"fmt"

	"github.com/stackvista/sandbox-operator/internal/config"
	"github.com/stackvista/sandbox-operator/pkg/operator"
	pkgsandbox "github.com/stackvista/sandbox-operator/pkg/sandbox"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	devopsv1 "github.com/stackvista/sandbox-operator/apis/devops/v1"
	mgr "sigs.k8s.io/controller-runtime/pkg/manager"
)

type SandboxReconcilerFactory struct {
	Config *config.Config
}

func (f *SandboxReconcilerFactory) NewReconciler(ctx context.Context, mgr mgr.Manager) (operator.Reconciler, error) {
	return &SandboxReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Sandbox"),
		Scheme: mgr.GetScheme(),
	}, nil
}

// SandboxReconciler reconciles a Sandbox object
type SandboxReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=devops.stackstate.com,resources=sandboxes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.stackstate.com,resources=sandboxes/status,verbs=get;update;patch

func (r *SandboxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	namespaceName := pkgsandbox.SandboxName(sandbox)

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

		log.WithValues("status.phase", newNs.Status.Phase).Info("Created namespace is now...")
		_, err = controllerutil.CreateOrUpdate(ctx, r.Client, sandbox.DeepCopy(), func() error {
			sandbox.Status.NamespaceStatus = newNs.Status
			return r.Client.Status().Update(ctx, sandbox, &client.UpdateOptions{})
		})
		if err != nil {
			log.Error(err, "Unable to update sandbox status")
			return ctrl.Result{}, err
		}
	} else {
		if sandbox.Status.NamespaceStatus.Phase == v1.NamespaceActive {
			return ctrl.Result{}, nil
		}

		log.WithValues("status.phase", sandbox.Status.NamespaceStatus.Phase).Info("Namespace exists, but sandbox status is not Active")

		return ctrl.Result{}, fmt.Errorf("Namespace for sandbox already exists")
	}

	return ctrl.Result{}, nil
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
