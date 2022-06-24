package main

import (
	"fmt"
	"os"
	"os/user"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	var config *rest.Config

	if k8s_port := os.Getenv("KUBERNETES_PORT"); k8s_port == "" {
		// fmt.Println("Using local kubeconfig") // Disable all non-compatible printout
		var kubeconfig string
		home := usr.HomeDir
		if home != "" {
			kubeconfig = fmt.Sprintf("%s/.kube/config", home)
		} else {
			panic("home directory unknown")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// fmt.Println("Using in cluster authentication") // Disable all non-compatible printout
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watchlist := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"events",
		v1.NamespaceAll,
		fields.Everything(),
	)
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Event{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				event, ok := obj.(*v1.Event)
				if ok {
					fmt.Printf("[k8s-event-logger] Namespace: %s, Kind: %s, Name: %s, Type: %s, Reason: %s, Message: %s\n", event.InvolvedObject.Namespace, event.InvolvedObject.Kind, event.InvolvedObject.Name, event.Type, event.Reason, event.Message)
				}
			},
		},
	)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}

}

// For tricking the annoying "declared but not used", use: Use(var1, var2, var3 ...)
func Use(vals ...interface{}) {
	for _, val := range vals {
		_ = val
	}
}
