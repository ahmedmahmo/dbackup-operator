/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	batchv1 "github.com/ahmedmahmo/discovery-operator/api/v1"
	kubebatchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DbackupReconciler reconciles a Dbackup object
type DbackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=batch.k8s.htw-berlin.de,resources=dbackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.k8s.htw-berlin.de,resources=dbackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.k8s.htw-berlin.de,resources=dbackups/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

func (r *DbackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)

	var dbackup batchv1.Dbackup
	log.Info("reconciling")
	if err := r.Get(ctx, req.NamespacedName, &dbackup); err != nil {
		log.Error(err, "unable to fetch CronJob")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var jobs kubebatchv1.JobList
	if err := r.List(ctx, &jobs, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "unable to list Jobs")
		return ctrl.Result{}, err
	}

	log.V(1).Info(dbackup.Name)
	job, err := CreateBackupJob(&dbackup)
	if err != nil {
		log.Error(err, "unable to create job from teplate")

	}

	if err := r.Create(ctx, job); err != nil {
		log.Error(err, "unable to create Job for Dbackup", "job", job)
		return ctrl.Result{}, err
	}

	log.V(1).Info("created Job for Dbackup run", "job", job)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DbackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Dbackup{}).
		Complete(r)
}

func IsBackupFinished(job *kubebatchv1.Job) (bool, kubebatchv1.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == kubebatchv1.JobComplete || c.Type == kubebatchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}

	return false, ""
}

func CreateBackupJob(backupJob *batchv1.Dbackup) (*kubebatchv1.Job, error) {

	job := &kubebatchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupJob.Name,
			Namespace: backupJob.Namespace,
		},
		Spec: *backupJob.Spec.BackupTemplate.Spec.DeepCopy(),
	}

	return job, nil
}
