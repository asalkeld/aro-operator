package pullsecret

import (
	"context"
	"os"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_pullsecret")

// Add creates a new Secret Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, stopCh <-chan struct{}) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSecret{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pull-secret-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secrets (TODO putting the name/namespace in here does't filter by name)
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSecret implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSecret{}

// ReconcileSecret reconciles a Secret object
type ReconcileSecret struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the pull-secret object and makes sure the ACR
// repository is always configured
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	if request.Namespace != "openshift-config" || request.Name != "pull-secret" {
		// filter out other secrets.
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Reconciling pull-secret")

	// Fetch the Secret instance
	instance := &corev1.Secret{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new pull secret object
	ps := newPullSecret(instance)
	isCreate := false
	found := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ps.Name, Namespace: ps.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		isCreate = true
	} else if err == nil {
		ps = found
	} else if err != nil {
		return reconcile.Result{}, err
	}

	changed, err := r.repair(ps)
	if err != nil {
		return reconcile.Result{}, err
	}

	if isCreate {
		reqLogger.Info("Creating a new Pull Secret", "Secret.Namespace", ps.Namespace, "Secret.Name", ps.Name)
		err = r.client.Create(context.TODO(), ps)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if changed {
		reqLogger.Info("Updating Pull Secret", "Secret.Namespace", ps.Namespace, "Secret.Name", ps.Name)
		err = r.client.Update(context.TODO(), ps)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else {
		reqLogger.Info("Skip reconcile: Pull Secret not changed", "Secret.Namespace", ps.Namespace, "Secret.Name", ps.Name)
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileSecret) repair(cr *corev1.Secret) (bool, error) {
	if cr.Data == nil {
		cr.Data = map[string][]byte{}
	}

	// The idea here is you mount a secret as a file under /pull-secrets with
	// the same name as the registry in the pull secret.
	psPath := "/pull-secrets"
	pathOverride := os.Getenv("PULL_SECRET_PATH") // for development
	if pathOverride != "" {
		psPath = pathOverride
	}

	newPS, changed, err := repair(cr.Data[corev1.DockerConfigJsonKey], psPath)
	if err != nil {
		return false, err
	}
	if changed {
		cr.Data[corev1.DockerConfigJsonKey] = newPS
	}
	return changed, nil
}

// newPullSecret returns a pull-secret with the same name/namespace as the cr
func newPullSecret(cr *corev1.Secret) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Type: v1.SecretTypeDockerConfigJson,
	}
}
