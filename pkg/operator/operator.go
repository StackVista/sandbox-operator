//nolint:gochecknoinits
package operator

import (
	"context"
	"os"

	devopsv1 "github.com/stackvista/sandbox-operator/apis/devops/v1"

	"github.com/butonic/zerologr"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // Blank import GCP AuthProvider
	ctrl "sigs.k8s.io/controller-runtime"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(devopsv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
} //nolint:wsl

type Config struct {
	MetricsAddr          string
	EnableLeaderElection bool
}

func StartOperator(ctx context.Context, config *Config, reconcilerFactory ReconcilerFactory) error {
	logger := log.Ctx(ctx)

	ctrl.SetLogger(zerologr.NewWithOptions(zerologr.Options{
		Logger: logger,
	}))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: config.MetricsAddr,
		Port:               9443,
		LeaderElection:     config.EnableLeaderElection,
		LeaderElectionID:   "sandboxer.stackstate.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	r, err := reconcilerFactory.NewReconciler(ctx, mgr)
	if err != nil {
		setupLog.Error(err, "unable to create reconciler", "reconciler", "Tenant")
		os.Exit(1)
	}

	if err := r.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Tenant")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

	return nil
}
