package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zaccone/spf"
)

func main() {
	r := mux.NewRouter()
	r.Handle("/bootstrap.min.css", http.FileServer(http.Dir("./static/")))
	r.HandleFunc("/{iaas}", handleRequest)
	r.HandleFunc("/{iaas}/{version}", handleRequest)
	r.Handle("/", http.FileServer(http.Dir("./static/")))

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

	if iaasString == "auto" {
		xff := r.Header.Get("X-Forwarded-For")
		splitXff := strings.Split(xff, ", ")
		source, err := autodetectSource(net.ParseIP(splitXff[0]))
		if err != nil || source == "" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("could not autodetect IaaS"))
			return
		}
		iaasString = source
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

func autodetectSource(ipAddress net.IP) (string, error) {
	gcp, err := isGCPAddress(ipAddress)
	if err != nil {
		return "", err
	}
	if gcp {
		return "gcp", nil
	}

	aws, err := isAWSAddress(ipAddress)
	if err != nil {
		return "", err
	}
	if aws {
		return "aws", nil
	}

	azure, err := isAzureAddress(ipAddress)
	if err != nil {
		return "", err
	}
	if azure {
		return "azure", nil
	}

	return "", nil
}

func isGCPAddress(ipAddress net.IP) (bool, error) {
	r, _, err := spf.CheckHost(ipAddress, "_cloud-netblocks.googleusercontent.com", "")
	if err != nil {
		return false, err
	}

	return r == spf.Pass, nil
}

func isAWSAddress(ipAddress net.IP) (bool, error) {
	names, err := net.LookupAddr(ipAddress.String())
	if err != nil {
		// will return an error if the address is not found
		return false, nil
	}

	return strings.Contains(names[0], "amazonaws.com."), nil
}

func isAzureAddress(ipAddress net.IP) (bool, error) {
	r, err := http.Get(fmt.Sprintf("http://www.azurespeed.com/api/region?ipOrUrl=%s", url.QueryEscape(ipAddress.String())))
	if err != nil {
		return false, err
	}

	jsonoutput := map[string]interface{}{}
	json.NewDecoder(r.Body).Decode(&jsonoutput)
	if jsonoutput["cloud"] == nil {
		return false, nil
	}
	return jsonoutput["cloud"].(string) == "Azure", nil
}
