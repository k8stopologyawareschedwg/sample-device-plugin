package sample_device_plugin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/gomega"

	"k8s.io/klog"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	// Client defines the API client to run CRUD operations, that will be used for testing
	Client client.Client
	// K8sClient defines k8s client to run subresource operations, for example you should use it to get pod logs
	K8sClient *kubernetes.Clientset
	// TestingNamespace is the namespace the tests will use for running test pods
	TestingNamespace *corev1.Namespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sample-device-plugin-test-",
		},
	}
)

func Setup() error {
	var err error
	Client, err = New()
	if err != nil {
		klog.Info("Failed to initialize client, check the KUBECONFIG env variable", err.Error())
		return fmt.Errorf("New failed")
	}
	K8sClient, err = NewK8s()
	if err != nil {
		klog.Info("Failed to initialize k8s client, check the KUBECONFIG env variable", err.Error())
		return fmt.Errorf("NewK8S failed")
	}

	// create test namespace

	err = Client.Create(context.TODO(), TestingNamespace)
	if err != nil {
		klog.Info("Failed to create the testing namespace", err.Error())
		return fmt.Errorf("Create TestingNamespace failed")
	}

	klog.Infof("Using testing namespace %q", TestingNamespace.Name)
	return nil
}

func Teardown() error {
	if Client == nil {
		// nothing to do!
		return nil
	}

	if TestingNamespace == nil {
		klog.Infof("Inconsistent state: TestingNamespace is nil")
		return fmt.Errorf("NS missing")
	}

	nsName := TestingNamespace.Name
	err := Client.Delete(context.TODO(), TestingNamespace)
	if err != nil {
		klog.Infof("Failed deleting namespace %q", nsName)
		return fmt.Errorf("NS delete failed")
	}

	err = WaitForNamespaceDeletion(nsName, 5*time.Minute)
	if err != nil {
		klog.Infof("Timeout waiting for the deletion of namespace %q", nsName)
		return fmt.Errorf("NS delete timeout")
	}
	return nil
}

// New returns a controller-runtime client.
func New() (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	c, err := client.New(cfg, client.Options{})
	return c, err
}

// NewK8s returns a kubernetes clientset
func NewK8s() (*kubernetes.Clientset, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Exit(err.Error())
	}
	return clientset, nil
}

func GetWithRetry(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	var err error
	EventuallyWithOffset(1, func() error {
		err = Client.Get(ctx, key, obj)
		if err != nil {
			klog.Infof("Getting %s failed, retrying: %v", key.Name, err)
		}
		return err
	}, 1*time.Minute, 10*time.Second).ShouldNot(HaveOccurred(), "Max numbers of retries reached")
	return err
}

// WaitForNamespaceDeletion waits until the namespace will be removed from the cluster
func WaitForNamespaceDeletion(name string, timeout time.Duration) error {
	key := types.NamespacedName{
		Name:      name,
		Namespace: metav1.NamespaceNone,
	}
	return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		ns := &corev1.Namespace{}
		if err := Client.Get(context.TODO(), key, ns); errors.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	})
}

// GetTestPod returns pod with the busybox image
func GetTestPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-",
			Labels: map[string]string{
				"test": "",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "test",
					Image:   "busybox",
					Command: []string{"sleep", "10h"},
				},
			},
		},
	}
}

// WaitForPodDeletion waits until the pod will be removed from the cluster
func WaitForPodDeletion(pod *corev1.Pod, timeout time.Duration) error {
	key := types.NamespacedName{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}
	return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		pod := &corev1.Pod{}
		if err := Client.Get(context.TODO(), key, pod); errors.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	})
}

// WaitForPodCondition waits until the pod will have specified condition type with the expected status
func WaitForPodCondition(pod *corev1.Pod, conditionType corev1.PodConditionType, conditionStatus corev1.ConditionStatus, timeout time.Duration) error {
	key := types.NamespacedName{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}
	return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		updatedPod := &corev1.Pod{}
		if err := Client.Get(context.TODO(), key, updatedPod); err != nil {
			return false, nil
		}

		for _, c := range updatedPod.Status.Conditions {
			if c.Type == conditionType && c.Status == conditionStatus {
				return true, nil
			}
		}
		return false, nil
	})
}

// WaitForPodPhase waits until the pod will have specified phase
func WaitForPodPhase(pod *corev1.Pod, phase corev1.PodPhase, timeout time.Duration) error {
	key := types.NamespacedName{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}
	return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		updatedPod := &corev1.Pod{}
		if err := Client.Get(context.TODO(), key, updatedPod); err != nil {
			return false, nil
		}

		if updatedPod.Status.Phase == phase {
			return true, nil
		}

		return false, nil
	})
}

// GetPodLogs returns logs of the specified pod
func GetPodLogs(c *kubernetes.Clientset, pod *corev1.Pod) (string, error) {
	logStream, err := c.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{}).Stream(context.TODO())
	if err != nil {
		return "", err
	}
	defer logStream.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, logStream); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExecCommandOnPod returns the output of the command execution on the pod
func ExecCommandOnPod(pod *corev1.Pod, command []string) ([]byte, error) {
	initialArgs := []string{
		"exec",
		"-i",
		"-n", pod.Namespace,
		pod.Name,
		"--",
	}
	args := append(initialArgs, command...)
	return ExecAndLogCommand("kubectl", args...)
}

func ExecAndLogCommand(name string, arg ...string) ([]byte, error) {
	out, err := exec.Command(name, arg...).CombinedOutput()
	klog.Infof("run command '%v':\n  out=%s\n  err=%v", arg, out, err)
	return out, err
}
