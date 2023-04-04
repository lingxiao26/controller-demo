package main

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

func initClientset() kubernetes.Interface {
	homedir, _ := os.UserHomeDir()
	kubeconfig := filepath.Join(homedir, ".kube", "s1")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return c
}

func newResourceEventHandler(q workqueue.RateLimitingInterface) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			k, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				q.Add(k)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			k, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				q.Add(k)
			}
		},
		DeleteFunc: func(obj interface{}) {
			k, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				q.Add(k)
			}
		},
	}
}

func main() {
	clientset := initClientset()
	factorys := informers.NewSharedInformerFactoryWithOptions(
		clientset, 0, informers.WithNamespace("default"),
	)
	informer := factorys.Apps().V1().Deployments().Informer()
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	informer.AddEventHandler(newResourceEventHandler(q))
	indexer := informer.GetIndexer()

	c := newController(q, indexer)

	// run controller when informer has synced
	stopCh := make(chan struct{})
	defer close(stopCh)
	// run informer
	factorys.Start(stopCh)

	if cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		c.Run(stopCh)
	}

	<-stopCh
}
