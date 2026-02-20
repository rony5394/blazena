package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"net/http"

	"github.com/moby/moby/client"
)

// Add mutex.
var ApiClient *client.Client;
var scale sync.Map;
var token string = "12345";

type aService struct{
	ServiceId string `json:"serviceId"`;
	VolumeNames []string `json:"volumeNames"`;
	Node string `json:"node"`;
}

func Run(){
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM);

	var err error;
	ApiClient, err = client.New(client.FromEnv);
	if(err != nil){
		panic("Docker client was not able to init from env!" + err.Error());
	}

	info, err := ApiClient.Info(context.Background(), client.InfoOptions{});
	if(err != nil){
		panic("Error getting info!" + err.Error());
	}

	if(!info.Info.Swarm.ControlAvailable){
		panic("Node is not a swarm manager.");
	}

	server := &http.Server{
		Addr: ":1234",
	}

	http.HandleFunc("/services", listServices);
	http.HandleFunc("/prepare", prepare);
	http.HandleFunc("/cleanup", cleanup);
	go func(){
		err = server.ListenAndServe(); 
		if err == http.ErrServerClosed {
			return;

		}
		if(err != nil){
			panic("Unable to start http server!" + err.Error());
		}

	}();

	time.Sleep(10*time.Millisecond);
	<-ctx.Done();
	fmt.Println("Stopping http server.");
	server.Close();
	fmt.Println("Exiting!");
}

func bearerAuth(w http.ResponseWriter, r *http.Request)bool {
	authHeader := r.Header.Get("Authorization")
	expected := "Bearer " + token

	if authHeader != expected {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Unauthorized")
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

	list, err := ApiClient.ServiceList(context.Background(), client.ServiceListOptions{});
	if(err != nil){
		panic("Unable to list services!" + err.Error());
	}

	var services []aService;

	for _, service:= range list.Items{
		var settings map[string]string = service.Spec.Labels; 
		

		if(settings["blazena.enable"] != "true"){
			continue;

		}

		targetVolumes := strings.Split(settings["blazena.volumes"], ",");

		services = append(services, aService{
			ServiceId: service.ID,
			VolumeNames: targetVolumes,
			Node: settings["blazena.node"],
		});

	}

	bytes, err := json.Marshal(services);

	if err != nil{
		panic("Error during response encoding!" + err.Error());
	}

	fmt.Fprint(w, string(bytes));
}
