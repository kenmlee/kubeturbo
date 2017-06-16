package app

import (
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"

	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/kubernetes/scheme"

	//_ "k8s.io/kubernetes/plugin/pkg/scheduler/algorithmprovider"

	//"k8s.io/kubernetes/pkg/apis/componentconfig"
	//"k8s.io/kubernetes/pkg/client/leaderelection"
	//"k8s.io/kubernetes/pkg/client/leaderelection/resourcelock"
	"k8s.io/kubernetes/pkg/util/configz"
	"k8s.io/apiserver/pkg/server/healthz"

	kubeturbo "github.com/turbonomic/kubeturbo/pkg"
	"github.com/turbonomic/kubeturbo/pkg/discovery/probe"
	"github.com/turbonomic/kubeturbo/pkg/discovery/probe/stitching"
	"github.com/turbonomic/kubeturbo/pkg/turbostore"
	"github.com/turbonomic/kubeturbo/test/flag"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
)

const (
	// The default port for vmt service server
	VMTPort = 10265
)

// VMTServer has all the context and params needed to run a Scheduler
type VMTServer struct {
	Port            int
	Address         string
	Master          string
	K8sTAPSpec      string
	TestingFlagPath string
	KubeConfig      string
	BindPodsQPS     float32
	BindPodsBurst   int
	CAdvisorPort    int

	//LeaderElection componentconfig.LeaderElectionConfiguration

	EnableProfiling bool

	// If the underlying infrastructure is VMWare, we cannot reply on IP address for stitching. Instead we use the
	// systemUUID of each node, which is equal to UUID of corresponding VM discovered by VM probe.
	// The default value is false.
	UseVMWare bool
}

// NewVMTServer creates a new VMTServer with default parameters
func NewVMTServer() *VMTServer {
	s := VMTServer{
		Port:    VMTPort,
		Address: "127.0.0.1",
	}
	return &s
}

// AddFlags adds flags for a specific VMTServer to the specified FlagSet
func (s *VMTServer) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&s.Port, "port", s.Port, "The port that the kubeturbo's http service runs on")
	fs.IntVar(&s.CAdvisorPort, "cadvisor-port", 4194, "The port of the cadvisor service runs on")
	fs.StringVar(&s.Master, "master", s.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	fs.StringVar(&s.K8sTAPSpec, "turboconfig", s.K8sTAPSpec, "Path to the config file.")
	fs.StringVar(&s.TestingFlagPath, "testingflag", s.TestingFlagPath, "Path to the testing flag.")
	fs.StringVar(&s.KubeConfig, "kubeconfig", s.KubeConfig, "Path to kubeconfig file with authorization and master location information.")
	fs.BoolVar(&s.EnableProfiling, "profiling", false, "Enable profiling via web interface host:port/debug/pprof/.")
	fs.BoolVar(&s.UseVMWare, "usevmware", false, "If the underlying infrastructure is VMWare.")
	//leaderelection.BindFlags(&s.LeaderElection, fs)
}

func createRecorder(kubecli *kubernetes.Clientset) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{
								Interface: v1core.New(kubecli.Core().RESTClient()).Events("")})
	return eventBroadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{Component: "kubeturbo"})
}

func (s *VMTServer) createKubeClient() (*kubernetes.Clientset, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags(s.Master, s.KubeConfig)
	if err != nil {
		glog.Errorf("Error getting kubeconfig:  %s", err)
		return nil, err
	}
	// This specifies the number and the max number of query per second to the api server.
	kubeConfig.QPS = 20.0
	kubeConfig.Burst = 30

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		glog.Fatalf("Invalid API configuration: %v", err)
	}

	return kubeClient, nil
}

// dependency problem to enable leaderElection.
//func (s *VMTServer) RunWithLeaderElection(kubeClient *kubernetes.Clientset) error {
//
//    id, err := os.Hostname()
//    if err != nil {
//		return err
//	}
//
//	rl, err := resourcelock.New(s.LeaderElection.ResourceLock,
//		"kube-system",
//		"kubeturbo",
//		kubeClient,
//		resourcelock.ResourceLockConfig{
//			Identity:      id,
//			EventRecorder: vmtConfig.Recorder,
//		})
//	if err != nil {
//		glog.Fatalf("error creating lock: %v", err)
//	}
//
//	leaderelection.RunOrDie(leaderelection.LeaderElectionConfig{
//		Lock:          rl,
//		LeaseDuration: s.LeaderElection.LeaseDuration.Duration,
//		RenewDeadline: s.LeaderElection.RenewDeadline.Duration,
//		RetryPeriod:   s.LeaderElection.RetryPeriod.Duration,
//		Callbacks: leaderelection.LeaderCallbacks{
//			OnStartedLeading: run,
//			OnStoppedLeading: func() {
//				glog.Fatalf("lost master")
//			},
//		},
//	})
//}

// Run runs the specified VMTServer.  This should never exit.
func (s *VMTServer) Run(_ []string) error {
	if s.KubeConfig == "" && s.Master == "" {
		glog.Warningf("Neither --kubeconfig nor --master was specified.  Using default API client.  This might not work.")
	}

	glog.V(3).Infof("Master is %s", s.Master)

	if s.TestingFlagPath != "" {
		flag.SetPath(s.TestingFlagPath)
	}

	if s.CAdvisorPort == 0 {
		s.CAdvisorPort = 4194
	}

	// The default property type for stitching is IP.
	pType := stitching.IP
	if s.UseVMWare {
		// If the underlying hypervisor is vCenter, use UUID.
		// Refer to Bug: https://vmturbo.atlassian.net/browse/OM-18139
		pType = stitching.UUID
	}
	probeConfig := &probe.ProbeConfig{
		CadvisorPort:          s.CAdvisorPort,
		StitchingPropertyType: pType,
	}

	go startHttp(s)

	glog.V(3).Infof("spec path is: %v", s.K8sTAPSpec)
	k8sTAPSpec, err := kubeturbo.ParseK8sTAPServiceSpec(s.K8sTAPSpec)
	if err != nil {
		glog.Errorf("Failed to generate correct TAP config: %s", err)
		os.Exit(1)
	}

	kubeClient, err := s.createKubeClient()
	if err != nil {
		glog.Errorf("Failed to get kubeClient: %v", err)
		os.Exit(1)
	}

	broker := turbostore.NewPodBroker()
	vmtConfig := kubeturbo.NewVMTConfig(kubeClient, probeConfig, broker, k8sTAPSpec)
	glog.V(3).Infof("Finished creating turbo configuration: %++v", vmtConfig)

	vmtConfig.Recorder = createRecorder(kubeClient)

	vmtService := kubeturbo.NewKubeturboService(vmtConfig)

	run := func(_ <-chan struct{}) {
		vmtService.Run()
		select {}
	}

	//if !s.LeaderElection.LeaderElect {
		glog.Infof("No leader election")
		run(nil)
		glog.Fatal("this statement is unreachable")
		panic("unreachable")
	//}

	glog.Fatal("this statement is unreachable")
	panic("unreachable")
}

func startHttp(s *VMTServer) {
	mux := http.NewServeMux()

	//healthz
	healthz.InstallHandler(mux)

	//debug
	if s.EnableProfiling {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	//config
	if c, err := configz.New("componentconfig"); err == nil {
		c.Set(s)
	} else {
		glog.Errorf("unable to register configz: %v", err)
	}
	configz.InstallHandler(mux)

	//prometheus.metrics
	mux.Handle("/metrics", prometheus.Handler())

	server := &http.Server{
		Addr:    net.JoinHostPort(s.Address, strconv.Itoa(s.Port)),
		Handler: mux,
	}
	glog.Fatal(server.ListenAndServe())
}
