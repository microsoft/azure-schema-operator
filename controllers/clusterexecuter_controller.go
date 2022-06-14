// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
	clusterUtils "github.com/microsoft/azure-schema-operator/pkg/cluster"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	clusterStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "schemaop",
		Subsystem: "cluster_executor",
		Name:      "status",
		Help:      "The execution status of the cluster for a given version, 1-success 0-fail",
	},
		[]string{"cluster", "version"},
	)
	clusterSuccessTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "schemaop",
		Subsystem: "cluster_executor",
		Name:      "success_time",
		Help:      "The time of success for a cluster for a given version.",
	},
		[]string{"cluster", "version"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(clusterStatusGauge, clusterSuccessTime)
}

// ClusterExecutorReconciler reconciles a ClusterExecutor object
type ClusterExecutorReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=clusterexecutors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=clusterexecutors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=clusterexecutors/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterExecutor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *ClusterExecutorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("ClusterExecutor", req.NamespacedName)

	// TODO: change to parameter
	maxFailures := 3

	executor := &schemav1alpha1.ClusterExecutor{}
	err := r.Get(ctx, req.NamespacedName, executor)
	if err != nil {
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	annotations := executor.GetAnnotations()
	if val, ok := annotations["lock"]; ok {
		if strings.ToLower(val) == "true" {
			log.Info("Executor locked - done")
			return ctrl.Result{}, nil
		}
	}

	if executor.Status.Running {
		log.Info("Executor already Running - wait patiently")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	if executor.Status.Failed && executor.Status.NumFailures > maxFailures {
		log.Info("Executor max failure retries exhosted")
		return ctrl.Result{Requeue: false}, fmt.Errorf("Max retries exhausted")
	}

	notifier := func(pct int) {
		executor.Status.CompletedPCT = pct
		err = r.Status().Update(ctx, executor)
		if err != nil {
			log.Error(err, "Failed to update execution PCT ", "completed", pct, "cluster", executor.Spec.ClusterUri)
		}
	}

	cluster := clusterUtils.NewCluster(executor.Spec.Type, executor.Spec.ClusterUri, r.Client, notifier)
	targets, err := cluster.AquireTargets(executor.Spec.ApplyTo)
	if err != nil {
		log.Error(err, "Failed retriving targets from cluster", "request", req.String())
		return ctrl.Result{}, err
	}

	if executor.Status.Executed {
		log.Info("Rxecutor already done - comparing db list")
		if reflect.DeepEqual(targets, executor.Status.Targets) {
			log.Info("Targets already executed - returning")
			return ctrl.Result{}, nil
		}
		log.Info("targets changed - re-running")
		executor.Status.Targets = targets
		executor.Status.Running = true
		executor.Status.Executed = false
		executor.Status.Failed = false
		err = r.Status().Update(ctx, executor)
		if err != nil {
			log.Error(err, "Failed updating executor status due to db list chahnge", "request", req.String())
			return ctrl.Result{}, err
		}
	}

	log.Info("Executor Controller: getting cfgMap")
	cfgMap := &v1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName(executor.Spec.ConfigMapName), cfgMap)
	if err != nil {
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		return ctrl.Result{}, err
	}

	// Filter out targers already executed
	targetsToRun := clusterUtils.Difference(targets, executor.Status.DoneTargets)
	execConfiguration, err := cluster.CreateExecConfiguration(targetsToRun, cfgMap, executor.Spec.FailIfDataLoss)
	if err != nil {
		log.Error(err, "Failed creating delta-kusto configuration", "request", req.String())
		return ctrl.Result{}, err
	}
	// log.Info("Config file generated: ", "file-name", deltaCfgFile)
	executor.Status.Targets = targets
	executor.Status.Running = true
	executor.Status.Config = execConfiguration
	err = r.Status().Update(ctx, executor)
	if err != nil {
		log.Error(err, "Failed updating executor status", "request", req.String())
		return ctrl.Result{}, err
	}

	// your logic here
	r.recorder.Event(executor, v1.EventTypeNormal, "Started", "Cluster executor started")
	// log.Info("running : ", "file-name", deltaCfgFile)
	_, err = cluster.Execute(targetsToRun, execConfiguration)

	if err != nil {
		log.Error(err, "Failed executing the schema on the cluster")
		clusterStatusGauge.WithLabelValues(clusterUtils.ClusterNameFromURI(executor.Spec.ClusterUri), strconv.Itoa(int(executor.Spec.Revision))).Set(0)
		r.recorder.Eventf(executor, v1.EventTypeWarning, "Failed", "Failed to execute cluster: %s ", executor.Spec.ClusterUri)
		meta.SetStatusCondition(&executor.Status.Conditions, metav1.Condition{
			Type:    schemav1alpha1.ConditionExecution,
			Status:  metav1.ConditionFalse,
			Reason:  "Failed",
			Message: err.Error(),
		})
		executor.Status.Executed = false
		executor.Status.Running = false
		executor.Status.Failed = true
		executor.Status.NumFailures = executor.Status.NumFailures + 1
		err = r.Status().Update(ctx, executor)
		if err != nil {
			log.Error(err, "Failed updating executor status", "request", req.String())
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	clusterStatusGauge.WithLabelValues(clusterUtils.ClusterNameFromURI(executor.Spec.ClusterUri), strconv.Itoa(int(executor.Spec.Revision))).Set(1)
	clusterSuccessTime.WithLabelValues(clusterUtils.ClusterNameFromURI(executor.Spec.ClusterUri), strconv.Itoa(int(executor.Spec.Revision))).SetToCurrentTime()
	r.recorder.Event(executor, v1.EventTypeNormal, "Executed", "Cluster executor finished")
	meta.SetStatusCondition(&executor.Status.Conditions, metav1.Condition{
		Type:   schemav1alpha1.ConditionExecution,
		Status: metav1.ConditionTrue,
		Reason: "Executed",
	})
	executor.Status.Running = false
	executor.Status.Executed = true
	executor.Status.DoneTargets = executor.Status.Targets

	err = r.Status().Update(ctx, executor)
	if err != nil {
		log.Error(err, "Failed updating executor status", "request", req.String())
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterExecutorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("ClusterExecutor")
	return ctrl.NewControllerManagedBy(mgr).
		For(&schemav1alpha1.ClusterExecutor{}).
		Complete(r)
}
