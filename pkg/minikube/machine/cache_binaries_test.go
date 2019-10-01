package machine

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
)

type copyFailRunner struct {
	command.Runner
}

func (copyFailRunner) Copy(a assets.CopyableFile) error {
	return fmt.Errorf("test error during copy file")
}

func TestCopyBinary(t *testing.T) {

	fakeCommandRunnerCopyFail := func() command.Runner {
		r := command.NewFakeCommandRunner()
		return copyFailRunner{r}
	}

	var tc = []struct {
		lastUpdateCheckFilePath string
		src, dst, desc          string
		err                     bool
		runner                  command.Runner
	}{
		{
			desc:   "not existing src",
			dst:    "/tmp/testCopyBinary1",
			src:    "/tmp/testCopyBinary2",
			err:    true,
			runner: command.NewFakeCommandRunner(),
		},
		{
			desc:   "src /etc/hosts",
			dst:    "/tmp/testCopyBinary1",
			src:    "/etc/hosts",
			err:    false,
			runner: command.NewFakeCommandRunner(),
		},
		{
			desc:   "existing src, dst without permissions",
			dst:    "/etc/passwd",
			src:    "/etc/hosts",
			err:    true,
			runner: fakeCommandRunnerCopyFail(),
		},
	}
	for _, test := range tc {
		t.Run(test.desc, func(t *testing.T) {
			err := CopyBinary(test.runner, test.src, test.dst)
			if err != nil && !test.err {
				t.Fatalf("Error %v expected but not occured", err)
			}
			if err == nil && test.err {
				t.Fatal("Unexpected error")
			}
		})
	}
}

func TestCacheBinariesForBootstrapper(t *testing.T) {
	var tc = []struct {
		version, clusterBootstrapper string
		err                          bool
	}{
		{
			version:             "v1.16.0",
			clusterBootstrapper: bootstrapper.BootstrapperTypeKubeadm,
		},
		{
			version:             "invalid version",
			clusterBootstrapper: bootstrapper.BootstrapperTypeKubeadm,
			err:                 true,
		},
	}
	for _, test := range tc {
		t.Run(test.version, func(t *testing.T) {
			err := CacheBinariesForBootstrapper(test.version, test.clusterBootstrapper)
			if err != nil && !test.err {
				t.Fatalf("Got unexpected error %v", err)
			}
			if err == nil && test.err {
				t.Fatalf("Expected error but got %v", err)
			}
		})
	}
}
func TestCacheBinary(t *testing.T) {
	oldMinikubeHome := os.Getenv("MINIKUBE_HOME")
	defer os.Setenv("MINIKUBE_HOME", oldMinikubeHome)

	minikubeHome, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		t.Fatalf("error during creating tmp dir: %v", err)
	}
	defer os.RemoveAll(minikubeHome)

	var tc = []struct {
		desc, version, osName, archName   string
		minikubeHome, binary, description string
		err                               bool
	}{
		{
			desc:         "ok kubeadm",
			version:      "v1.16.0",
			osName:       runtime.GOOS,
			archName:     runtime.GOARCH,
			binary:       "kubeadm",
			err:          false,
			minikubeHome: minikubeHome,
		},
		{
			desc:         "minikube home is dev/null",
			version:      "v1.16.0",
			osName:       runtime.GOOS,
			archName:     "arm",
			binary:       "kubectl",
			err:          true,
			minikubeHome: "/dev/null",
		},
		{
			desc:         "minikube home in etc and arm runtime",
			version:      "v1.16.0",
			osName:       runtime.GOOS,
			archName:     "arm",
			binary:       "kubectl",
			err:          true,
			minikubeHome: "/etc",
		},
		{
			desc:         "minikube home in etc",
			version:      "v1.16.0",
			osName:       runtime.GOOS,
			archName:     runtime.GOARCH,
			binary:       "kubectl",
			err:          true,
			minikubeHome: "/etc",
		},
		{
			desc:         "binary foo",
			version:      "v1.16.0",
			osName:       runtime.GOOS,
			archName:     runtime.GOARCH,
			binary:       "foo",
			err:          true,
			minikubeHome: minikubeHome,
		},
		{
			desc:         "version 9000",
			version:      "v9000",
			osName:       runtime.GOOS,
			archName:     runtime.GOARCH,
			binary:       "foo",
			err:          true,
			minikubeHome: minikubeHome,
		},
		{
			desc:         "bad os",
			version:      "v1.16.0",
			osName:       "no-such-os",
			archName:     runtime.GOARCH,
			binary:       "kubectl",
			err:          true,
			minikubeHome: minikubeHome,
		},
	}
	for _, test := range tc {
		t.Run(test.desc, func(t *testing.T) {
			os.Setenv("MINIKUBE_HOME", test.minikubeHome)
			_, err := CacheBinary(test.binary, test.version, test.osName, test.archName)
			if err != nil && !test.err {
				t.Fatalf("Got unexpected error %v", err)
			}
			if err == nil && test.err {
				t.Fatalf("Expected error but got %v", err)
			}
		})
	}
}
