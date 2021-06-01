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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	zhhnzwv1 "github.com/zhhnzw/operator-demo/v1/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CustomTypeReconciler reconciles a CustomType object
type CustomTypeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=zhhnzw.mock.com,resources=customtypes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=zhhnzw.mock.com,resources=customtypes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=zhhnzw.mock.com,resources=customtypes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CustomType object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *CustomTypeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("customtype", req.NamespacedName)

	// your logic here

	instance := &zhhnzwv1.CustomType{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// 没取到实例，重试
			return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, nil
		}
	}

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
		r.Log.Error(err, "取已存在的Pod失败")
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, nil
	}

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

	// 从 kube-apiserver 已存在的Pod状态与期望的Pod状态对比，不相同就改
	actualState := zhhnzwv1.CustomTypeStatus{
		Replicas: len(existingPodNames),
		PodNames: existingPodNames,
	}
	if !reflect.DeepEqual(instance.Status, actualState) {
		instance.Status = actualState
		err := r.Client.Status().Update(ctx, instance, &client.UpdateOptions{})
		if err != nil {
			r.Log.Error(err, "更新Pod失败")
			return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, nil
		}
	}

	// 如果实际的pod副本数量比期望的副本要多，scale down，delete
	if len(existingPodNames) > instance.Spec.Replicas {
		pod := existingPods.Items[0]
		err := r.Client.Delete(ctx, &pod)
		if err != nil {
			r.Log.Error(err, "删除Pod失败")
			return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, nil
		}
	}

	// 如果实际的pod副本数量比期望的副本要少，scale up，create
	if len(existingPodNames) > instance.Spec.Replicas {
		err := r.Client.Create(ctx, instance, &client.CreateOptions{})
		if err != nil {
			r.Log.Error(err, "创建Pod失败")
			return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, nil
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomTypeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&zhhnzwv1.CustomType{}).
		Complete(r)
}
