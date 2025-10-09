package vm

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
)

type KubeVirtProvider struct {
	client kubecli.KubevirtClient
}

func NewKubeVirtProvider() *KubeVirtProvider {
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(
		kubecli.DefaultClientConfig(&pflag.FlagSet{}),
	)
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt client: %v\n", err)
	}

	return &KubeVirtProvider{
		client: virtClient,
	}
}

func (k *KubeVirtProvider) Name() string {
	return "kubevirt"
}

func (k *KubeVirtProvider) ProviderID() string {
	return "94969e49-3804-4eb6-b6b6-6473fe2f42df"
}

func (k *KubeVirtProvider) Description() string {
	return "KubeVirt VM Service Provider"
}

func (k *KubeVirtProvider) CreateVM(ctx context.Context, request Request) (DeclaredVM, error) {
	logger := zap.S().Named("kubevirt:create_vm")

	logger.Info("Starting deployment for Virtual Machine")

	// Create Namespace for the Virtual Machine
	namespace := request.Namespace
	logger.Info("Creating namespace ", namespace)
	// Check Namespace exists
	_, err := k.client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		// Create Namespace
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		_, err = k.client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		if err != nil {
			logger.Error("Error occurred when creating namespace", err)
			return DeclaredVM{}, fmt.Errorf("failed to create namespace %s: %w", namespace, err)
		}
	}
	logger.Info("Successfully created namespace ", "Namespace ", namespace)

	// Create the VirtualMachine object
	memory := resource.MustParse(fmt.Sprintf("%dGi", request.Ram))
	virtualMachine := &kubevirtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", request.VMName),
			Namespace:    namespace,
			Labels: map[string]string{
				"app-id": request.RequestId,
			},
		},
		Spec: kubevirtv1.VirtualMachineSpec{
			RunStrategy: &[]kubevirtv1.VirtualMachineRunStrategy{kubevirtv1.RunStrategyRerunOnFailure}[0],
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					AccessCredentials: []kubevirtv1.AccessCredential{
						{
							SSHPublicKey: &kubevirtv1.SSHPublicKeyAccessCredential{
								Source: kubevirtv1.SSHPublicKeyAccessCredentialSource{
									Secret: &kubevirtv1.AccessCredentialSecretSource{
										SecretName: "myssh",
									},
								},
								PropagationMethod: kubevirtv1.SSHPublicKeyAccessCredentialPropagationMethod{
									NoCloud: &kubevirtv1.NoCloudSSHPublicKeyAccessCredentialPropagation{},
								},
							},
						},
					},
					Architecture: "amd64",
					Domain: kubevirtv1.DomainSpec{
						CPU: &kubevirtv1.CPU{
							Cores: uint32(request.Cpu),
						},
						Memory: &kubevirtv1.Memory{
							Guest: &memory,
						},
						Devices: kubevirtv1.Devices{
							Disks: []kubevirtv1.Disk{
								{
									Name:      fmt.Sprintf("%s-disk", request.VMName),
									BootOrder: &[]uint{1}[0],
									DiskDevice: kubevirtv1.DiskDevice{
										Disk: &kubevirtv1.DiskTarget{
											Bus: kubevirtv1.DiskBusVirtio,
										},
									},
								},
								{
									Name:      "cloudinitdisk",
									BootOrder: &[]uint{2}[0],
									DiskDevice: kubevirtv1.DiskDevice{
										Disk: &kubevirtv1.DiskTarget{
											Bus: kubevirtv1.DiskBusVirtio,
										},
									},
								},
							},
							Interfaces: []kubevirtv1.Interface{
								{
									Name: "myvmnic",
									InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
										Bridge: &kubevirtv1.InterfaceBridge{},
									},
								},
							},
							Rng: &kubevirtv1.Rng{},
						},
						Features: &kubevirtv1.Features{
							ACPI: kubevirtv1.FeatureState{},
							SMM: &kubevirtv1.FeatureState{
								Enabled: &[]bool{true}[0],
							},
						},
						Machine: &kubevirtv1.Machine{
							Type: "pc-q35-rhel9.6.0",
						},
					},
					Networks: []kubevirtv1.Network{
						{
							Name: "myvmnic",
							NetworkSource: kubevirtv1.NetworkSource{
								Pod: &kubevirtv1.PodNetwork{},
							},
						},
					},
					TerminationGracePeriodSeconds: &[]int64{180}[0],
					Volumes: []kubevirtv1.Volume{
						{
							Name: fmt.Sprintf("%s-disk", request.VMName),
							VolumeSource: kubevirtv1.VolumeSource{
								ContainerDisk: &kubevirtv1.ContainerDiskSource{
									Image: request.OsImage,
								},
							},
						},
						{
							Name: "cloudinitdisk",
							VolumeSource: kubevirtv1.VolumeSource{
								CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
									UserData: k.generateCloudInitUserData(request.VMName, &request),
								},
							},
						},
					},
				},
			},
		},
	}

	// Create the VirtualMachine in the cluster
	_, err = k.client.VirtualMachine(namespace).Create(ctx, virtualMachine, metav1.CreateOptions{})
	if err != nil {
		return DeclaredVM{}, fmt.Errorf("failed to create VirtualMachine: %w", err)
	}

	return DeclaredVM{ID: request.RequestId, RequestInfo: request}, nil
}

func (k *KubeVirtProvider) GetVM(ctx context.Context, vmID string) (DeclaredVM, error) {
	return DeclaredVM{}, nil
}

func (k *KubeVirtProvider) DeleteVM(ctx context.Context, vmID string) (DeclaredVM, error) {
	return DeclaredVM{}, nil
}

func (k *KubeVirtProvider) ListVMs(ctx context.Context) ([]DeclaredVM, error) {
	return []DeclaredVM{}, nil
}

// generateCloudInitUserData generates cloud-init user data for the VM
func (k *KubeVirtProvider) generateCloudInitUserData(appName string, vm *Request) string {
	return fmt.Sprintf(`#cloud-config
user: %s
password: auto-generated-pass
chpasswd: { expire: False }
hostname: %s
`, vm.OsImage, appName)
}
