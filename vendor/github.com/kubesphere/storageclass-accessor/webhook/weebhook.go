package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"net/http"
)

var (
	certFile string
	keyFile  string
	port     int
)

// CmdWebhook is user by Cobra
var CmdWebhook = &cobra.Command{
	Use:  "storageClass-webhook",
	Args: cobra.MaximumNArgs(0),
	Run:  main,
}

func init() {
	CmdWebhook.Flags().StringVar(&certFile, "tls-cert-file", "",
		"File containing the x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert). Required.")
	CmdWebhook.Flags().StringVar(&keyFile, "tls-private-key-file", "",
		"File containing the x509 private key matching --tls-cert-file. Required.")
	CmdWebhook.Flags().IntVar(&port, "port", 443,
		"Secure port that the webhook listens on")
	CmdWebhook.MarkFlagRequired("tls-cert-file")
	CmdWebhook.MarkFlagRequired("tls-private-key-file")
}

// admitV1beta1Func handles a v1 admission
type admitV1Func func(v1.AdmissionReview) *v1.AdmissionResponse

// admitHandler is a handler, for both validators and mutators, that supports multiple admission review versions
type admitHandler struct {
	v1 admitV1Func
}

func newDelegateToV1AdmitHandler(f admitV1Func) admitHandler {
	return admitHandler{v1: f}
}

func server(w http.ResponseWriter, r *http.Request, admit admitHandler) {
	var body []byte
	if r.Body == nil {
		msg := "Expected request body to be non-empty"
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Request could not be decoded: %v", err)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
	}
	body = data

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		msg := fmt.Sprintf("contentType=%s, expect application/json", contentType)
		klog.Errorf(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	klog.V(2).Info(fmt.Sprintf("handling request: %s", body))

	deserializer := codecs.UniversalDeserializer()
	obj, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Request could not be decoded: %v", err)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	var responseObj runtime.Object
	switch *gvk {
	// TODO v1beta1 admissionReview
	case v1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1.AdmissionReview)
		if !ok {
			msg := fmt.Sprintf("Expected v1.AdmissionReview but got: %T", obj)
			klog.Errorf(msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		responseAdmissionReview := &v1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	default:
		msg := fmt.Sprintf("Unsupported group version kind: %v", gvk)
		klog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	klog.V(2).Info(fmt.Sprintf("sending response: %v", responseObj))

	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		klog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

func serverPVCRequest(w http.ResponseWriter, r *http.Request) {
	server(w, r, newDelegateToV1AdmitHandler(AdmitPVC))
}

func startServer(ctx context.Context, tlsConfig *tls.Config, cw *CertWatcher) error {
	go func() {
		klog.Info("Starting certificate watcher")
		if err := cw.Start(ctx); err != nil {
			klog.Errorf("certificate watcher error: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/persistentvolumeclaims", serverPVCRequest)
	srv := &http.Server{
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	// listener is always closed by srv.Serve
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		return err
	}
	return srv.Serve(listener)
}
func main(cmd *cobra.Command, args []string) {
	// Create new cert watcher
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()
	cw, err := NewCertWatcher(certFile, keyFile)
	if err != nil {
		klog.Fatalf("failed to initialize new cert watcher: %v", err)
	}
	tslConfig := &tls.Config{
		GetCertificate: cw.GetCertificate,
	}

	if err := startServer(ctx, tslConfig, cw); err != nil {
		klog.Fatalf("server stopped: %v", err)
	}
}
