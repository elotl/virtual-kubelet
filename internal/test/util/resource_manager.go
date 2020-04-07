package util

import (
	"time"

	"github.com/virtual-kubelet/virtual-kubelet/internal/manager"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

// FakeResourceManager returns an instance of the resource manager that will return the specified objects when its "GetX" methods are called.
// Objects can be any valid Kubernetes object (corev1.Pod, corev1.ConfigMap, corev1.Secret, ...).
func FakeResourceManager(objects ...runtime.Object) *manager.ResourceManager {
	pInformer, mInformer, sInformer, svcInformer := MakeFakeInformers(objects...)

	// Create a new instance of the resource manager using the listers for pods, configmaps and secrets.
	r, err := manager.NewResourceManager(pInformer.Lister(), sInformer.Lister(), mInformer.Lister(), svcInformer.Lister())
	if err != nil {
		panic(err)
	}
	return r
}

// MakeFakeInformers creates informers for the subset of resource types virtual kubelet nees to work with
func MakeFakeInformers(objects ...runtime.Object) (corev1informers.PodInformer, corev1informers.ConfigMapInformer, corev1informers.SecretInformer, corev1informers.ServiceInformer) {
	// Create a fake Kubernetes client that will list the specified objects.
	kubeClient := fake.NewSimpleClientset(objects...)
	// Create a shared informer factory from where we can grab informers and listers for pods, configmaps, secrets and services.
	kubeInformerFactory := informers.NewSharedInformerFactory(kubeClient, 30*time.Second)
	// Grab informers for pods, configmaps and secrets.
	pInformer := kubeInformerFactory.Core().V1().Pods()
	mInformer := kubeInformerFactory.Core().V1().ConfigMaps()
	sInformer := kubeInformerFactory.Core().V1().Secrets()
	svcInformer := kubeInformerFactory.Core().V1().Services()
	// Start all the required informers.
	go pInformer.Informer().Run(wait.NeverStop)
	go mInformer.Informer().Run(wait.NeverStop)
	go sInformer.Informer().Run(wait.NeverStop)
	go svcInformer.Informer().Run(wait.NeverStop)
	// Wait for the caches to be synced.
	if !cache.WaitForCacheSync(wait.NeverStop, pInformer.Informer().HasSynced, mInformer.Informer().HasSynced, sInformer.Informer().HasSynced, svcInformer.Informer().HasSynced) {
		panic("failed to wait for caches to be synced")
	}
	return pInformer, mInformer, sInformer, svcInformer
}
