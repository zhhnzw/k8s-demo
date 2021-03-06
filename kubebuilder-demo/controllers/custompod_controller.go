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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"

	zhhnzwv1 "github.com/zhhnzw/k8s-demo/kubebuilder-demo/api/v1"
)

// CustomPodReconciler reconciles a CustomPod object
type CustomPodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=zhhnzw.mock.com,resources=custompods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=zhhnzw.mock.com,resources=custompods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=zhhnzw.mock.com,resources=custompods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CustomPod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *CustomPodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)

	// your logic here
	logr.Info("entry")
	instance := &zhhnzwv1.CustomPod{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// ???????????????
			logr.Error(err, "IsNotFound,???????????????")
			return ctrl.Result{}, err
		}
		logr.Error(err, "??????????????????")
		return ctrl.Result{}, err
	}
	logr.Info(instance.ObjectMeta.Namespace)
	logr.Info("1...")
	// ?????? kube-apiserver ????????????Pod??????
	ls := labels.Set{
		"app": instance.Name,
	}
	existingPods := &corev1.PodList{}
	err = r.Client.List(ctx, existingPods, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(ls), // ???????????????label??????????????????Pod
		Namespace:     req.Namespace,
	})
	if err != nil {
		logr.Error(err, "???????????????Pod??????")
		return ctrl.Result{}, err
	}
	logr.Info(strconv.Itoa(len(existingPods.Items)))
	logr.Info("2...")
	// ????????????Pod?????????????????????Pod?????????
	existingPodNames := make([]string, 0, instance.Spec.Replicas)
	for _, pod := range existingPods.Items {
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			// ????????????Pod?????????????????????
			continue
		}
		// pod.Status.Phase ??? Pod ???????????????
		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {
			existingPodNames = append(existingPodNames, pod.GetObjectMeta().GetName())
		}
	}

	logr.Info("3...")
	// ??? kube-apiserver ????????????Pod??????????????????Pod??????????????????????????????
	actualState := zhhnzwv1.CustomPodStatus{
		Replicas: len(existingPodNames),
		PodNames: existingPodNames,
	}
	if !reflect.DeepEqual(instance.Status, actualState) {
		logr.Info("??????Pod...")
		instance.Status = actualState
		err := r.Client.Status().Update(ctx, instance, &client.UpdateOptions{})
		if err != nil {
			logr.Error(err, "??????Pod??????")
			return ctrl.Result{}, err
		}
		logr.Info("??????Pod??????!")
		return ctrl.Result{}, nil
	}

	logr.Info("4...")
	logr.Info(strconv.Itoa(instance.Spec.Replicas))
	logr.Info(strconv.Itoa(len(existingPodNames)))
	// ???????????????pod???????????????????????????????????????scale down???delete
	if len(existingPodNames) > instance.Spec.Replicas {
		logr.Info("scale down...")
		for i := 0; i < len(existingPodNames)-instance.Spec.Replicas; i++ {
			pod := existingPods.Items[i]
			err := r.Client.Delete(ctx, &pod)
			if err != nil {
				logr.Error(err, "??????Pod??????")
				return ctrl.Result{}, err
			}
		}
		logr.Info("scale down success!")
		return ctrl.Result{}, nil
	}

	logr.Info("5...")
	// ???????????????pod???????????????????????????????????????scale up???create
	if len(existingPodNames) < instance.Spec.Replicas {
		logr.Info(fmt.Sprintf("creating pod, current and expected num: %d %d", len(existingPodNames), instance.Spec.Replicas))
		logr.Info("scale up...")
		pod := newPodForCR(instance)
		if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
			logr.Error(err, "scale up failed: SetControllerReference")
			return ctrl.Result{}, err
		}
		for i := 0; i < instance.Spec.Replicas-len(existingPodNames); i++ {
			err := r.Client.Create(ctx, pod, &client.CreateOptions{})
			if err != nil {
				logr.Error(err, "??????Pod??????")
				return ctrl.Result{}, err
			}
		}
		logr.Info("scale up success!")
		return ctrl.Result{}, nil
	}
	logr.Info("end...")
	return ctrl.Result{}, nil
}

func newPodForCR(cr *zhhnzwv1.CustomPod) *corev1.Pod {
	ls := map[string]string{"app": cr.Name}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cr.Name + "-pod-",
			Namespace:    cr.Namespace,
			Labels:       ls,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomPodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&zhhnzwv1.CustomPod{}).
		Complete(r)
}
