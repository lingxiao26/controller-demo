package main

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	queue   workqueue.RateLimitingInterface
	indexer cache.Indexer
}

func newController(q workqueue.RateLimitingInterface, indexer cache.Indexer) *controller {
	return &controller{
		queue:   q,
		indexer: indexer,
	}
}

func (c *controller) Run(stopCh chan struct{}) {
	fmt.Println("Start Controller...")
	go wait.Until(c.worker, 2, stopCh)
}

func (c *controller) worker() {
	for c.processItem() {
	}
}

func (c *controller) processItem() bool {
	k, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Forget(k)

	// 业务逻辑
	fmt.Printf("process %s\n", k.(string))

	return true
}
