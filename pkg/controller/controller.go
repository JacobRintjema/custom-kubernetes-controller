package controller

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// Controller manages the lifecycle of our custom resources
type Controller struct {
	DynClient dynamic.Interface
	Resource  schema.GroupVersionResource
	Informer  cache.SharedIndexInformer
	Workqueue workqueue.RateLimitingInterface
	StopCh    chan struct{}
}

// NewController creates a new controller instance
func NewController(config *rest.Config) (*Controller, error) {
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Define the resource we want to watch
	resource := schema.GroupVersionResource{
		Group:    "myk8s.io",
		Version:  "v1",
		Resource: "samples",
	}

	controller := &Controller{
		DynClient: dynClient,
		Resource:  resource,
		Workqueue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(),
			"Samples",
		),
		StopCh: make(chan struct{}),
	}

	controller.setupInformer()
	return controller, nil
}

// SetupInformer creates and configures the informer
func (c *Controller) setupInformer() {
	c.Informer = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return c.DynClient.Resource(c.Resource).Namespace("").List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return c.DynClient.Resource(c.Resource).Namespace("").Watch(context.TODO(), options)
			},
		},
		&unstructured.Unstructured{},
		0, // No resync
		cache.Indexers{},
	)
}

// Run starts the controller and runs until stopCh is closed
func (c *Controller) Run(numWorkers int) {
	defer close(c.StopCh)
	defer c.Workqueue.ShutDown()

	// Start informer
	go c.Informer.Run(c.StopCh)

	// Wait for cache sync
	if !cache.WaitForCacheSync(c.StopCh, c.Informer.HasSynced) {
		klog.Error("Timeout waiting for cache sync")
		return
	}

	klog.InfoS("CONTROLLER: Started successfully", "status", "running")

	// Start workers
	for i := 0; i < numWorkers; i++ {
		workerID := i + 1
		go func(id int) {
			klog.InfoS("WORKER: Started", "id", id)
			wait.Until(func() { c.runWorker(id) }, time.Second, c.StopCh)
		}(workerID)
	}

	// Wait for shutdown
	<-c.StopCh
}

// runWorker processes work items continuously
func (c *Controller) runWorker(id int) {
	for c.processNextWorkItem(id) {
	}
}

// processNextWorkItem processes a single item from the workqueue
func (c *Controller) processNextWorkItem(workerID int) bool {
	// Get the next item
	obj, shutdown := c.Workqueue.Get()

	// Exit if shutting down
	if shutdown {
		klog.InfoS("WORKER: Shutting down", "id", workerID)
		return false
	}

	// Ensure we mark the item as done when we finish
	defer c.Workqueue.Done(obj)

	// Process the item
	err := func(obj interface{}) error {
		// Convert to key
		key, ok := obj.(string)
		if !ok {
			c.Workqueue.Forget(obj)
			klog.ErrorS(nil, "QUEUE: Invalid item type", "worker", workerID, "object", obj)
			return nil
		}

		klog.InfoS("QUEUE: Item dequeued", "worker", workerID, "resource", key)

		// Parse the key
		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			klog.ErrorS(err, "WORKER: Invalid resource key", "worker", workerID, "key", key)
			c.Workqueue.Forget(obj)
			return nil
		}

		// Get the object
		obj, exists, err := c.Informer.GetIndexer().GetByKey(key)
		if err != nil {
			klog.ErrorS(err, "WORKER: Error retrieving object from cache", "worker", workerID, "key", key)
			return err
		}

		// Process based on whether the object exists
		if !exists {
			klog.InfoS("WORKER: Processing deleted resource",
				"worker", workerID,
				"resource", key,
				"namespace", namespace,
				"name", name)
		} else {
			if u, ok := obj.(*unstructured.Unstructured); ok {
				klog.InfoS("WORKER: Processing resource",
					"worker", workerID,
					"resource", key,
					"namespace", namespace,
					"name", name,
					"kind", u.GetKind())
			} else {
				klog.InfoS("WORKER: Processing resource",
					"worker", workerID,
					"resource", key)
			}
		}

		// Simulate work
		time.Sleep(100 * time.Millisecond)

		// Done with this item
		c.Workqueue.Forget(obj)
		klog.InfoS("QUEUE: Item processed successfully",
			"worker", workerID,
			"resource", key)
		return nil
	}(obj)

	if err != nil {
		klog.ErrorS(err, "WORKER: Error processing item",
			"worker", workerID)
		c.Workqueue.AddRateLimited(obj)
		return true
	}

	return true
}

