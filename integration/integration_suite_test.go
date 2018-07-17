package integration_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var (
	serverPort int
	session    *gexec.Session
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)

	BeforeSuite(func() {
		var (
			err       error
			pathToBin string
		)

		pathToBin, err = gexec.Build("github.com/benchapman/boshstemcells")
		Expect(err).ToNot(HaveOccurred())

		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}

		serverPort = listener.Addr().(*net.TCPAddr).Port

		listener.Close()

		cmd := exec.Command(pathToBin)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", serverPort))
		pwd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		cmd.Dir = filepath.Join(pwd, "..")

		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(time.Second)
	})

	AfterSuite(func() {
		session.Kill()
		gexec.CleanupBuildArtifacts()
	})

	RunSpecs(t, "Integration Suite")
}
