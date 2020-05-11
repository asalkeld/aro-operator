package statusreporter

import (
	"context"

	"github.com/operator-framework/operator-sdk/pkg/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	aro "github.com/asalkeld/aro-operator/pkg/apis/aro/v1alpha1"
)

var (
	log = logf.Log.WithName("statusreporter")
)

type StatusReporter struct {
	client client.Client
	name   types.NamespacedName
}

var emptyConditions = []status.Condition{
	{
		Type:    aro.InternetReachable,
		Status:  corev1.ConditionUnknown,
		Reason:  "",
		Message: "",
	},
	{
		Type:    aro.ClusterSupportable,
		Status:  corev1.ConditionUnknown,
		Reason:  "",
		Message: "",
	},
}

func NewStatusReporter(client_ client.Client, namespace, name string) *StatusReporter {
	return &StatusReporter{
		client: client_,
		name:   types.NamespacedName{Name: name, Namespace: namespace},
	}
}

func (r *StatusReporter) SetNoInternetConnection(connectionErr error) error {
	ctx := context.TODO()
	co := &aro.Cluster{}
	err := r.client.Get(ctx, r.name, co)
	if apierrors.IsNotFound(err) {
		co = r.newCluster()
		err = r.client.Create(ctx, co)
	}
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	time := metav1.Now()

	co.Status.Conditions.SetCondition(status.Condition{
		Type:               aro.InternetReachable,
		Status:             corev1.ConditionFalse,
		Message:            "Outgoing connection failed: " + connectionErr.Error(),
		Reason:             "CheckFailed",
		LastTransitionTime: time})

	// we need only one condition to make it not supportable
	co.Status.Conditions.SetCondition(status.Condition{
		Type:               aro.ClusterSupportable,
		Status:             corev1.ConditionFalse,
		Message:            "Cluster is NOT supportable.",
		Reason:             "SomeChecksFailed",
		LastTransitionTime: time})
	return r.client.Status().Update(ctx, co)
}

func (r *StatusReporter) SetInternetConnected() error {
	ctx := context.TODO()
	co := &aro.Cluster{}
	err := r.client.Get(ctx, r.name, co)
	if apierrors.IsNotFound(err) {
		co = r.newCluster()
		err = r.client.Create(ctx, co)
	}
	if err != nil {
		return err
	}

	time := metav1.Now()
	co.Status.Conditions.SetCondition(status.Condition{
		Type:               aro.InternetReachable,
		Status:             corev1.ConditionTrue,
		Message:            "Outgoing connection successful.",
		Reason:             "CheckDone",
		LastTransitionTime: time})

	supportable := true
	for _, cond := range co.Status.Conditions {
		if cond.IsFalse() {
			supportable = false
		}
	}

	if supportable {
		co.Status.Conditions.SetCondition(status.Condition{
			Type:               aro.ClusterSupportable,
			Status:             corev1.ConditionTrue,
			Message:            "Cluster is supportable.",
			Reason:             "AllChecksDone",
			LastTransitionTime: time})
	} else {
		co.Status.Conditions.SetCondition(status.Condition{
			Type:               aro.ClusterSupportable,
			Status:             corev1.ConditionFalse,
			Message:            "Cluster is NOT supportable.",
			Reason:             "SomeChecksFailed",
			LastTransitionTime: time})
	}
	return r.client.Status().Update(ctx, co)
}

func newRelatedObjects(namespace string) []corev1.ObjectReference {
	return []corev1.ObjectReference{
		{Kind: "Namespace", Name: namespace},
		{Kind: "Secret", Name: "pull-secret", Namespace: "openshift-config"},
	}
}

func (r *StatusReporter) newCluster() *aro.Cluster {
	co := &aro.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "aro.openshift.io/v1alpha1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.name.Name,
			Namespace: r.name.Namespace,
		},
		Spec: aro.ClusterSpec{},
		Status: aro.ClusterStatus{
			Conditions: emptyConditions,
		},
	}
	co.Status.RelatedObjects = newRelatedObjects(r.name.Namespace)
	return co
}
