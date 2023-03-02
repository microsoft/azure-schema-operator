/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package kusto

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/go-logr/logr"
	kustov1alpha1 "github.com/microsoft/azure-schema-operator/apis/kusto/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils/types"
	corev1 "k8s.io/api/core/v1"
)

// StoredFunctionReconciler reconciles a StoredFunction object
type StoredFunctionReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kusto.microsoft.com,resources=storedfunctions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kusto.microsoft.com,resources=storedfunctions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kusto.microsoft.com,resources=storedfunctions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the StoredFunction object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *StoredFunctionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("StoredFunction", req.NamespacedName)

	storedFunction := &kustov1alpha1.StoredFunction{}
	err := r.Get(ctx, req.NamespacedName, storedFunction)
	if err != nil {
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	kustoFunc := types.KustoFunction{
		Name:       storedFunction.Spec.Name,
		Parameters: storedFunction.Spec.Parameters,
		Body:       storedFunction.Spec.Body,
		DocString:  storedFunction.Spec.DocString,
		Folder:     storedFunction.Spec.Folder,
	}

	// Loop over all clusters - check if the policy is set - if not - set it
	clustersDone := make([]string, 0)
	var executionError error
	for _, cluster := range storedFunction.Spec.ClusterUris {
		kcsb := kusto.NewConnectionStringBuilder(cluster).WithDefaultAzureCredential()
		client, err := kusto.New(kcsb)
		if err != nil {
			log.Error(err, "Failed to create Kusto Client")
			r.recorder.Eventf(storedFunction, corev1.EventTypeWarning, "Failed", "Failed to create function in cluster  %s", cluster)
			executionError = multierror.Append(executionError, err)
			continue
		}
		defer client.Close()

		funcInDB, err := kustoutils.GetFunction(ctx, client, storedFunction.Spec.DB, kustoFunc, false)
		if err != nil || !kustoFunc.Equals(funcInDB) {
			log.Info("Need to create Function")
			funcInDB, err := kustoutils.GetFunction(ctx, client, storedFunction.Spec.DB, kustoFunc, true)
			if err != nil || !kustoFunc.Equals(funcInDB) {
				log.Error(err, "Failed Creating Function")
				r.recorder.Eventf(storedFunction, corev1.EventTypeWarning, "Failed", "Failed to create function in cluster  %s", cluster)
				executionError = multierror.Append(executionError, err)
				continue

			}
			r.recorder.Eventf(storedFunction, corev1.EventTypeNormal, "Executed", "Function %s created in cluster  %s", kustoFunc.Name, cluster)
		}
		clustersDone = append(clustersDone, cluster)
	}

	storedFunction.Status.ClustersDone = clustersDone
	storedFunction.Status.Status = "Success"

	if executionError != nil {
		storedFunction.Status.Status = "Failed"
	}

	err = r.Status().Update(ctx, storedFunction)
	if err != nil {
		executionError = multierror.Append(executionError, err)
	}
	if executionError != nil {
		log.Error(executionError, "failed updating stored function status", "request", req.String())
		return ctrl.Result{RequeueAfter: 10 * time.Minute}, executionError
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StoredFunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("StoredFunction")
	return ctrl.NewControllerManagedBy(mgr).
		For(&kustov1alpha1.StoredFunction{}).
		Complete(r)
}
