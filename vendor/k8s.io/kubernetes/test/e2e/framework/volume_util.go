/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
 * This test checks that various VolumeSources are working.
 *
 * There are two ways, how to test the volumes:
 * 1) With containerized server (NFS, Ceph, Gluster, iSCSI, ...)
 * The test creates a server pod, exporting simple 'index.html' file.
 * Then it uses appropriate VolumeSource to import this file into a client pod
 * and checks that the pod can see the file. It does so by importing the file
 * into web server root and loadind the index.html from it.
 *
 * These tests work only when privileged containers are allowed, exporting
 * various filesystems (NFS, GlusterFS, ...) usually needs some mounting or
 * other privileged magic in the server pod.
 *
 * Note that the server containers are for testing purposes only and should not
 * be used in production.
 *
 * 2) With server outside of Kubernetes (Cinder, ...)
 * Appropriate server (e.g. OpenStack Cinder) must exist somewhere outside
 * the tested Kubernetes cluster. The test itself creates a new volume,
 * and checks, that Kubernetes can use it as a volume.
 */

package framework

import (
	"fmt"
	"strconv"
	"time"

	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	imageutils "k8s.io/kubernetes/test/utils/image"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	Kb  int64 = 1000
	Mb  int64 = 1000 * Kb
	Gb  int64 = 1000 * Mb
	Tb  int64 = 1000 * Gb
	KiB int64 = 1024
	MiB int64 = 1024 * KiB
	GiB int64 = 1024 * MiB
	TiB int64 = 1024 * GiB

	// Waiting period for volume server (Ceph, ...) to initialize itself.
	VolumeServerPodStartupSleep = 20 * time.Second
)

// Configuration of one tests. The test consist of:
// - server pod - runs serverImage, exports ports[]
// - client pod - does not need any special configuration
type VolumeTestConfig struct {
	Namespace string
	// Prefix of all pods. Typically the test name.
	Prefix string
	// Name of container image for the server pod.
	ServerImage string
	// Ports to export from the server pod. TCP only.
	ServerPorts []int
	// Commands to run in the container image.
	ServerCmds []string
	// Arguments to pass to the container image.
	ServerArgs []string
	// Volumes needed to be mounted to the server container from the host
	// map <host (source) path> -> <container (dst.) path>
	// if <host (source) path> is empty, mount a tmpfs emptydir
	ServerVolumes map[string]string
	// Wait for the pod to terminate successfully
	// False indicates that the pod is long running
	WaitForCompletion bool
	// ServerNodeName is the spec.nodeName to run server pod on.  Default is any node.
	ServerNodeName string
	// ClientNodeName is the spec.nodeName to run client pod on.  Default is any node.
	ClientNodeName string
	// NodeSelector to use in pod spec (server, client and injector pods).
	NodeSelector map[string]string
}

// VolumeTest contains a volume to mount into a client pod and its
// expected content.
type VolumeTest struct {
	Volume          v1.VolumeSource
	File            string
	ExpectedContent string
}

// NFS-specific wrapper for CreateStorageServer.
func NewNFSServer(cs clientset.Interface, namespace string, args []string) (config VolumeTestConfig, pod *v1.Pod, ip string) {
	config = VolumeTestConfig{
		Namespace:     namespace,
		Prefix:        "nfs",
		ServerImage:   imageutils.GetE2EImage(imageutils.VolumeNFSServer),
		ServerPorts:   []int{2049},
		ServerVolumes: map[string]string{"": "/exports"},
	}
	if len(args) > 0 {
		config.ServerArgs = args
	}
	pod, ip = CreateStorageServer(cs, config)
	return config, pod, ip
}

// GlusterFS-specific wrapper for CreateStorageServer. Also creates the gluster endpoints object.
func NewGlusterfsServer(cs clientset.Interface, namespace string) (config VolumeTestConfig, pod *v1.Pod, ip string) {
	config = VolumeTestConfig{
		Namespace:   namespace,
		Prefix:      "gluster",
		ServerImage: imageutils.GetE2EImage(imageutils.VolumeGlusterServer),
		ServerPorts: []int{24007, 24008, 49152},
	}
	pod, ip = CreateStorageServer(cs, config)

	By("creating Gluster endpoints")
	endpoints := &v1.Endpoints{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Endpoints",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Prefix + "-server",
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: ip,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     "gluster",
						Port:     24007,
						Protocol: v1.ProtocolTCP,
					},
				},
			},
		},
	}
	endpoints, err := cs.CoreV1().Endpoints(namespace).Create(endpoints)
	Expect(err).NotTo(HaveOccurred(), "failed to create endpoints for Gluster server")

	return config, pod, ip
}

// iSCSI-specific wrapper for CreateStorageServer.
func NewISCSIServer(cs clientset.Interface, namespace string) (config VolumeTestConfig, pod *v1.Pod, ip string) {
	config = VolumeTestConfig{
		Namespace:   namespace,
		Prefix:      "iscsi",
		ServerImage: imageutils.GetE2EImage(imageutils.VolumeISCSIServer),
		ServerPorts: []int{3260},
		ServerVolumes: map[string]string{
			// iSCSI container needs to insert modules from the host
			"/lib/modules": "/lib/modules",
		},
	}
	pod, ip = CreateStorageServer(cs, config)
	return config, pod, ip
}

// CephRBD-specific wrapper for CreateStorageServer.
func NewRBDServer(cs clientset.Interface, namespace string) (config VolumeTestConfig, pod *v1.Pod, ip string) {
	config = VolumeTestConfig{
		Namespace:   namespace,
		Prefix:      "rbd",
		ServerImage: imageutils.GetE2EImage(imageutils.VolumeRBDServer),
		ServerPorts: []int{6789},
		ServerVolumes: map[string]string{
			"/lib/modules": "/lib/modules",
		},
	}
	pod, ip = CreateStorageServer(cs, config)

	// Ceph server container needs some time to start. Tests continue working if
	// this sleep is removed, however kubelet logs (and kubectl describe
	// <client pod>) would be cluttered with error messages about non-existing
	// image.
	Logf("sleeping a bit to give ceph server time to initialize")
	time.Sleep(VolumeServerPodStartupSleep)

	return config, pod, ip
}

// Wrapper for StartVolumeServer(). A storage server config is passed in, and a pod pointer
// and ip address string are returned.
// Note: Expect() is called so no error is returned.
func CreateStorageServer(cs clientset.Interface, config VolumeTestConfig) (pod *v1.Pod, ip string) {
	pod = StartVolumeServer(cs, config)
	Expect(pod).NotTo(BeNil(), "storage server pod should not be nil")
	ip = pod.Status.PodIP
	Expect(len(ip)).NotTo(BeZero(), fmt.Sprintf("pod %s's IP should not be empty", pod.Name))
	Logf("%s server pod IP address: %s", config.Prefix, ip)
	return pod, ip
}

// Starts a container specified by config.serverImage and exports all
// config.serverPorts from it. The returned pod should be used to get the server
// IP address and create appropriate VolumeSource.
func StartVolumeServer(client clientset.Interface, config VolumeTestConfig) *v1.Pod {
	podClient := client.CoreV1().Pods(config.Namespace)

	portCount := len(config.ServerPorts)
	serverPodPorts := make([]v1.ContainerPort, portCount)

	for i := 0; i < portCount; i++ {
		portName := fmt.Sprintf("%s-%d", config.Prefix, i)

		serverPodPorts[i] = v1.ContainerPort{
			Name:          portName,
			ContainerPort: int32(config.ServerPorts[i]),
			Protocol:      v1.ProtocolTCP,
		}
	}

	volumeCount := len(config.ServerVolumes)
	volumes := make([]v1.Volume, volumeCount)
	mounts := make([]v1.VolumeMount, volumeCount)

	i := 0
	for src, dst := range config.ServerVolumes {
		mountName := fmt.Sprintf("path%d", i)
		volumes[i].Name = mountName
		if src == "" {
			volumes[i].VolumeSource.EmptyDir = &v1.EmptyDirVolumeSource{}
		} else {
			volumes[i].VolumeSource.HostPath = &v1.HostPathVolumeSource{
				Path: src,
			}
		}

		mounts[i].Name = mountName
		mounts[i].ReadOnly = false
		mounts[i].MountPath = dst

		i++
	}

	serverPodName := fmt.Sprintf("%s-server", config.Prefix)
	By(fmt.Sprint("creating ", serverPodName, " pod"))
	privileged := new(bool)
	*privileged = true

	restartPolicy := v1.RestartPolicyAlways
	if config.WaitForCompletion {
		restartPolicy = v1.RestartPolicyNever
	}
	serverPod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serverPodName,
			Labels: map[string]string{
				"role": serverPodName,
			},
		},

		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  serverPodName,
					Image: config.ServerImage,
					SecurityContext: &v1.SecurityContext{
						Privileged: privileged,
					},
					Command:      config.ServerCmds,
					Args:         config.ServerArgs,
					Ports:        serverPodPorts,
					VolumeMounts: mounts,
				},
			},
			Volumes:       volumes,
			RestartPolicy: restartPolicy,
			NodeName:      config.ServerNodeName,
			NodeSelector:  config.NodeSelector,
		},
	}

	var pod *v1.Pod
	serverPod, err := podClient.Create(serverPod)
	// ok if the server pod already exists. TODO: make this controllable by callers
	if err != nil {
		if apierrs.IsAlreadyExists(err) {
			Logf("Ignore \"already-exists\" error, re-get pod...")
			By(fmt.Sprintf("re-getting the %q server pod", serverPodName))
			serverPod, err = podClient.Get(serverPodName, metav1.GetOptions{})
			ExpectNoError(err, "Cannot re-get the server pod %q: %v", serverPodName, err)
			pod = serverPod
		} else {
			ExpectNoError(err, "Failed to create %q pod: %v", serverPodName, err)
		}
	}
	if config.WaitForCompletion {
		ExpectNoError(WaitForPodSuccessInNamespace(client, serverPod.Name, serverPod.Namespace))
		ExpectNoError(podClient.Delete(serverPod.Name, nil))
	} else {
		ExpectNoError(WaitForPodRunningInNamespace(client, serverPod))
		if pod == nil {
			By(fmt.Sprintf("locating the %q server pod", serverPodName))
			pod, err = podClient.Get(serverPodName, metav1.GetOptions{})
			ExpectNoError(err, "Cannot locate the server pod %q: %v", serverPodName, err)
		}
	}
	return pod
}

// Clean both server and client pods.
func VolumeTestCleanup(f *Framework, config VolumeTestConfig) {
	By(fmt.Sprint("cleaning the environment after ", config.Prefix))

	defer GinkgoRecover()

	client := f.ClientSet

	err := DeletePodWithWaitByName(f, client, config.Prefix+"-client", config.Namespace)
	Expect(err).To(BeNil(), "Failed waiting for pod %v to be not running in namespace %v", config.Prefix+"-client", config.Namespace)

	if config.ServerImage != "" {
		err := DeletePodWithWaitByName(f, client, config.Prefix+"-server", config.Namespace)
		Expect(err).To(BeNil(), "Failed waiting for pod %v to be not running in namespace %v", config.Prefix+"-server", config.Namespace)
	}
}

// Start a client pod using given VolumeSource (exported by startVolumeServer())
// and check that the pod sees expected data, e.g. from the server pod.
// Multiple VolumeTests can be specified to mount multiple volumes to a single
// pod.
func TestVolumeClient(client clientset.Interface, config VolumeTestConfig, fsGroup *int64, tests []VolumeTest) {
	By(fmt.Sprint("starting ", config.Prefix, " client"))
	var gracePeriod int64 = 1
	clientPod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Prefix + "-client",
			Labels: map[string]string{
				"role": config.Prefix + "-client",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:       config.Prefix + "-client",
					Image:      BusyBoxImage,
					WorkingDir: "/opt",
					// An imperative and easily debuggable container which reads vol contents for
					// us to scan in the tests or by eye.
					// We expect that /opt is empty in the minimal containers which we use in this test.
					Command: []string{
						"/bin/sh",
						"-c",
						"while true ; do cat /opt/0/index.html ; sleep 2 ; ls -altrh /opt/  ; sleep 2 ; done ",
					},
					VolumeMounts: []v1.VolumeMount{},
				},
			},
			TerminationGracePeriodSeconds: &gracePeriod,
			SecurityContext: &v1.PodSecurityContext{
				SELinuxOptions: &v1.SELinuxOptions{
					Level: "s0:c0,c1",
				},
			},
			Volumes:      []v1.Volume{},
			NodeName:     config.ClientNodeName,
			NodeSelector: config.NodeSelector,
		},
	}
	podsNamespacer := client.CoreV1().Pods(config.Namespace)

	if fsGroup != nil {
		clientPod.Spec.SecurityContext.FSGroup = fsGroup
	}

	for i, test := range tests {
		volumeName := fmt.Sprintf("%s-%s-%d", config.Prefix, "volume", i)
		clientPod.Spec.Containers[0].VolumeMounts = append(clientPod.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      volumeName,
			MountPath: fmt.Sprintf("/opt/%d", i),
		})
		clientPod.Spec.Volumes = append(clientPod.Spec.Volumes, v1.Volume{
			Name:         volumeName,
			VolumeSource: test.Volume,
		})
	}
	clientPod, err := podsNamespacer.Create(clientPod)
	if err != nil {
		Failf("Failed to create %s pod: %v", clientPod.Name, err)
	}
	ExpectNoError(WaitForPodRunningInNamespace(client, clientPod))

	By("Checking that text file contents are perfect.")
	for i, test := range tests {
		fileName := fmt.Sprintf("/opt/%d/%s", i, test.File)
		_, err = LookForStringInPodExec(config.Namespace, clientPod.Name, []string{"cat", fileName}, test.ExpectedContent, time.Minute)
		Expect(err).NotTo(HaveOccurred(), "failed: finding the contents of the mounted file %s.", fileName)
	}

	if fsGroup != nil {
		By("Checking fsGroup is correct.")
		_, err = LookForStringInPodExec(config.Namespace, clientPod.Name, []string{"ls", "-ld", "/opt/0"}, strconv.Itoa(int(*fsGroup)), time.Minute)
		Expect(err).NotTo(HaveOccurred(), "failed: getting the right privileges in the file %v", int(*fsGroup))
	}
}

// Insert index.html with given content into given volume. It does so by
// starting and auxiliary pod which writes the file there.
// The volume must be writable.
func InjectHtml(f *Framework, client clientset.Interface, config VolumeTestConfig, volume v1.VolumeSource, content string) {
	By(fmt.Sprint("starting ", config.Prefix, " injector"))
	podClient := client.CoreV1().Pods(config.Namespace)

	injectPod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Prefix + "-injector",
			Labels: map[string]string{
				"role": config.Prefix + "-injector",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    config.Prefix + "-injector",
					Image:   BusyBoxImage,
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", "echo '" + content + "' > /mnt/index.html && chmod o+rX /mnt /mnt/index.html"},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      config.Prefix + "-volume",
							MountPath: "/mnt",
						},
					},
				},
			},
			SecurityContext: &v1.PodSecurityContext{
				SELinuxOptions: &v1.SELinuxOptions{
					Level: "s0:c0,c1",
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
			Volumes: []v1.Volume{
				{
					Name:         config.Prefix + "-volume",
					VolumeSource: volume,
				},
			},
			NodeSelector: config.NodeSelector,
		},
	}

	defer func() {
		err := DeletePodWithWaitByName(f, client, injectPod.GetName(), config.Namespace)
		Expect(err).To(BeNil(), "Failed deleting pod %v in namespace %v", injectPod.GetName(), config.Namespace)
	}()

	injectPod, err := podClient.Create(injectPod)
	ExpectNoError(err, "Failed to create injector pod: %v", err)
	err = WaitForPodSuccessInNamespace(client, injectPod.Name, injectPod.Namespace)
	Expect(err).NotTo(HaveOccurred())
}

func CreateGCEVolume() (*v1.PersistentVolumeSource, string) {
	diskName, err := CreatePDWithRetry()
	ExpectNoError(err)
	return &v1.PersistentVolumeSource{
		GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
			PDName:   diskName,
			FSType:   "ext3",
			ReadOnly: false,
		},
	}, diskName
}
