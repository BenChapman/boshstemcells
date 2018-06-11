package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	r.HandleFunc("/{iaas}", handleRequest)
	r.HandleFunc("/{iaas}/{version}", handleRequest)

	err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), r)
	if err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var iaas, version string

	vars := mux.Vars(r)
	iaasString, ok := vars["iaas"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	versionString, ok := vars["version"]
	if ok {
		version = fmt.Sprintf("?v=%s", versionString)
	}

	switch iaasString {
	case "aws", "amazon":
		iaas = "aws-xen-hvm"
	case "azure":
		iaas = "azure-hyperv"
	case "gcp", "google":
		iaas = "google-kvm"
	case "openstack":
		iaas = "openstack-kvm"
	case "softlayer":
		iaas = "softlayer-xen"
	case "vsphere":
		iaas = "vsphere-esxi"
	case "vcloud":
		iaas = "vcloud-esxi"
	case "lite", "boshlite":
		iaas = "warden-boshlite"
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("https://bosh.io/d/stemcells/bosh-%s-ubuntu-trusty-go_agent%s", iaas, version), 301)
}
