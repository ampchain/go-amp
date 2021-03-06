package p2pv2

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io/ioutil"
	math_rand "math/rand"
	"os"
	"strconv"
	"time"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/ampchain/go-amp/common/config"
	crypto_client "github.com/ampchain/go-amp/crypto/client"
	"github.com/ampchain/go-amp/crypto/hash"
	"github.com/ampchain/go-amp/pb"
)

// GenerateKeyPairWithPath generate amper net key pair
func GenerateKeyPairWithPath(path string) error {
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return err
	}

	if len(path) == 0 {
		path = config.DefaultNetKeyPath
	}

	privData, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(path, 0777); err != nil {
		return err
	}

	return ioutil.WriteFile(path+"net_private.key", []byte(base64.StdEncoding.EncodeToString(privData)), 0700)
}

// GetPeerIDFromPath return peer id of given private key path
func GetPeerIDFromPath(keypath string) (string, error) {
	pk, err := GetKeyPairFromPath(keypath)
	if err != nil {
		return "", err
	}

	pid, err := peer.IDFromPrivateKey(pk)
	if err != nil {
		return "", err
	}
	return pid.Pretty(), nil
}

// GetKeyPairFromPath get amper net key from file path
func GetKeyPairFromPath(path string) (crypto.PrivKey, error) {
	if len(path) == 0 {
		path = config.DefaultNetKeyPath
	}

	data, err := ioutil.ReadFile(path + "net_private.key")
	if err != nil {
		return nil, err
	}

	privData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalPrivateKey(privData)
}

// GenerateUniqueRandList get a random unique number list
func GenerateUniqueRandList(size int, max int) []int {
	r := math_rand.New(math_rand.NewSource(time.Now().UnixNano()))
	if max <= 0 || size <= 0 {
		return nil
	}
	if size > max {
		size = max
	}
	randList := r.Perm(max)
	return randList[:size]
}

// AChainAddrInfo my AmpChain addr info
type AChainAddrInfo struct {
	Addr   string
	Pubkey []byte
	Prikey []byte
	PeerID string
}

// GetAuthRequest get auth request for authentication
func GetAuthRequest(v *AChainAddrInfo) (*pb.IdentityAuth, error) {
	cryptoClient, err := crypto_client.CreateCryptoClientFromJSONPublicKey(v.Pubkey)
	if err != nil {
		return nil, errors.New("GetAuthRequest: Create crypto client error")
	}

	privateKey, err := cryptoClient.GetEcdsaPrivateKeyFromJSON(v.Prikey)
	if err != nil {
		return nil, err
	}

	timeUnix := time.Now().Unix()
	ts := strconv.FormatInt(timeUnix, 10)

	digestHash := hash.UsingSha256([]byte(v.PeerID + v.Addr + ts))
	sign, err := cryptoClient.SignECDSA(privateKey, digestHash)
	if err != nil {
		return nil, err
	}

	identityAuth := &pb.IdentityAuth{
		Sign:      sign,
		Pubkey:    v.Pubkey,
		Addr:      v.Addr,
		PeerID:    v.PeerID,
		Timestamp: ts,
	}

	return identityAuth, nil
}
