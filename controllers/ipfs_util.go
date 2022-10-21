package controllers

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/redhat-et/ipfs-operator/api/v1alpha1"
)

// ci "github.com/libp2p/go-libp2p-core/crypto"
// peer "github.com/libp2p/go-libp2p-core/peer"

// func newClusterSecret() (string, error) {
// 	const secretLen = 32
// 	buf := make([]byte, secretLen)
// 	_, err := rand.Read(buf)
// 	if err != nil {
// 		return "", err
// 	}
// 	return hex.EncodeToString(buf), nil
// }

// // newKey Generates a new private key and returns that along with the identity.
// func newKey() (ci.PrivKey, peer.ID, error) {
// 	const edDSAKeyLen = 4096
// 	priv, pub, err := ci.GenerateKeyPair(ci.Ed25519, edDSAKeyLen)
// 	if err != nil {
// 		return nil, "", err
// 	}
// 	peerid, err := peer.IDFromPublicKey(pub)
// 	if err != nil {
// 		return nil, "", err
// 	}
// 	return priv, peerid, nil
// }

// // generateIdentity Generates a new key and returns the peer ID and private key
// // encoded as a base64 string using standard encoding, or an error if the key could not be generated.
// func generateIdentity() (peer.ID, string, error) {
// 	priv, peerid, err := newKey()
// 	if err != nil {
// 		return "", "", fmt.Errorf("cannot generate new key: %w", err)
// 	}
// 	privBytes, err := ci.MarshalPrivateKey(priv)
// 	if err != nil {
// 		return "", "", fmt.Errorf("cannot get bytes from private key: %w", err)
// 	}
// 	privStr := base64.StdEncoding.EncodeToString(privBytes)
// 	return peerid, privStr, nil
// }

// EnsureRelayCircuitInfo Returns information about the configured CircuitRelays.
func (r *IpfsClusterReconciler) EnsureRelayCircuitInfo(ctx context.Context, ipfs *v1alpha1.IpfsCluster) (
	relayPeers []peer.AddrInfo, relayStatic []ma.Multiaddr, err error) {
	if relayPeers, err = r.getCircuitInfo(ctx, ipfs); err != nil {
		return
	}
	if relayStatic, err = staticAddrsFromRelayPeers(relayPeers); err != nil {
		return
	}
	return
}
