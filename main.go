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
	"time"

	"github.com/gorilla/mux"
	"github.com/zaccone/spf"
)

func main() {
	r := mux.NewRouter()
	r.Handle("/bootstrap.min.css", http.FileServer(http.Dir("./static/")))
	r.HandleFunc("/{iaas}", handleRequest)
	r.HandleFunc("/{iaas}/{versionOrLine}", handleRequest)
	r.HandleFunc("/{iaas}/{versionOrLine}/{version}", handleRequest)
	r.Handle("/", http.FileServer(http.Dir("./static/")))

	err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), r)
	if err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var iaas, version string

	var line = "ubuntu-xenial"

	vars := mux.Vars(r)
	iaasString, ok := vars["iaas"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	versionOrLineString, ok := vars["versionOrLine"]
	if ok {
		if ok, lineVar := isLineVariable(vars["versionOrLine"]); ok {
			line = lineVar
		} else {
			if versionOrLineString != "latest" {
				version = fmt.Sprintf("?v=%s", versionOrLineString)
			}
		}
	}

	versionString, ok := vars["version"]
	if ok {
		if versionString != "latest" {
			version = fmt.Sprintf("?v=%s", versionString)
		}
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

	http.Redirect(w, r, fmt.Sprintf("https://bosh.io/d/stemcells/bosh-%s-%s-go_agent%s", iaas, line, version), 301)
}

func isLineVariable(line string) (bool, string) {
	switch line {
	case "trusty", "ubuntu-trusty", "ubuntutrusty", "t":
		return true, "ubuntu-trusty"
	case "xenial", "ubuntu-xenial", "ubuntuxenial", "ubuntu", "x":
		return true, "ubuntu-xenial"
	case "windows", "windows2016", "windows16":
		return true, "windows2016"
	case "windows2012", "windows12":
		return true, "windows2012R2"
	case "centos", "centos7", "centos-7":
		return true, "centos-7"
	default:
		return false, ""
	}
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
	timeout := time.Duration(time.Second)
	httpClient := http.Client{
		Timeout: timeout,
	}

	r, err := httpClient.Get(fmt.Sprintf("http://www.azurespeed.com/api/region?ipOrUrl=%s", url.QueryEscape(ipAddress.String())))
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
