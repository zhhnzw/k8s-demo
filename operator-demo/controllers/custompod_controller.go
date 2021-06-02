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

	zhhnzwv1 "github.com/zhhnzw/operator-demo/v1/api/v1"
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
			// 没取到实例
			logr.Error(err, "IsNotFound,没取到实例")
			return ctrl.Result{}, err
		}
		logr.Error(err, "获取实例失败")
		return ctrl.Result{}, err
	}
	logr.Info(instance.ObjectMeta.Namespace)
	logr.Info("1...")
	// 获取 kube-apiserver 已存在的Pod列表
	ls := labels.Set{
		"app": instance.Name,
	}
	existingPods := &corev1.PodList{}
	err = r.Client.List(ctx, existingPods, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(ls), // 通过资源的label来查找对应的Pod
		Namespace:     req.Namespace,
	})
	if err != nil {
		logr.Error(err, "取已存在的Pod失败")
		return ctrl.Result{}, err
	}
	logr.Info(strconv.Itoa(len(existingPods.Items)))
	logr.Info("2...")
	// 从上一步Pod列表中拿到对应Pod的名称
	var existingPodNames []string
	for _, pod := range existingPods.Items {
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			// 说明这个Pod被删除了，跳过
			continue
		}
		// pod.Status.Phase 是 Pod 的运行状态
		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {
			existingPodNames = append(existingPodNames, pod.GetObjectMeta().GetName())
		}
	}

	logr.Info("3...")
	// 从 kube-apiserver 已存在的Pod状态与期望的Pod状态对比，不相同就改
	actualState := zhhnzwv1.CustomPodStatus{
		Replicas: len(existingPodNames),
		PodNames: existingPodNames,
	}
	if !reflect.DeepEqual(instance.Status, actualState) {
		logr.Info("更新Pod...")
		instance.Status = actualState
		err := r.Client.Status().Update(ctx, instance, &client.UpdateOptions{})
		if err != nil {
			logr.Error(err, "更新Pod失败")
			return ctrl.Result{}, err
		}
		logr.Info("更新Pod成功!")
		return ctrl.Result{}, nil
	}

	logr.Info("4...")
	logr.Info(strconv.Itoa(instance.Spec.Replicas))
	logr.Info(strconv.Itoa(len(existingPodNames)))
	// 如果实际的pod副本数量比期望的副本要多，scale down，delete
	if len(existingPodNames) > instance.Spec.Replicas {
		logr.Info("scale down...")
		for i := 0; i < len(existingPodNames)-instance.Spec.Replicas; i++ {
			pod := existingPods.Items[i]
			err := r.Client.Delete(ctx, &pod)
			if err != nil {
				logr.Error(err, "删除Pod失败")
				return ctrl.Result{}, err
			}
		}
		logr.Info("scale down success!")
		return ctrl.Result{}, nil
	}

	logr.Info("5...")
	// 如果实际的pod副本数量比期望的副本要少，scale up，create
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
				logr.Error(err, "创建Pod失败")
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
			GenerateName: cr.Name + "-pod",
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
