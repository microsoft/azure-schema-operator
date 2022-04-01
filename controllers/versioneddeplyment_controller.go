// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package controllers

import (
	"context"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/microsoft/azure-schema-operator/api/v1alpha1"
	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/cluster"
	"github.com/rs/zerolog/log"
)

// VersionedDeplymentReconciler reconciles a VersionedDeplyment object
type VersionedDeplymentReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=versioneddeplyments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=versioneddeplyments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=versioneddeplyments/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;update;create;patch;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VersionedDeplyment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *VersionedDeplymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("SchemaDeployment", req.NamespacedName)
	log.Info("VersionedDeplyment - start")
	versionedDeplyment := &schemav1alpha1.VersionedDeplyment{}
	err := r.Get(ctx, req.NamespacedName, versionedDeplyment)
	if err != nil {
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	annotations := versionedDeplyment.GetAnnotations()
	if val, ok := annotations["lock"]; ok {
		if strings.ToLower(val) == "true" {
			log.Info("executer locked - verifing executers are locked and done")
			r.ensureExecutersLocked(ctx, versionedDeplyment)
			return ctrl.Result{}, nil
		}
	}
	// your logic here
	log.Info("VersionedDeplyment - get cfgMap", "Namespace", versionedDeplyment.Spec.ConfigMapName.Namespace, "Name", versionedDeplyment.Spec.ConfigMapName.Name)

	cfgMap := &v1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName(versionedDeplyment.Spec.ConfigMapName), cfgMap)
	if err != nil {
		log.Error(err, "Failed to get cfgMap")
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		return ctrl.Result{RequeueAfter: 15 * time.Second}, client.IgnoreNotFound(err)
	}
	isOwned := false
	for _, ownerRef := range cfgMap.GetOwnerReferences() {
		if ownerRef.UID == versionedDeplyment.GetUID() {
			isOwned = true
		}
	}
	// Set versioned deployment the owner of the and controller of the configMap
	if !isOwned {
		log.Info("Setting ownership on cfg")
		err = ctrl.SetControllerReference(versionedDeplyment, cfgMap, r.Scheme)
		if err != nil {
			log.Error(err, "Failed to set cfgMap reference for the versioned deployment")
		}
		err = r.Update(ctx, cfgMap)
		if err != nil {
			log.Error(err, "Failed to update the cfgMap ownership", "Namespace", cfgMap.Namespace, "Name", cfgMap.Name)
			return ctrl.Result{}, err
		}
	}

	log.Info("VersionedDeplyment - check executers")
	if len(versionedDeplyment.Status.Executers) == 0 {
		versionedDeplyment.Status.Executers = make([]schemav1alpha1.NamespacedName, len(versionedDeplyment.Spec.ApplyTo.ClusterUris))
		log.Info("Creating Executers array", "length", len(versionedDeplyment.Status.Executers))
	} else if len(versionedDeplyment.Spec.ApplyTo.ClusterUris) > len(versionedDeplyment.Status.Executers) {
		log.Info("we need to extend the Executers array")
		tempSlice := make([]schemav1alpha1.NamespacedName, len(versionedDeplyment.Spec.ApplyTo.ClusterUris))
		copy(tempSlice, versionedDeplyment.Status.Executers)
		versionedDeplyment.Status.Executers = tempSlice
	} else {
		log.Info("Executers already exist with enough capacity", "length", len(versionedDeplyment.Status.Executers))
	}

	newExecuters := false
	// b. loop over all clusters defined:
	for i, uri := range versionedDeplyment.Spec.ApplyTo.ClusterUris {

		execKey := types.NamespacedName{
			Name:      versionedDeplyment.Name + "-" + cluster.ClusterNameFromURI(uri),
			Namespace: versionedDeplyment.Namespace,
		}
		found := &schemav1alpha1.ClusterExecuter{}
		err = r.Get(ctx, execKey, found)
		if err != nil && errors.IsNotFound(err) {

			// Create cluster executer objects
			ce, err := r.executerForCluster(uri, versionedDeplyment)
			if err != nil {
				log.Error(err, "Failed to create new cluster executer", "Namespace", ce.Namespace, "Name", ce.Name)
				return ctrl.Result{}, err
			}
			err = r.Create(ctx, ce)
			if err != nil {
				log.Error(err, "Failed to create new cluster executer", "Namespace", ce.Namespace, "Name", ce.Name)
				return ctrl.Result{}, err
			}
			log.Info("executer file generated: ", "cluster-uri", uri)
			r.recorder.Event(versionedDeplyment, v1.EventTypeNormal, "Created", "new cluster executers")
			versionedDeplyment.Status.Executers[i] = v1alpha1.NamespacedName{
				Namespace: execKey.Namespace,
				Name:      execKey.Name,
			}
			newExecuters = true
		} else if err != nil {
			log.Error(err, "Failed to get executer")
			return ctrl.Result{}, err
		} else {
			log.Info("Checking if versioned deployment needs to update Cluster executer")
			changed, err := r.compareAndUpdateExecuter(ctx, versionedDeplyment, found)
			if err != nil {
				log.Error(err, "Failed to compare and update Cluster executer")
				return ctrl.Result{}, err
			}
			if changed {
				log.Info("Versioned Deployment changed successfully - return and requeue")
				newExecuters = true
			}
		}

		// // This handles cases where the list order was re-arranged for some reason.
		// if !reflect.DeepEqual(versionedDeplyment.Status.Executers[i], v1alpha1.NamespacedName{
		// 	Namespace: execKey.Namespace,
		// 	Name:      execKey.Name,
		// }) {
		// 	versionedDeplyment.Status.Executers[i] = v1alpha1.NamespacedName{
		// 		Namespace: execKey.Namespace,
		// 		Name:      execKey.Name,
		// 	}
		// 	newExecuters = true
		// }

	}
	if newExecuters {
		err = r.Status().Update(ctx, versionedDeplyment)
		if err != nil {
			log.Error(err, "failed updating versionedDeplyment new executers status", "request", req.String())
			return ctrl.Result{}, err
		}
		log.Info("New executers created - requeuing after 5 sec to let them create.")
		// return ctrl.Result{RequeueAfter: time.Minute}, nil
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	log.Info("Checking executers status")
	err = r.statusCheck(ctx, versionedDeplyment)
	if err != nil {
		log.Error(err, "failed updating versionedDeplyment execution status", "request", req.String())
		return ctrl.Result{}, err
	}
	log.Info("versiond deployment done")
	return ctrl.Result{}, nil
}

func (r *VersionedDeplymentReconciler) executerForCluster(uri string, versionedDeplyment *schemav1alpha1.VersionedDeplyment) (*schemav1alpha1.ClusterExecuter, error) {
	exec := &schemav1alpha1.ClusterExecuter{
		ObjectMeta: metav1.ObjectMeta{
			Name:        versionedDeplyment.Name + "-" + cluster.ClusterNameFromURI(uri),
			Namespace:   versionedDeplyment.Namespace,
			Annotations: make(map[string]string),
		},
		Spec: schemav1alpha1.ClusterExecuterSpec{
			ClusterUri: uri,
			ApplyTo:    versionedDeplyment.Spec.ApplyTo,
			Type:       versionedDeplyment.Spec.Type,
			ConfigMapName: schemav1alpha1.NamespacedName{
				Namespace: versionedDeplyment.Spec.ConfigMapName.Namespace,
				Name:      versionedDeplyment.Spec.ConfigMapName.Name,
			},
			FailIfDataLoss: versionedDeplyment.Spec.FailIfDataLoss,
			Revision:       versionedDeplyment.Spec.Revision,
		},
		Status: schemav1alpha1.ClusterExecuterStatus{},
	}
	err := ctrl.SetControllerReference(versionedDeplyment, exec, r.Scheme)
	return exec, err
}
func (r *VersionedDeplymentReconciler) ensureExecutersLocked(ctx context.Context, versionedDeplyment *schemav1alpha1.VersionedDeplyment) {
	for _, executer := range versionedDeplyment.Status.Executers {
		exec := &schemav1alpha1.ClusterExecuter{}

		err := r.Get(ctx, types.NamespacedName(executer), exec)
		if err != nil {
			log.Error().Err(err).Msg("failed to find the executer")
		}
		res, err := meta.Accessor(exec)
		if err != nil {
			log.Error().Err(err).Msg("failed to get the executer annotations.")
		}
		annotations := res.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations["lock"] = "true"
		res.SetAnnotations(annotations)
		err = r.Update(ctx, exec)
		if err != nil {
			log.Error().Err(err).Msg("failed to update the executer annotations")
		}

	}
}
func (r *VersionedDeplymentReconciler) compareAndUpdateExecuter(ctx context.Context, versionedDeplyment *schemav1alpha1.VersionedDeplyment, executer *schemav1alpha1.ClusterExecuter) (bool, error) {
	var err error
	changed := false
	if versionedDeplyment.Spec.ApplyTo.DB != executer.Spec.ApplyTo.DB {
		executer.Spec.ApplyTo.DB = versionedDeplyment.Spec.ApplyTo.DB
		changed = true
	}
	if versionedDeplyment.Spec.ApplyTo.Schema != executer.Spec.ApplyTo.Schema {
		executer.Spec.ApplyTo.Schema = versionedDeplyment.Spec.ApplyTo.Schema
		changed = true
	}

	if versionedDeplyment.Spec.FailIfDataLoss != executer.Spec.FailIfDataLoss {
		executer.Spec.FailIfDataLoss = versionedDeplyment.Spec.FailIfDataLoss
		changed = true
	}

	if changed {
		err = r.Update(ctx, executer)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to update cluster executer with the chnages to the deployment")
		}
	}
	return changed, err
}

func (r *VersionedDeplymentReconciler) statusCheck(ctx context.Context, versionedDeplyment *schemav1alpha1.VersionedDeplyment) error {
	log := r.Log.WithValues("function", "statusCheck")
	log.Info("Starting status check of all executers")
	failed := 0
	done := 0
	running := 0
	donePCT := 0
	for i, exec := range versionedDeplyment.Status.Executers {
		// TODO: check if all executers finished successfully
		log.Info("Checking executer", "i", i, "exec", exec)
		found := &schemav1alpha1.ClusterExecuter{}
		err := r.Get(ctx, types.NamespacedName{Namespace: exec.Namespace, Name: exec.Name}, found)
		if err != nil {
			failed = failed + 1
		} else {
			if found.Status.Executed {
				done = done + 1
			}
			if found.Status.Running {
				running = running + 1
			}
			if found.Status.Failed {
				failed = failed + 1
			}
			donePCT = donePCT + found.Status.CompletedPCT
		}
	}
	versionedDeplyment.Status.CompletedPCT = (donePCT / len(versionedDeplyment.Status.Executers))
	versionedDeplyment.Status.Failed = int32(failed)
	versionedDeplyment.Status.Running = int32(running)
	versionedDeplyment.Status.Succeeded = int32(done)
	versionedDeplyment.Status.Executed = (len(versionedDeplyment.Status.Executers) == int(versionedDeplyment.Status.Succeeded))

	err := r.Status().Update(ctx, versionedDeplyment)
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *VersionedDeplymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("VersionedDeplyment")
	return ctrl.NewControllerManagedBy(mgr).
		For(&schemav1alpha1.VersionedDeplyment{}).
		Owns(&schemav1alpha1.ClusterExecuter{}).
		Owns(&v1.ConfigMap{}).
		Complete(r)
}
