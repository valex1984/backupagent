package utils

import (
	"fmt"
	"net/http"
	"net/url"

	lxd "github.com/lxc/lxd/client"
)

func PrepareLxcBackupRequest(containerName string, name string, lxc *lxd.InstanceServer) (*http.Client, *http.Request, error) {

	i := *lxc
	if !i.HasExtension("container_backup") {
		return nil, nil, fmt.Errorf("the server is missing the required \"container_backup\" API extension")
	}
	client, err := i.GetHTTPClient()
	if err != nil {
		return nil, nil, err
	}
	uri := fmt.Sprintf("%s/1.0/containers/%s/backups/%s/export", "http://unix.socket",
		url.PathEscape(containerName), url.PathEscape(name))

	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, nil, err
	}

	return client, request, nil
}
