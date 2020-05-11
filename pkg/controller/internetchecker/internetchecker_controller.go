package internetchecker

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_internetchecker")

// Add creates a new Secret Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, stopCh <-chan struct{}) error {
	return add(mgr, newReconciler(mgr), stopCh)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &internetChecker{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, stopCh <-chan struct{}) error {
	c, err := controller.New("internetchecker-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	events := make(chan event.GenericEvent)
	timerSource := source.Channel{Source: events}
	ticker := time.NewTicker(10 * time.Second)
	timerSource.InjectStopChannel(stopCh)
	go func() {
		for {
			select {
			case <-ticker.C:
				events <- event.GenericEvent{
					Meta:   &metav1.ObjectMeta{},
					Object: &unstructured.Unstructured{},
				}
			case <-stopCh:
				log.Info("shutting down ticker")
				ticker.Stop()
				return
			}
		}
	}()
	pred := &predicate.Funcs{
		GenericFunc: r.(*internetChecker).pollEvent,
	}
	return c.Watch(&timerSource, &handler.EnqueueRequestForObject{}, pred)
}

// blank assignment to verify that internetChecker implements reconcile.Reconciler
var _ reconcile.Reconciler = &internetChecker{}

// internetChecker reconciles a Secret object
type internetChecker struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile is required by the interface, but not used
func (r *internetChecker) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (r *internetChecker) pollEvent(ev event.GenericEvent) bool {
	log.Info("Polling")
	return true
}
