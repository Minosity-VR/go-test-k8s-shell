package main

import (
	"testing"
	"time"

	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

// If you want to us a debugger, here is a test function
func TestSimpleShell(t *testing.T) {
	ctx := t.Context()
	k8sCfg, err := ctrl.GetConfig()
	if err != nil {
		t.Fatalf("failed to create k8s config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(k8sCfg)
	if err != nil {
		t.Fatalf("failed to create k8s clientset: %v", err)
	}

	err = testSimpleShell(ctx, clientset, k8sCfg)
	if err != nil {
		t.Fatalf("failed to test simple shell: %v", err)
	}

	time.Sleep(1 * time.Second)
}
