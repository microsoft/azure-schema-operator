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
	schemav1alpha1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
	clusterUtils "github.com/microsoft/azure-schema-operator/pkg/cluster"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	clusterStatusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "schemaop",
		Subsystem: "cluster_executer",
		Name:      "status",
		Help:      "The execution status of the cluster for a given version, 1-success 0-fail",
	},
		[]string{"cluster", "version"},
	)
	clusterSuccessTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "schemaop",
		Subsystem: "cluster_executer",
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

// ClusterExecuterReconciler reconciles a ClusterExecuter object
type ClusterExecuterReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=clusterexecuters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=clusterexecuters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=clusterexecuters/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterExecuter object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *ClusterExecuterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("ClusterExecuter", req.NamespacedName)

	// TODO: change to parameter
	maxFailures := 3

	executer := &schemav1alpha1.ClusterExecuter{}
	err := r.Get(ctx, req.NamespacedName, executer)
	if err != nil {
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	annotations := executer.GetAnnotations()
	if val, ok := annotations["lock"]; ok {
		if strings.ToLower(val) == "true" {
			log.Info("executer locked - done")
			return ctrl.Result{}, nil
		}
	}

	if executer.Status.Running {
		log.Info("executer already Running - wait patiently")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	if executer.Status.Failed && executer.Status.NumFailures > maxFailures {
		log.Info("executer max failure retries exhosted")
		return ctrl.Result{Requeue: false}, fmt.Errorf("max retries exhosted")
	}

	notifier := func(pct int) {
		executer.Status.CompletedPCT = pct
		err = r.Status().Update(ctx, executer)
		if err != nil {
			log.Error(err, "Failed to update execution PCT ", "completed", pct, "cluster", executer.Spec.ClusterUri)
		}
	}

	cluster := clusterUtils.NewCluster(executer.Spec.Type, executer.Spec.ClusterUri, r.Client, notifier)
	targets, err := cluster.AquireTargets(executer.Spec.ApplyTo)
	if err != nil {
		log.Error(err, "failed retriving targets from cluster", "request", req.String())
		return ctrl.Result{}, err
	}

	if executer.Status.Executed {
		log.Info("executer already done - comparing db list")
		if reflect.DeepEqual(targets, executer.Status.Targets) {
			log.Info("targets already executed - returning")
			return ctrl.Result{}, nil
		}
		log.Info("targets changed - re-running")
		executer.Status.Targets = targets
		executer.Status.Running = true
		executer.Status.Executed = false
		executer.Status.Failed = false
		err = r.Status().Update(ctx, executer)
		if err != nil {
			log.Error(err, "failed updating executer status due to db list chahnge", "request", req.String())
			return ctrl.Result{}, err
		}
	}

	log.Info("Executer Controller: getting cfgMap")
	cfgMap := &v1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName(executer.Spec.ConfigMapName), cfgMap)
	if err != nil {
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		return ctrl.Result{}, err
	}

	// Filter out targers already executed
	targetsToRun := clusterUtils.Difference(targets, executer.Status.DoneTargets)
	execConfiguration, err := cluster.CreateExecConfiguration(targetsToRun, cfgMap, executer.Spec.FailIfDataLoss)
	if err != nil {
		log.Error(err, "failed creating delta-kusto configuration", "request", req.String())
		return ctrl.Result{}, err
	}
	// log.Info("Config file generated: ", "file-name", deltaCfgFile)
	executer.Status.Targets = targets
	executer.Status.Running = true
	executer.Status.Config = execConfiguration
	err = r.Status().Update(ctx, executer)
	if err != nil {
		log.Error(err, "failed updating executer status", "request", req.String())
		return ctrl.Result{}, err
	}

	// your logic here
	r.recorder.Event(executer, v1.EventTypeNormal, "Started", "cluster executer started")
	// log.Info("running : ", "file-name", deltaCfgFile)
	_, err = cluster.Execute(targetsToRun, execConfiguration)

	if err != nil {
		log.Error(err, "failed executing the schema on the cluster")
		clusterStatusGauge.WithLabelValues(clusterUtils.ClusterNameFromURI(executer.Spec.ClusterUri), strconv.Itoa(int(executer.Spec.Revision))).Set(0)
		r.recorder.Eventf(executer, v1.EventTypeWarning, "Failed", "failed to execute cluster: %s ", executer.Spec.ClusterUri)
		meta.SetStatusCondition(&executer.Status.Conditions, metav1.Condition{
			Type:    schemav1alpha1.ConditionExecution,
			Status:  metav1.ConditionFalse,
			Reason:  "Failed",
			Message: err.Error(),
		})
		executer.Status.Executed = false
		executer.Status.Running = false
		executer.Status.Failed = true
		executer.Status.NumFailures = executer.Status.NumFailures + 1
		err = r.Status().Update(ctx, executer)
		if err != nil {
			log.Error(err, "failed updating executer status", "request", req.String())
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	clusterStatusGauge.WithLabelValues(clusterUtils.ClusterNameFromURI(executer.Spec.ClusterUri), strconv.Itoa(int(executer.Spec.Revision))).Set(1)
	clusterSuccessTime.WithLabelValues(clusterUtils.ClusterNameFromURI(executer.Spec.ClusterUri), strconv.Itoa(int(executer.Spec.Revision))).SetToCurrentTime()
	r.recorder.Event(executer, v1.EventTypeNormal, "Executed", "cluster executer finished")
	meta.SetStatusCondition(&executer.Status.Conditions, metav1.Condition{
		Type:   schemav1alpha1.ConditionExecution,
		Status: metav1.ConditionTrue,
		Reason: "Executed",
	})
	executer.Status.Running = false
	executer.Status.Executed = true
	executer.Status.DoneTargets = executer.Status.Targets

	err = r.Status().Update(ctx, executer)
	if err != nil {
		log.Error(err, "failed updating executer status", "request", req.String())
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterExecuterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("ClusterExecuter")
	return ctrl.NewControllerManagedBy(mgr).
		For(&schemav1alpha1.ClusterExecuter{}).
		Complete(r)
}
