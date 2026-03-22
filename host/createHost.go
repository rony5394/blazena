package host

import (
	cfg "github.com/rony5394/blazena/config"
	"github.com/docker/docker/client"
	"encoding/json"
	"encoding/base64"
	"context"
	"io"
	"os"
	"archive/tar"
	"bytes"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
);

func createStorageContainer(Config cfg.Config, DockerClient *client.Client, sshSkPem string, sshHostPkPem string){
	authConfig := registry.AuthConfig{
		Username: Config.RegistryAuth.Username, 
		Password: Config.RegistryAuth.Password,
	}

	authJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic("Failed to marshal auth config!"+ err.Error());
	}

	authString := base64.URLEncoding.EncodeToString(authJSON);

	ipc, err := DockerClient.ImagePull(context.Background(), Config.BlazenaImageUrl, image.PullOptions{RegistryAuth: authString});
	if err != nil {
		panic("Failed to pull blazena image!"+ err.Error());
	}
	defer ipc.Close();

	io.Copy(io.Discard, ipc);

	cr, err := DockerClient.ContainerCreate(context.Background(), &container.Config{
		Image: Config.BlazenaImageUrl, 
		Labels: map[string]string{
			"blazena.storage": "true",
		},
		Cmd: strslice.StrSlice{"sleep", "3h"},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
				mount.Mount{
					Type: mount.TypeBind,
					Source: Config.LocalBasePath,
					Target: "/volume",
					ReadOnly: true,
				},
			},
		//AutoRemove: true,
		NetworkMode: container.NetworkMode(Config.Constants.OverlayNetworkName),
	}, &network.NetworkingConfig{
		}, &v1.Platform{}, Config.Constants.StorageContainerName);

	if err != nil {
		panic("Failed to create BlazenaStorage container!"+err.Error());
	}

	err = DockerClient.ContainerStart(context.Background(), cr.ID, container.StartOptions{}); 
	
	if err != nil{
		panic("Failed to start BlazenaStorage container!"+err.Error());
	}


	var buf bytes.Buffer;
	tw := tar.NewWriter(&buf);

	addToTar(tw, "ssh-key", sshSkPem);
	addToTar(tw, "expected-host-key", "[tasks."+ Config.Constants.HelperServiceName +"]:2222 "+ sshHostPkPem);
	tw.Close();
	if err != nil {panic("The british are comming!")}

	os.WriteFile("/tmp/test", buf.Bytes(), os.ModeAppend);
	
	err = DockerClient.CopyToContainer(context.Background(), Config.Constants.StorageContainerName, "/", &buf, container.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	});
	if err != nil {
		panic("Failed to copy ssh key to container!"+err.Error());
	}


}


