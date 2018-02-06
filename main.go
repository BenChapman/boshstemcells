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
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://bosh.io/", 301)
	})
	r.HandleFunc("/{iaas}", func(w http.ResponseWriter, r *http.Request) {
		var iaas string

		vars := mux.Vars(r)
		iaasString, ok := vars["iaas"]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch iaasString {
		case "gcp":
			iaas = "google-kvm"
		case "aws":
			iaas = "aws-xen-hvm"
		case "azure":
			iaas = "azure-hyperv"
		case "vsphere":
			iaas = "vsphere-esxi"
		case "openstack":
			iaas = "openstack-kvm"
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("https://bosh.io/d/stemcells/bosh-%s-ubuntu-trusty-go_agent", iaas), 301)
	})

	err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), r)
	if err != nil {
		log.Fatal(err)
	}
}
