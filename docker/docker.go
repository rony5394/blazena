package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"net/http"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	cfg "github.com/rony5394/blazena/config"
)

// Add mutex.
var ApiClient *client.Client;
var scale sync.Map;
var token string = "12345";
var theConfig cfg.Config;

type aService struct{
	ServiceId string `json:"serviceId"`;
	VolumeNames []string `json:"volumeNames"`;
	Node string `json:"node"`;
	Dependents []string `json:"dependents"`;
}

func Run(Config cfg.Config){
	// Before touching the line below think.
	theConfig = Config;

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM);

	var err error;
	ApiClient, err = client.NewClientWithOpts(client.FromEnv);
	if(err != nil){
		slog.Error("Failed to create docker client!", slog.String("note", "Try to look into DOCKER_HOST env var or check if socket exists and works"));
		os.Exit(1);
	}

	info, err := ApiClient.Info(context.Background())
	if(err != nil){
		slog.Error("Error getting info from docker socket!", slog.String("note", "This is kind of ping."));
		os.Exit(1);
	}

	if(!info.Swarm.ControlAvailable){
		slog.Error("This node is not a swarm manager!");
		os.Exit(1);
	}

	server := &http.Server{
		Addr: ":1234",
	}

	http.HandleFunc("/services", listServices);
	http.HandleFunc("/scale/up", scaleUp);
	http.HandleFunc("/scale/down", scaleDown);
	http.HandleFunc("/prepare", prepare);
	http.HandleFunc("/cleanup", cleanup);
	http.HandleFunc("/keys", exchangeKeys);
	// I'll make it better someday.
	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if !bearerAuth(w, r){return}
		fmt.Fprint(w, "Shutdown!");
		time.Sleep(1*time.Second);
		stop();
	});

	ApiClient.NetworkCreate(context.Background(), Config.Constants.OverlayNetworkName, network.CreateOptions{
		Attachable: true,
		// Internal: true,
		Driver: "overlay",
		Labels: map[string]string{
			"blazena.pohar": "true",
		},
	});

	go func(){
		err = server.ListenAndServe(); 
		if err == http.ErrServerClosed {
			return;

		}
		if(err != nil){
			slog.Error("Unable to start http server!", slog.Any("propagatedError", err));
			os.Exit(1);
		}

	}();

	fmt.Println("Api has been started!");

	time.Sleep(10*time.Millisecond);
	<-ctx.Done();
	fmt.Println("Stopping http server.");
	server.Close();

	ApiClient.NetworkRemove(context.Background(), Config.Constants.OverlayNetworkName);
	ApiClient.ConfigRemove(context.Background(), "blazenaSSHPublicKey")
	ApiClient.SecretRemove(context.Background(), "blazenaSSHHostPrivateKey");

	fmt.Println("Exiting!");
}

func bearerAuth(w http.ResponseWriter, r *http.Request)bool {
	authHeader := r.Header.Get("Authorization")
	expected := "Bearer " + token

	if authHeader != expected {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Unauthorized")
		slog.Warn("Unauthorized request received", slog.Any("request", *r));
		return false;
	}
	return true;
}

func listServices(w http.ResponseWriter, r *http.Request){
	if(r.Method != http.MethodGet){
		w.WriteHeader(http.StatusMethodNotAllowed);
		fmt.Fprintln(w, "Method Not Allowed");
		return;
	}

	if !bearerAuth(w,r) {return};

	list, err := ApiClient.ServiceList(context.Background(), swarm.ServiceListOptions{}); 
	if(err != nil){
		panic("Unable to list services!" + err.Error());
	}

	var services []aService;


	nodes, err := ApiClient.NodeList(context.Background(), swarm.NodeListOptions{});

	var validNodeHostnames []string;

	for _, node := range nodes{
		validNodeHostnames = append(validNodeHostnames, node.Description.Hostname);
	}


	SERVICES:
	for _, service := range list{
		var settings map[string]string = service.Spec.Labels; 
		

		if(settings["blazena.enable"] != "true"){
			continue;
		}

		if !contains(validNodeHostnames, settings["blazena.node"]) {
			errMsg := "Node with hostname:'"+ settings["blazena.node"] +"' does not exist.";
			slog.Warn("Invalid Service Config!", slog.String("serviceId", service.ID), slog.String("errMessage", errMsg));
			continue SERVICES;
		}

		targetVolumes := strings.Split(settings["blazena.volumes"], ",");

		var validVolumeNames []string;
		for _, mnt := range service.Spec.TaskTemplate.ContainerSpec.Mounts{
			if mnt.Type != mount.TypeVolume {
				continue
			}
			validVolumeNames = append(validVolumeNames, mnt.Source);
		}

		for _, volume := range targetVolumes{
			if contains(validVolumeNames, volume){
				continue;
			}

			errMsg := "Volume name '"+ volume + "' is not in the service spec!";
			slog.Warn("Invalid Service Config!", slog.String("serviceId", service.ID), slog.String("errMessage", errMsg));
			continue SERVICES;
		}

		
		var dependents []string;

		if(settings["blazena.dependents"] != ""){
			dependents = strings.Split(settings["blazena.dependents"], ",");
		} else {
			dependents = make([]string, 0);
		}

		var validDependents []string;
		for _, x := range list{
			validDependents = append(validDependents, x.Spec.Name);
		}
		
		slog.Debug("validDependents", slog.Any("value", validDependents));
		slog.Debug("dependents", slog.Any("value", dependents));

		for _, dependent := range dependents {
			if contains(validDependents, dependent){continue;}

			errMsg := "Dependent named '"+ dependent +"' was not found in this cluster!";
			slog.Warn("Invalid Service Config!", slog.String("serviceId", service.ID), slog.String("errMessage", errMsg));
			continue SERVICES;
		}

		services = append(services, aService{
			ServiceId: service.ID,
			VolumeNames: targetVolumes,
			Node: settings["blazena.node"],
			Dependents: dependents,
		});

	}

	bytes, err := json.Marshal(services);

	if err != nil{
		slog.Error("Error during response encoding!", slog.Any("propagatedError", err));
		os.Exit(1);
	}

	fmt.Fprint(w, string(bytes));
}
