# Tests

`StreamWithContext` is the k8s lib function to exec a command into a pod.

When provided a non reachable pod, the function errors but the stdin pipe is never fully read,
leading to a hanging process

## Run

```shell
go run `ls *.go | grep -v _test.go`
```

There is also a test function you can run in vscode (or whatever you use) with a debugger.
Set a checkpoint on line 66 (`stdinPipe.Write`)

## Solution

To avoid waiting indefinitely, you need to add a `done` channel that is closed when the
StreamWithContext function exits, and wait for either the write to finish, or the `done` channel to
be closed.

Don't forget to handle the channel closure on error.
