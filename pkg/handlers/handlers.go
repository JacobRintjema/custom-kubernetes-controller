package handlers

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// Handler contains references needed for event handling
type Handler struct {
	Queue workqueue.RateLimitingInterface
}

// NewHandler creates a new event handler
func NewHandler(queue workqueue.RateLimitingInterface) *Handler {
	return &Handler{
		Queue: queue,
	}
}

// SetupHandlers registers event handlers with the informer
func (h *Handler) SetupHandlers(informer cache.SharedIndexInformer) {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    h.handleAdd,
		UpdateFunc: h.handleUpdate,
		DeleteFunc: h.handleDelete,
	})
}

// handleAdd is called when a resource is created
func (h *Handler) handleAdd(obj interface{}) {
	if u, ok := obj.(*unstructured.Unstructured); ok {
		key, err := cache.MetaNamespaceKeyFunc(obj)
		if err != nil {
			klog.ErrorS(err, "Failed to get key from object", "object", obj)
			return
		}
		
		klog.InfoS("EVENT: Resource added → Queueing",
			"resource", key,
			"name", u.GetName(),
			"namespace", u.GetNamespace(),
			"kind", u.GetKind())
		
		h.Queue.Add(key)
	} else {
		klog.InfoS("EVENT: Resource added but could not convert to Unstructured", 
			"object", obj)
	}
}

// handleUpdate is called when a resource is updated
func (h *Handler) handleUpdate(oldObj, newObj interface{}) {
	if u, ok := newObj.(*unstructured.Unstructured); ok {
		key, err := cache.MetaNamespaceKeyFunc(newObj)
		if err != nil {
			klog.ErrorS(err, "Failed to get key from object", "object", newObj)
			return
		}
		
		klog.InfoS("EVENT: Resource updated → Queueing",
			"resource", key,
			"name", u.GetName(),
			"namespace", u.GetNamespace(),
			"kind", u.GetKind())
		
		h.Queue.Add(key)
	} else {
		klog.InfoS("EVENT: Resource updated but could not convert to Unstructured", 
			"object", newObj)
	}
}

// handleDelete is called when a resource is deleted
func (h *Handler) handleDelete(obj interface{}) {
	var key string
	var err error
	var resourceName, resourceNamespace, resourceKind string

	if u, ok := obj.(*unstructured.Unstructured); ok {
		key, err = cache.MetaNamespaceKeyFunc(obj)
		resourceName = u.GetName()
		resourceNamespace = u.GetNamespace()
		resourceKind = u.GetKind()
	} else {
		// Handle tombstone objects
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.ErrorS(nil, "Error decoding tombstone", "object", obj)
			return
		}

		// Get resource info from tombstone
		u, ok := tombstone.Obj.(*unstructured.Unstructured)
		if !ok {
			klog.ErrorS(nil, "Tombstone contained non-Unstructured object", "object", tombstone.Obj)
			return
		}

		key, err = cache.MetaNamespaceKeyFunc(u)
		resourceName = u.GetName()
		resourceNamespace = u.GetNamespace()
		resourceKind = u.GetKind()
	}

	if err != nil {
		klog.ErrorS(err, "Failed to get key from object", "object", obj)
		return
	}

	klog.InfoS("EVENT: Resource deleted → Queueing",
		"resource", key,
		"name", resourceName,
		"namespace", resourceNamespace,
		"kind", resourceKind)
	
	h.Queue.Add(key)
}

