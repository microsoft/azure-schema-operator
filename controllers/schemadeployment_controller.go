// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// telemetry "github.com/Azure/azure-service-operator/pkg/telemetry"
	"github.com/go-logr/logr"
	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/utils/schemaversions"
	"github.com/rs/zerolog/log"
)

// SchemaDeploymentReconciler reconciles a SchemaDeployment object
type SchemaDeploymentReconciler struct {
	client.Client
	// Telemetry telemetry.TelemetryClient
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=schemadeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=schemadeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbschema.microsoft.com,resources=schemadeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;update;create;patch;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SchemaDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *SchemaDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("SchemaDeployment", req.NamespacedName)
	log.Info("SchemaDeploymentReconciler - start ")
	template := &schemav1alpha1.SchemaDeployment{}
	err := r.Get(ctx, req.NamespacedName, template)
	if err != nil {
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Start logic here...

	//a. get configMap to file
	cfgMap := &corev1.ConfigMap{}

	err = r.Get(ctx, types.NamespacedName(template.Spec.Source), cfgMap)
	if err != nil {
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		log.Error(err, "Failed to fetch the configMap")
		return ctrl.Result{}, err
	}
	// Set template instance as the owner and controller of the configMap
	err = ctrl.SetControllerReference(template, cfgMap, r.Scheme)
	if err != nil {
		log.Error(err, "Failed to set the cfgMap ownership", "Namespace", cfgMap.Namespace, "Name", cfgMap.Name)
		return ctrl.Result{}, err
	}
	err = r.Update(ctx, cfgMap)
	if err != nil {
		log.Error(err, "Failed to update the cfgMap ownership", "Namespace", cfgMap.Namespace, "Name", cfgMap.Name)
		return ctrl.Result{}, err
	}
	if template.Status.CurrentConfigMap.Name == "" {
		log.Info("First run - revision 0")
		template.Status.CurrentRevision = 0
	} else if !r.compareConfigMap(ctx, template.Status.CurrentConfigMap, cfgMap) {
		log.Info("the config map changed - increase revision")
		template.Status.CurrentRevision = template.Status.CurrentRevision + 1
	} else {
		log.Info("the config map remained the same - do nothing", "revision", template.Status.CurrentRevision)
	}
	versionedDeplymentName := template.Name + "-" + strconv.Itoa(int(template.Status.CurrentRevision))

	// Check if the versioned deployment already exists, if not create a new one
	log.Info("SchemaDeploymentReconciler - start versioned deployment")
	versionedDeployment := &schemav1alpha1.VersionedDeplyment{}
	err = r.Get(ctx, types.NamespacedName{Name: versionedDeplymentName, Namespace: template.Namespace}, versionedDeployment)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Versioned deployment and immutable config map")
		// Define a new deployment
		imm := true
		verCfgMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      schemaversions.NameForConfigMap(template.Spec.Source.Name, template.Status.CurrentRevision),
				Namespace: template.Spec.Source.Namespace,
			},
			Data:       cfgMap.Data,
			BinaryData: cfgMap.BinaryData,
			Immutable:  &imm,
		}
		err = r.Create(ctx, verCfgMap)
		if err != nil {
			log.Error(err, "Failed to create new versioned cfgMap", "Namespace", verCfgMap.Namespace, "Name", verCfgMap.Name)
			return ctrl.Result{}, err
		}

		dep := &schemav1alpha1.VersionedDeplyment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        versionedDeplymentName,
				Namespace:   template.Namespace,
				Annotations: make(map[string]string),
			},
			Spec: schemav1alpha1.VersionedDeplymentSpec{

				Revision: template.Status.CurrentRevision,
				ConfigMapName: schemav1alpha1.NamespacedName{
					Name:      schemaversions.NameForConfigMap(template.Spec.Source.Name, template.Status.CurrentRevision),
					Namespace: template.Namespace,
				},
				ApplyTo:        template.Spec.ApplyTo,
				Type:           template.Spec.Type,
				FailIfDataLoss: template.Spec.FailIfDataLoss,
			},
		}
		// Set template instance as the owner and controller
		err = ctrl.SetControllerReference(template, dep, r.Scheme)
		if err != nil {
			log.Error(err, "Failed to set controller reference for the new versioned Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new versioned Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}

		// template.Status.Executed = true // TO-DO: compute from the running versioned deployment.
		if template.Status.CurrentVerDeployment.Name != "" {
			err = r.lockVersionedDeployment(ctx, template.Status.CurrentVerDeployment)
			if err != nil {
				log.Error(err, "Failed to lock the previous versioned Deployment", "Deployment.Namespace", template.Status.CurrentVerDeployment.Namespace, "Deployment.Name", template.Status.CurrentVerDeployment.Name)
				return ctrl.Result{}, err
			}
			template.Status.OldVerDeployment = append(template.Status.OldVerDeployment, template.Status.CurrentVerDeployment)
		}
		template.Status.CurrentVerDeployment = schemav1alpha1.NamespacedName{
			Name:      versionedDeplymentName,
			Namespace: template.Namespace,
		}
		template.Status.CurrentConfigMap = schemav1alpha1.NamespacedName{
			Name:      schemaversions.NameForConfigMap(template.Spec.Source.Name, template.Status.CurrentRevision),
			Namespace: template.Namespace,
		}

		// template.Status = status
		err = r.Status().Update(ctx, template)
		if err != nil {
			log.Error(err, "failed updating status", "request", req.String())
			return ctrl.Result{}, err
		}

		// Deployment created successfully - return and requeue
		// return ctrl.Result{Requeue: true}, nil
		log.Info("Versioned Deployment created successfully - return and requeue")
		r.recorder.Eventf(template, corev1.EventTypeNormal, "Created", "Created versioned deployment %q", dep.Name)
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil

	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	log.Info("Checking if template needs to update versioned deployment")
	changed, err := r.compareAndUpdateVersionedDeployment(ctx, template, versionedDeployment)
	if err != nil {
		log.Error(err, "Failed to compare and update versioned deployment")
		return ctrl.Result{}, err
	}
	if changed {
		log.Info("Versioned Deployment changed successfully - return and requeue")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	log.Info("Checking versioned deployment status")
	err = r.Get(ctx, req.NamespacedName, template)
	if err != nil {
		log.Error(err, "failed to refresh template", "request", req.String())
		return ctrl.Result{}, err
	}

	if versionedDeployment.IsExecuted() {
		template.Status.Executed = true // don't override - maybe need refresh?
		template.Status.LastSuccessfulRevision = template.Status.CurrentRevision

		// template.Status = status

		r.recorder.Eventf(template, corev1.EventTypeNormal, "Executed", "Scheme was deployed")
		meta.SetStatusCondition(&template.Status.Conditions, metav1.Condition{
			Type:   schemav1alpha1.ConditionExecution,
			Status: metav1.ConditionTrue,
			Reason: "Executed",
		})
		err = r.Status().Update(ctx, template)
		if err != nil {
			log.Error(err, "failed updating status to executed", "request", req.String())
			return ctrl.Result{}, err
		}
	} else if versionedDeployment.IsRunning() {
		log.Info("Still running - wait more")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	} else if versionedDeployment.IsFailed() {
		log.Info("Failed to execute schema change")

		r.recorder.Eventf(template, corev1.EventTypeWarning, "Failed", "failed to deploy schema ")
		meta.SetStatusCondition(&template.Status.Conditions, metav1.Condition{
			Type:    schemav1alpha1.ConditionExecution,
			Status:  metav1.ConditionFalse,
			Reason:  "Failed",
			Message: "Schema execution failure",
		})
		err = r.Status().Update(ctx, template)
		if err != nil {
			log.Error(err, "failed updating status ", "request", req.String())
			return ctrl.Result{}, err
		}
		return r.handleFailure(template)
	}

	log.Info("exiting reconciliation")
	return ctrl.Result{}, err
}

func (r *SchemaDeploymentReconciler) compareConfigMap(ctx context.Context, currentConfigMap schemav1alpha1.NamespacedName, cfgMap *corev1.ConfigMap) bool {
	if currentConfigMap.Name == "" {
		log.Info().Msg("current Map is empty - new template.")
		return false
	}
	currCfgMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Namespace: currentConfigMap.Namespace, Name: currentConfigMap.Name}, currCfgMap)
	if err != nil {
		// r.Telemetry.LogInfoByInstance("ignorable error", "error during fetch from api server", req.String())
		log.Info().Msg("fail to find the current config map")
		return false
	}
	log.Info().Str("curr", currCfgMap.Data["kql"]).Str("new", cfgMap.Data["kql"]).Msg("Compare kql strings")

	return (reflect.DeepEqual(currCfgMap.Data, cfgMap.Data) && reflect.DeepEqual(currCfgMap.BinaryData, cfgMap.BinaryData))
}

func (r *SchemaDeploymentReconciler) compareAndUpdateVersionedDeployment(ctx context.Context, template *schemav1alpha1.SchemaDeployment, deployment *schemav1alpha1.VersionedDeplyment) (bool, error) {
	var err error
	changed := false
	if !reflect.DeepEqual(template.Spec.ApplyTo, deployment.Spec.ApplyTo) {
		deployment.Spec.ApplyTo = template.Spec.ApplyTo
		changed = true
	}
	if template.Spec.FailIfDataLoss != deployment.Spec.FailIfDataLoss {
		deployment.Spec.FailIfDataLoss = template.Spec.FailIfDataLoss
		changed = true
	}

	if changed {
		err = r.Update(ctx, deployment)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to update versioned deployment with chnages to the template")
		}
	}
	return changed, err
}

// func (r *SchemaDeploymentReconciler) checkExecutionStatus(template *schemav1alpha1.SchemaDeployment, deployment *schemav1alpha1.VersionedDeplyment) (bool, error) {
// 	return deployment.IsExecuted(), nil
// }

func (r *SchemaDeploymentReconciler) handleFailure(template *schemav1alpha1.SchemaDeployment) (ctrl.Result, error) {
	switch policy := template.Spec.FailurePolicy; policy {
	case schemav1alpha1.FailurePolicyAbort:
		log.Info().Msg("handling failure - abort policy.")
	case schemav1alpha1.FailurePolicyRollback:
		log.Info().Msg("handling failure - rollback policy.")
		if template.Status.CurrentRevision == 0 {
			return ctrl.Result{}, fmt.Errorf("on first revision - no where back to go")
		}
		err := schemaversions.RollbackToVersion(r.Client, template, template.Status.LastSuccessfulRevision)
		if err != nil {
			log.Error().Err(err).Msg("Failed to rollback source schema")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	case schemav1alpha1.FailurePolicyIgnore:
		log.Info().Msg("handling failure - ignore policy.")
	default:
		log.Error().Msg("got unsupported policy")
		return ctrl.Result{}, fmt.Errorf("unsupported policy")
	}
	return ctrl.Result{}, nil
}

func (r *SchemaDeploymentReconciler) lockVersionedDeployment(ctx context.Context, deploymentName schemav1alpha1.NamespacedName) error {
	log := r.Log
	deployment := &schemav1alpha1.VersionedDeplyment{}
	err := r.Get(ctx, types.NamespacedName(deploymentName), deployment)
	if err != nil {
		log.Error(err, "failed to find the current versioned deployment")
		return err
	}
	res, err := meta.Accessor(deployment)
	if err != nil {
		log.Error(err, "failed to find the current versioned deployment")
		return err
	}
	annotations := res.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["lock"] = "true"
	res.SetAnnotations(annotations)
	err = r.Update(ctx, deployment)
	if err != nil {
		log.Error(err, "failed to update the versioned deployment annotations")
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SchemaDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("SchemaDeployment")
	return ctrl.NewControllerManagedBy(mgr).
		For(&schemav1alpha1.SchemaDeployment{}).
		Owns(&schemav1alpha1.VersionedDeplyment{}).
		Owns(&corev1.ConfigMap{}).
		// WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
