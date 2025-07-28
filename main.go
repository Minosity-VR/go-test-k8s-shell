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

func writeToPipe(pipeWriter *io.PipeWriter, data []byte, result chan error) {
	_, err := pipeWriter.Write(data)
	result <- err
}

func testSimpleShell(ctx context.Context, clientset *kubernetes.Clientset, kubeConfig *rest.Config) error {
	stdinPipeReader, stdinPipeWriter := io.Pipe()

	stdOut := bytes.Buffer{}
	stdErr := bytes.Buffer{}

	done := make(chan struct{})

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
			Stdin:  stdinPipeReader,
			Stdout: &stdOut,
			Stderr: &stdErr,
			Tty:    false,
		})
		if err != nil {
			fmt.Printf("shell session error: %v\n", err)
		}
		close(done)
	}()

	time.Sleep(1 * time.Second) // Wait for the shell to start

	// This version will hang on the write
	stdinPipeWriter.Write([]byte("echo 'Hello, World!'\n"))

	// This version will not hang on the write. It requires
	// - a done channel in the goroutine running the StreamWithContext
	// - a result channel to receive the error from the write
	// - the write to be started in a new goroutine
	// - a select case to either wait for the write, or catch the done signal
	result := make(chan error)
	go writeToPipe(stdinPipeWriter, []byte("echo 'Hello, World!'\n"), result)

	select {
	case err := <-result:
		if err != nil {
			fmt.Printf("error writing to pipe: %v\n", err)
		}
	case <-done:
		fmt.Println("done")
	}

	time.Sleep(1 * time.Second) // Wait for command to complete

	cmdOut := stdOut.String()
	cmdErr := stdErr.String()

	fmt.Printf("cmdOut: %s\n", cmdOut)
	fmt.Printf("cmdErr: %s\n", cmdErr)

	stdinPipeReader.Close()

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
