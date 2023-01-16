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
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Azure/azure-kusto-go/kusto"
	kustov1alpha1 "github.com/microsoft/azure-schema-operator/apis/kusto/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	corev1 "k8s.io/api/core/v1"
)

// RetentionPolicyReconciler reconciles a RetentionPolicy object
type RetentionPolicyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kusto.microsoft.com,resources=retentionpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kusto.microsoft.com,resources=retentionpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kusto.microsoft.com,resources=retentionpolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// This controller tries to keep things as simple as possible, so it loops over the clusters in the order given in the
// CRD. It will block on first failure - and retry.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *RetentionPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	retentionPolicy := &kustov1alpha1.RetentionPolicy{}
	err := r.Get(ctx, req.NamespacedName, retentionPolicy)
	if err != nil {
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Loop over all clusters - check if the policy is set - if not - set it
	clustersDone := make([]string, 0)
	for _, cluster := range retentionPolicy.Spec.ClusterUris {
		kcsb := kusto.NewConnectionStringBuilder(cluster).WithDefaultAzureCredential()
		client, err := kusto.New(kcsb)
		if err != nil {
			log.Error(err, "Failed to create Kusto Client")
			return ctrl.Result{RequeueAfter: 10 * time.Minute}, err
		}
		defer client.Close()
		tablePolicy, err := kustoutils.GetTableRetentionPolicy(ctx, client, retentionPolicy.Spec.DB, retentionPolicy.Spec.Table)
		if err != nil {
			log.Info("Failed to get retention Policy")
			return ctrl.Result{RequeueAfter: 10 * time.Minute}, err
		}
		if *tablePolicy != retentionPolicy.Spec.RetentionPolicy {
			changedPolicy, err := kustoutils.SetTableRetentionPolicy(ctx, client, retentionPolicy.Spec.DB, retentionPolicy.Spec.Table, &retentionPolicy.Spec.RetentionPolicy)
			if err != nil || *changedPolicy != retentionPolicy.Spec.RetentionPolicy {
				log.Error(err, "Failed to changing retention Policy")
				return ctrl.Result{RequeueAfter: 10 * time.Minute}, err
			}
			r.recorder.Eventf(retentionPolicy, corev1.EventTypeNormal, "Executed", "Set table policy in cluster  %s", cluster)
		}
		clustersDone = append(clustersDone, cluster)
	}

	retentionPolicy.Status.ClustersDone = clustersDone
	retentionPolicy.Status.Status = "Success"

	err = r.Status().Update(ctx, retentionPolicy)
	if err != nil {
		log.Error(err, "failed updating retention policy status", "request", req.String())
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RetentionPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("RetentionPolicy")
	return ctrl.NewControllerManagedBy(mgr).
		For(&kustov1alpha1.RetentionPolicy{}).
		Complete(r)
}
