package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/VedRatan/remediation-server/handlers"
	"github.com/VedRatan/remediation-server/k8s"
	"github.com/VedRatan/remediation-server/k8scontroller"
	"github.com/VedRatan/remediation-server/types"
	"github.com/gorilla/mux"
	k8sgptv1alpha1 "github.com/k8sgpt-ai/k8sgpt-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme    = runtime.NewScheme()
	k8sClient client.Client
)

func startServer(router *mux.Router) {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "7070" // default port
	}
	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // Set max header size (e.g., 1 MB)
	}

	log.Printf("Remediation server is starting at :%s", port)
	// Start the server
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func main() {
	var runAs string
	flag.StringVar(&runAs, "runAs", "k8s-controller", "run as a `server` or `k8s-controller`")
	flag.StringVar(&types.K8sAgentServiceURL, "k8s-agent-url", "", "The LoadBalancer IP or DNS of the k8s-agent-service (required)")
	flag.StringVar(&types.AiAgent, "ai", "gemini", "AI agent to use as a backend to provide remediations")
	flag.StringVar(&types.AiAgentKey, "api-key", "", "AI agent api key")
	flag.Parse()
	types.AiAgent = strings.ToLower(types.AiAgent) // make sure that the case is uniform

	// Validate the flag
	if types.K8sAgentServiceURL == "" {
		fmt.Println("Error: The --k8s-agent-url flag is required")
		flag.Usage()
		os.Exit(1) // Exit with a non-zero status code
	}
	if types.AiAgentKey == "" {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			fmt.Println("GEMINI_API_KEY or --api-key must be set")
			os.Exit(1) // Exit with a non-zero status code
		}
		types.AiAgentKey = apiKey
	}

	switch runAs {
	case "server":
		r := mux.NewRouter()
		// Define the HTTP server
		r.HandleFunc("/webhook", handlers.WebhookHandler)

		startServer(r)
	case "k8s-controller":
		utilruntime.Must(k8sgptv1alpha1.AddToScheme(scheme))
		utilruntime.Must(corev1.AddToScheme(scheme))
		k8sClient = k8s.NewOrDie(scheme)
		// Create a new controller
		c := k8scontroller.NewController(k8sClient)

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		var wg wait.Group

		wg.StartWithContext(ctx, func(ctx context.Context) {
			log.Println("informer starting... ", "core.k8sgpt.ai/v1alpha1/results")
			c.Informer.Run(ctx.Done())
		})
		if !cache.WaitForCacheSync(ctx.Done(), c.Informer.HasSynced) {
			cancel()
			fmt.Println("failed to wait for cache sync:", "core.k8sgpt.ai/v1alpha1/results")
			return
		}

		// Start the controller
		c.Start(ctx)

		select { //nolint:gosimple
		case <-ctx.Done():
			c.Stop()
			return
		}
	}
}
