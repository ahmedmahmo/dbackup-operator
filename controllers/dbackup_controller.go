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
	"fmt"
	// "math/rand"
	"time"

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

/*
	Kubebulder reads the comments that starts 
	with +kubebuilder as annotation.

	kubebuilder will create a kubernetes manifest in config/rbac 
	for ClusterRole manifest creation.

	This manifest is the policy that allows the operator to execute some API verbs like (POST,GET)
	on behald of the user.
*/

//+kubebuilder:rbac:groups=batch.k8s.htw-berlin.de,resources=dbackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.k8s.htw-berlin.de,resources=dbackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.k8s.htw-berlin.de,resources=dbackups/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

func (r *DbackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)

	/*
	Load Dbackup Object by Name
	Using Kubernetes client the object with be fetched 
	ex. 
		apiVersion: batch.k8s.htw-berlin.de/v1
		kind: Dbackup
		metadata:
		name: dbackup-sample
	*/

	var dbackup batchv1.Dbackup
	if err := r.Get(ctx, req.NamespacedName, &dbackup); err != nil {
		log.Error(err, "unable to fetch CronJob")
		// Ignore not found objects as they can't be fixed
		// On deletation event the object will be ommited
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	/*
	List Active jobs of type job in apiVersion batch/v1
	This is a generic kubernetes object to execute runs
	*/
	var jobs kubebatchv1.JobList
	if err := r.List(ctx, &jobs, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "unable to list Jobs")
		return ctrl.Result{}, err
	}

	/*
	function stored in avariable to create a job object with the name of the Dbackup object and time unix signature
	for unique naming for each created as a Job
	*/
	createBackupJob := func (backupJob *batchv1.Dbackup, creationTime time.Time) (*kubebatchv1.Job, error)  {
		name := fmt.Sprintf("%s-%d", backupJob.Name, creationTime.Unix())
		job := &kubebatchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: backupJob.Namespace,
			},
			Spec: *backupJob.Spec.BackupTemplate.Spec.DeepCopy(),
		}
		if err := ctrl.SetControllerReference(backupJob, job, r.Scheme); err != nil {
			return nil, err
		}

		return job, nil
	}

	// create the job
	job, err := createBackupJob(&dbackup, time.Now())
	if err != nil {
		log.Error(err, "unable to create job object")
	}

	// create the job in the cluster
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

func CreateBackupJob(backupJob *batchv1.Dbackup, creationTime time.Time) (*kubebatchv1.Job, error) {
	// rand.Seed(time.Now().Unix())
	// min := 10
	// max := 30
	// random := (rand.Intn(max-min+1) + min)
	name := fmt.Sprintf("%s-%d", backupJob.Name, creationTime.Unix())
	job := &kubebatchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: backupJob.Namespace,
		},
		Spec: *backupJob.Spec.BackupTemplate.Spec.DeepCopy(),
	}

	return job, nil
}
