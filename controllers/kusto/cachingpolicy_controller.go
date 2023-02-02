/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package kusto

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/go-logr/logr"
	kustov1alpha1 "github.com/microsoft/azure-schema-operator/apis/kusto/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	corev1 "k8s.io/api/core/v1"
)

// CachingPolicyReconciler reconciles a CachingPolicy object
type CachingPolicyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kusto.microsoft.com,resources=cachingpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kusto.microsoft.com,resources=cachingpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kusto.microsoft.com,resources=cachingpolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CachingPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *CachingPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("CachingPolicy", req.NamespacedName)

	cachingPolicy := &kustov1alpha1.CachingPolicy{}
	err := r.Get(ctx, req.NamespacedName, cachingPolicy)
	if err != nil {
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Loop over all clusters - check if the policy is set - if not - set it
	clustersDone := make([]string, 0)
	for _, cluster := range cachingPolicy.Spec.ClusterUris {
		kcsb := kusto.NewConnectionStringBuilder(cluster).WithDefaultAzureCredential()
		client, err := kusto.New(kcsb)
		if err != nil {
			log.Error(err, "Failed to create Kusto Client")
			return ctrl.Result{RequeueAfter: 10 * time.Minute}, err
		}
		defer client.Close()
		tablePolicy, err := kustoutils.GetTableCachingPolicy(ctx, client, cachingPolicy.Spec.DB, cachingPolicy.Spec.Table)
		if err != nil {
			log.Info("Failed to get caching Policy")
			return ctrl.Result{RequeueAfter: 10 * time.Minute}, err
		}
		if *tablePolicy != cachingPolicy.Spec.CachingPolicy {
			changedPolicy, err := kustoutils.SetTableCachingPolicy(ctx, client, cachingPolicy.Spec.DB, cachingPolicy.Spec.Table, &cachingPolicy.Spec.CachingPolicy)
			if err != nil || *changedPolicy != cachingPolicy.Spec.CachingPolicy {
				log.Error(err, "Failed to changing caching Policy")
				return ctrl.Result{RequeueAfter: 10 * time.Minute}, err
			}
			r.recorder.Eventf(cachingPolicy, corev1.EventTypeNormal, "Executed", "Set table policy in cluster  %s", cluster)
		}
		clustersDone = append(clustersDone, cluster)
	}

	cachingPolicy.Status.ClustersDone = clustersDone
	cachingPolicy.Status.Status = "Success"

	err = r.Status().Update(ctx, cachingPolicy)
	if err != nil {
		log.Error(err, "failed updating caching policy status", "request", req.String())
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CachingPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("CachingPolicy")
	return ctrl.NewControllerManagedBy(mgr).
		For(&kustov1alpha1.CachingPolicy{}).
		Complete(r)
}
