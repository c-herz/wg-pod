/*
Copyright © 2021 b-m-f<max@ehlers.berlin>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

 3. Neither the name of the copyright holder nor the names of its contributors
    may be used to endorse or promote products derived from this software
    without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package podman

import (
	"fmt"
	"github.com/b-m-f/wg-pod/pkg/shell"
	"path"
	"strconv"
	"strings"
)

func GetNamespace(name string, url string) (string, error) {
	var cmdArgs []string
	if url != "" {
		formattedUrl := fmt.Sprintf("unix://%s", url)
		cmdArgs = append(cmdArgs, "--url", formattedUrl)

	}
	cmdArgs = append(cmdArgs, "inspect", "--format", "{{.NetworkSettings.SandboxKey}}", name)
	namespaceFileOnSystem, err := shell.ExecuteCommand("podman", cmdArgs)
	if err != nil {
		return "", err

	}
	namespace := strings.TrimSpace(path.Base(namespaceFileOnSystem))
	if url != "" {
		// we're running rootfully, link the netns
		_, err = LinkRootlessNamespace(name, namespace, url)
		return namespace, err
	}
	return namespace, nil

}

func LinkRootlessNamespace(name string, namespace string, url string) (string, error) {
	bashCmd := fmt.Sprintf("podman --url unix://%s ps --format json --filter name=%s | jq -r '.[0].Pid'", url, name)
	pid, err := shell.ExecuteCommand("sh", []string{"-c", bashCmd})
	if err != nil {
		return "", err
	}
	pid = strings.TrimSpace(pid)
	// ensure the pid we got was a number
	if _, err := strconv.Atoi(pid); err != nil {
		return "", fmt.Errorf("Invalid PID: %s", pid)
	}
	_, err = shell.ExecuteCommand("ip", []string{"netns", "attach", namespace, pid})
	if err != nil {
		return "", fmt.Errorf("Failed to attach network namespace: %s", err)
	}
	return namespace, nil

}
