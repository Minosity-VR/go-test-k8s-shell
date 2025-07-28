package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	// Non existing resources to trigger a exec error
	NAMESPACE      = "NON_EXISTING"
	POD_NAME       = "NON_EXISTING"
	CONTAINER_NAME = "NON_EXISTING"

	// Put here a pod you know you can access
	// POD_NAME       =
	// NAMESPACE      =
	// CONTAINER_NAME =
)

func testSimpleShell(ctx context.Context, clientset *kubernetes.Clientset, kubeConfig *rest.Config) error {
	pipeReader, stdinPipe := io.Pipe()

	stdOut := bytes.Buffer{}
	stdErr := bytes.Buffer{}

	// Create the exec request
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(POD_NAME).
		Namespace(NAMESPACE).
		SubResource("exec").
		Param("container", CONTAINER_NAME).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "false").
		Param("command", "sh")
	exec, err := remotecommand.NewWebSocketExecutor(kubeConfig, "GET", req.URL().String())
	if err != nil {
		panic(err)
	}

	// Run the shell
	go func() {
		err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:  pipeReader,
			Stdout: &stdOut,
			Stderr: &stdErr,
			Tty:    false,
		})
		if err != nil {
			fmt.Printf("shell session error: %v\n", err)
		}
	}()

	time.Sleep(1 * time.Second) // Wait for the shell to start

	stdinPipe.Write([]byte("echo 'Hello, World!'\n"))

	time.Sleep(1 * time.Second) // Wait for command to complete

	cmdOut := stdOut.String()
	cmdErr := stdErr.String()

	fmt.Printf("cmdOut: %s\n", cmdOut)
	fmt.Printf("cmdErr: %s\n", cmdErr)

	pipeReader.Close()

	return nil
}

func main() {
	ctx := context.Background()
	k8sCfg, err := ctrl.GetConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(k8sCfg)
	if err != nil {
		panic(err)
	}

	// testShellDetonator(ctx, clientset, k8sCfg)
	testSimpleShell(ctx, clientset, k8sCfg)
}
