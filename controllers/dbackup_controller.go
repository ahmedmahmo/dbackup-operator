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
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	batchv1 "github.com/ahmedmahmo/discovery-operator/api/v1"
	cron "github.com/robfig/cron"
	kubebatchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	reference "k8s.io/client-go/tools/reference"
)

// DbackupReconciler reconciles a Dbackup object
type DbackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Time
}
type realTime struct{}

func (t realTime) Now() time.Time {
	return time.Now()
}

type Time interface {
	Now() time.Time
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

var (
	annotation = "batch.k8s.htw-berlin.de/scheduled-at"
	imageName  = "aws-runner"
	image      = "ahmedmahmoud25/dbackup-postgres-aws:master"
)

func (r *DbackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)

	/*
		Load Dbackup Object by Name
		Using Kubernetes client the object will be fetched
		ex.
			apiVersion: batch.k8s.htw-berlin.de/v1
			kind: Dbackup
			metadata:
			name: dbackup-sample
	*/

	var dbackup batchv1.Dbackup
	if err := r.Get(ctx, req.NamespacedName, &dbackup); err != nil {
		log.Error(err, "unable to fetch Dbackup Object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	/*
		List Active jobs of type job in apiVersion batch/v1
		This is a generic kubernetes object to execute runs
	*/
	var kubeJobs kubebatchv1.JobList
	if err := r.List(ctx, &kubeJobs, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "unable to list Kubernetes Jobs")
		return ctrl.Result{}, err
	}

	var activeKubeJobs []*kubebatchv1.Job
	var successfulKubeJobs []*kubebatchv1.Job
	var failedKubeJobs []*kubebatchv1.Job
	var mostRecentTime *time.Time

	// if the Kubernetes job has status of completed or failed then it finished
	didJobFinish := func(kubeJob *kubebatchv1.Job) (bool, kubebatchv1.JobConditionType) {
		for _, condition := range kubeJob.Status.Conditions {
			if (condition.Type == kubebatchv1.JobComplete || condition.Type == kubebatchv1.JobFailed) && condition.Status == corev1.ConditionTrue {
				return true, condition.Type
			}
		}
		return false, ""
	}

	/*
		Extract scheduled-at from annotation
	*/
	getScheduledTimeForJob := func(job *kubebatchv1.Job) (*time.Time, error) {
		raw := job.Annotations[annotation]
		if len(raw) == 0 {
			return nil, nil
		}

		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, err
		}
		return &parsed, nil
	}

	for i, job := range kubeJobs.Items {
		_, t := didJobFinish(&job)
		switch t {
		case "":
			activeKubeJobs = append(activeKubeJobs, &kubeJobs.Items[i])
		case kubebatchv1.JobFailed:
			failedKubeJobs = append(failedKubeJobs, &kubeJobs.Items[i])
		case kubebatchv1.JobComplete:
			successfulKubeJobs = append(successfulKubeJobs, &kubeJobs.Items[i])
		}

		scheduledTimeForJob, err := getScheduledTimeForJob(&job)
		if err != nil {
			log.Error(err, "unable to parse schedule time for job", "job", &job)
			continue
		}
		if scheduledTimeForJob != nil {
			if mostRecentTime == nil {
				mostRecentTime = scheduledTimeForJob
			} else if mostRecentTime.Before(*scheduledTimeForJob) {
				mostRecentTime = scheduledTimeForJob
			}
		}
	}

	dbackup.Status.Active = nil
	for _, activeKubeJob := range activeKubeJobs {
		jobRefrence, err := reference.GetReference(r.Scheme, activeKubeJob)
		if err != nil {
			log.Error(err, "No reference to active kubernetes job", "Kubernetes job", activeKubeJob)
			continue
		}
		dbackup.Status.Active = append(dbackup.Status.Active, *jobRefrence)
	}

	log.V(1).Info("jobs", "active kube jobs", len(activeKubeJobs),
		"successful kube jobs", len(successfulKubeJobs),
		"failed kube jobs", len(failedKubeJobs))

	/*
		Update Dbackup Status
	*/
	if err := r.Status().Update(ctx, &dbackup); err != nil {
		log.Error(err, "unable to update Dbackup status")
		return ctrl.Result{}, err
	}

	/*
		Extract next schedule based on the first creation of the a job -> var earliest
		and the givin cron specification -> cron.ParseStandard(dbackup.Spec.Schedule)
	*/
	getNextSchedule := func(dbackup *batchv1.Dbackup, now time.Time) (lastMissed time.Time, next time.Time, err error) {
		sched, err := cron.ParseStandard(dbackup.Spec.Schedule)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("Unparseable schedule %q: %v", dbackup.Spec.Schedule, err)
		}

		earliest := dbackup.ObjectMeta.CreationTimestamp.Time

		if earliest.After(now) {
			return time.Time{}, sched.Next(now), nil
		}

		for t := sched.Next(earliest); !t.After(now); t = sched.Next(t) {
			fmt.Println("t:", t)
			lastMissed = t
		}
		return lastMissed, sched.Next(now), nil
	}

	missed, next, err := getNextSchedule(&dbackup, r.Now())
	if err != nil {
		log.Error(err, "When is next schedule?")
		return ctrl.Result{}, nil
	}

	/*
		Requst a reconcile on schedule time
	*/
	result := ctrl.Result{RequeueAfter: next.Sub(r.Now())}
	log = log.WithValues("now", r.Now(), "next", next)

	if missed.IsZero() {
		log.V(1).Info("sleeping until next")
		return result, nil
	}

	/*
		Forbid deleting the running job if the policy is forbid
	*/
	if dbackup.Spec.ConcurrencyPolicy == batchv1.Forbid && len(activeKubeJobs) > 0 {
		log.V(1).Info("Policy Forbids creation", "Active Jobs", len(activeKubeJobs))
		return result, nil
	}

	/*
		if the pocliy specifices to replace running jobs in the same time
		the running job should be deleted and replaced
	*/
	if dbackup.Spec.ConcurrencyPolicy == batchv1.Replace {
		for _, activeKubeJob := range activeKubeJobs {
			if err := r.Delete(ctx, activeKubeJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
				log.Error(err, "unable to delete running kubernetes job", "job", activeKubeJob)
				return ctrl.Result{}, err
			}
		}
	}

	/*
		function stored in avariable to create a job object with the name of the Dbackup object and time unix signature
		for unique naming for each created as a Job
	*/
	createBackupJob := func(backupJob *batchv1.Dbackup, creationTime time.Time) (*kubebatchv1.Job, error) {
		name := fmt.Sprintf("%s-%d", backupJob.Name, creationTime.Unix())
		job := &kubebatchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   backupJob.Namespace,
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
			},
			Spec: kubebatchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{

						RestartPolicy: corev1.RestartPolicyOnFailure,
						Containers: []corev1.Container{
							{
								Name:            imageName,
								Image:           image,
								ImagePullPolicy: corev1.PullAlways,
								Env:             backupJob.Spec.Env,
							},
						},
					},
				},
			},
		}

		if err := ctrl.SetControllerReference(&dbackup, job, r.Scheme); err != nil {
			return nil, err
		}

		return job, nil
	}

	/*
		Create a Kubernets Job object from the given specification
	*/
	job, err := createBackupJob(&dbackup, missed)
	if err != nil {
		log.Error(err, "unable to create job object")
	}

	/*
		Reconcile to create the actaual job owned by the operaotr to run the backup
	*/
	if err := r.Create(ctx, job); err != nil {
		log.Error(err, "unable to create Job for Dbackup", "job", job)
		return ctrl.Result{}, err
	}

	log.V(1).Info("created Job for Dbackup run", "job", job)
	return result, nil
}

var apiGroupVersion = batchv1.GroupVersion.String()
var owner = ".metadata.controller"

func (r *DbackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Time = realTime{}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kubebatchv1.Job{}, owner, func(rawObj client.Object) []string {
		job := rawObj.(*kubebatchv1.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}

		if owner.APIVersion != apiGroupVersion || owner.Kind != "Dbackup" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Dbackup{}).
		Owns(&kubebatchv1.Job{}).
		Complete(r)
}
