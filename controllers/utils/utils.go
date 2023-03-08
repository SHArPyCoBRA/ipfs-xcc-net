package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/alecthomas/units"
	"github.com/go-logr/logr"

	ci "github.com/libp2p/go-libp2p/core/crypto"

	peer "github.com/libp2p/go-libp2p/core/peer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// CreateOrPatchTrackedObjects Goes through the map of tracked objects and attempts to
// apply the ctrl.createOrPatch function to each one. This function will return a
// boolean indicating whether or not the requeue should be set to true.
func CreateOrPatchTrackedObjects(
	ctx context.Context,
	trackedObjects map[client.Object]controllerutil.MutateFn,
	client client.Client,
	log logr.Logger,
) bool {
	var requeue bool
	var err error
	for obj, mut := range trackedObjects {
		var result controllerutil.OperationResult
		kind := obj.GetObjectKind().GroupVersionKind()
		name := obj.GetName()
		result, err = controllerutil.CreateOrPatch(ctx, client, obj, mut)
		if err != nil {
			log.Error(err, "error creating object", "objname", name, "objKind", kind.Kind, "result", result)
			requeue = true
		} else {
			log.Info("object changed", "objName", name, "objKind", kind.Kind, "result", result)
		}
	}
	return requeue
}

// ErrFunc Returns a function which returns the provided error when called.
func ErrFunc(err error) controllerutil.MutateFn {
	return func() error {
		return err
	}
}

// IPFSContainerResources Returns the resource requests/requirements for running a single IPFS Container
// depending on the storage requested by the user.
func IPFSContainerResources(ipfsStorageBytes int64) (ipfsResources corev1.ResourceRequirements) {
	// Determine resource constraints from how much we are storing.
	// for every TB of storage, Request 1GB of memory and limit if we exceed 2x this amount.
	// memory floor is 2G.
	// The CPU requirement starts at 4 cores and increases by 500m for every TB of storage
	// many block storage providers have a maximum block storage of 16TB, so in this case, the
	// biggest node we would allocate would request a minimum allocation of 16G of RAM and 12 cores
	// and would permit usage up to twice this size

	ipfsStorageTB := ipfsStorageBytes / int64(units.Tebibyte)
	ipfsMilliCoresMin := 250 + (500 * ipfsStorageTB)
	ipfsRAMGBMin := ipfsStorageTB
	if ipfsRAMGBMin < 2 {
		ipfsRAMGBMin = 1
	}

	ipfsRAMMinQuantity := resource.NewScaledQuantity(ipfsRAMGBMin, resource.Giga)
	ipfsRAMMaxQuantity := resource.NewScaledQuantity(2*ipfsRAMGBMin, resource.Giga)
	ipfsCoresMinQuantity := resource.NewScaledQuantity(ipfsMilliCoresMin, resource.Milli)
	ipfsCoresMaxQuantity := resource.NewScaledQuantity(2*ipfsMilliCoresMin, resource.Milli)

	ipfsResources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: *ipfsRAMMinQuantity,
			corev1.ResourceCPU:    *ipfsCoresMinQuantity,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: *ipfsRAMMaxQuantity,
			corev1.ResourceCPU:    *ipfsCoresMaxQuantity,
		},
	}
	return
}

// randomKey Returns a cryptographically-secure generated key.
func randomKey(len int) (buf []byte, err error) {
	buf = make([]byte, len)
	if _, err = rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// NewClusterSecret Returns a new IPFS Cluster secret.
func NewClusterSecret() (string, error) {
	buf, err := randomKey(32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// NewSwarmKey Generates and returns a key used for hosting a private swarm.
func NewSwarmKey() (string, error) {
	const swarmPrefix = "/key/swarm/psk/1.0.0"
	const multiBase = "/base16/"
	buf, err := randomKey(32)
	if err != nil {
		return "", err
	}
	key := hex.EncodeToString(buf)
	swarmKey := fmt.Sprintf("%s\n%s\n%s", swarmPrefix, multiBase, key)
	return swarmKey, nil
}

// NewKey Generates a new private key and returns that along with the identity.
func NewKey() (ci.PrivKey, peer.ID, error) {
	const edDSAKeyLen = 4096
	priv, pub, err := ci.GenerateKeyPair(ci.Ed25519, edDSAKeyLen)
	if err != nil {
		return nil, "", err
	}
	peerid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		return nil, "", err
	}
	return priv, peerid, nil
}

// GenerateIdentity Generates a new key and returns the peer ID and private key
// encoded as a base64 string using standard encoding, or an error if the key could not be generated.
func GenerateIdentity() (peerid peer.ID, privStr string, err error) {
	var privateKey ci.PrivKey
	privateKey, peerid, err = NewKey()
	if err != nil {
		return "", "", fmt.Errorf("cannot generate new key: %w", err)
	}
	privBytes, err := ci.MarshalPrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("cannot get bytes from private key: %w", err)
	}
	privStr = base64.StdEncoding.EncodeToString(privBytes)
	return
}
