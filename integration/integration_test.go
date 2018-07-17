package integration_test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("BoshStemcells.com", func() {
	It("should return 200 for the homepage", func() {
		response, err := http.Get(fmt.Sprintf("http://localhost:%d", serverPort))
		Expect(err).ToNot(HaveOccurred())
		Expect(response.StatusCode).To(Equal(200))
	})

	It("should include a BoshStemcells title", func() {
		response, err := http.Get(fmt.Sprintf("http://localhost:%d", serverPort))
		Expect(err).ToNot(HaveOccurred())
		Expect(ioutil.ReadAll(response.Body)).To(ContainSubstring("<h1>BoshStemcells.com</h1>"))
	})

	DescribeTable("IaaSes", func(path, boshUrlPath string) {
		client := &http.Client{
			CheckRedirect: func(r *http.Request, ra []*http.Request) error { return http.ErrUseLastResponse },
		}

		response, err := client.Get(fmt.Sprintf("http://localhost:%d%s", serverPort, path))
		Expect(err).ToNot(HaveOccurred())
		Expect(response.StatusCode).To(Equal(301))
		Expect(response.Header.Get("Location")).To(Equal(fmt.Sprintf("https://bosh.io/d/stemcells/bosh-%s-ubuntu-xenial-go_agent", boshUrlPath)))
	},
		Entry("gcp", "/gcp", "google-kvm"),
		Entry("vsphere", "/vsphere", "vsphere-esxi"),
		Entry("aws", "/aws", "aws-xen-hvm"),
		Entry("azure", "/azure", "azure-hyperv"),
		Entry("openstack", "/openstack", "openstack-kvm"),
		Entry("softlayer", "/softlayer", "softlayer-xen"),
		Entry("vcloud", "/vcloud", "vcloud-esxi"),
		Entry("lite", "/lite", "warden-boshlite"),
	)

	It("Redirects to versions", func() {
		client := &http.Client{
			CheckRedirect: func(r *http.Request, ra []*http.Request) error { return http.ErrUseLastResponse },
		}

		response, err := client.Get(fmt.Sprintf("http://localhost:%d/gcp/1234.56", serverPort))
		Expect(err).ToNot(HaveOccurred())
		Expect(response.StatusCode).To(Equal(301))
		Expect(response.Header.Get("Location")).To(Equal("https://bosh.io/d/stemcells/bosh-google-kvm-ubuntu-xenial-go_agent?v=1234.56"))
	})

	DescribeTable("Autodetects", func(ipAddress, boshUrlPath string) {
		client := &http.Client{
			CheckRedirect: func(r *http.Request, ra []*http.Request) error { return http.ErrUseLastResponse },
		}

		req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/auto", serverPort), nil)
		req.Header.Set("X-Forwarded-For", ipAddress)
		response, err := client.Do(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(response.StatusCode).To(Equal(301))
		Expect(response.Header.Get("Location")).To(Equal(fmt.Sprintf("https://bosh.io/d/stemcells/bosh-%s-ubuntu-xenial-go_agent", boshUrlPath)))
	},
		Entry("gcp", "35.203.192.88", "google-kvm"),
		Entry("aws", "52.210.132.254", "aws-xen-hvm"),
		Entry("azure", "52.164.240.179", "azure-hyperv"),
	)

	It("Gives an error when autodetect fails", func() {
		response, err := http.Get(fmt.Sprintf("http://localhost:%d/auto", serverPort))
		Expect(err).ToNot(HaveOccurred())
		responseBody, err := ioutil.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusNotFound))
		Expect(string(responseBody)).To(Equal("could not autodetect IaaS"))
	})
})
