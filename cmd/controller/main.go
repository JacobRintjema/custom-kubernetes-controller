package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"

	"github.com/jrintjema/custom-kubernetes-controller/pkg/config"
	"github.com/jrintjema/custom-kubernetes-controller/pkg/controller"
	"github.com/jrintjema/custom-kubernetes-controller/pkg/handlers"
)

func main() {
	// Initialize logging
	klog.InitFlags(nil)
	flag.Set("skip_headers", "true")
	flag.Set("logtostderr", "true")
	flag.Parse()
	defer klog.Flush()

	// Get Kubernetes configuration
	kubeConfig, err := config.GetKubeConfig()
	if err != nil {
		klog.ErrorS(err, "Failed to get Kubernetes config")
		os.Exit(1)
	}

	// Create controller
	ctrl, err := controller.NewController(kubeConfig)
	if err != nil {
		klog.ErrorS(err, "Failed to create controller")
		os.Exit(1)
	}

	// Set up event handlers
	eventHandler := handlers.NewHandler(ctrl.Workqueue)
	eventHandler.SetupHandlers(ctrl.Informer)

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		sig := <-sigCh
		klog.InfoS("Received shutdown signal", "signal", sig)
		close(ctrl.StopCh)
	}()

	// Start controller
	klog.Info("Starting controller")
	ctrl.Run(2) // Start with 2 workers
}
