package common

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	ctfconfig "github.com/smartcontractkit/chainlink-testing-framework/lib/config"
	"github.com/smartcontractkit/chainlink-testing-framework/lib/k8s/environment"
	"github.com/smartcontractkit/chainlink-testing-framework/lib/k8s/pkg/helm/chainlink"
	mock_adapter "github.com/smartcontractkit/chainlink-testing-framework/lib/k8s/pkg/helm/mock-adapter"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
	"github.com/smartcontractkit/chainlink/integration-tests/docker/test_env"
	"github.com/smartcontractkit/chainlink/v2/core/services/job"

	chainconfig "github.com/smartcontractkit/chainlink-starknet/integration-tests/config"
	"github.com/smartcontractkit/chainlink-starknet/integration-tests/testconfig"
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

type Common struct {
	ChainDetails    *chainconfig.Config
	TestEnvDetails  *TestEnvDetails
	Env             *environment.Environment
	RPCDetails      *RPCDetails
	ChainlinkConfig string
	TestConfig      *testconfig.TestConfig
}

type TestEnvDetails struct {
	TestDuration time.Duration
	K8Config     *environment.Config
	NodeOpts     []test_env.ClNodeOption
}

type RPCDetails struct {
	RPCL1Internal       string
	RPCL2Internal       string
	RPCL2InternalAPIKey string
	RPCL1External       string
	RPCL2External       string
	MockServerURL       string
	MockServerEndpoint  string
	P2PPort             string
}

func New(testConfig *testconfig.TestConfig) *Common {
	var c *Common
	chainDetails := chainconfig.DevnetConfig()

	duration, err := time.ParseDuration(*testConfig.OCR2.TestDuration)
	if err != nil {
		panic("Invalid test duration")
	}

	if *testConfig.Common.Network == "testnet" {
		chainDetails = chainconfig.SepoliaConfig()
		chainDetails.L2RPCInternal = *testConfig.Common.L2RPCUrl
		if testConfig.Common.L2RPCApiKey == nil {
			chainDetails.L2RPCInternalAPIKey = ""
		} else {
			chainDetails.L2RPCInternalAPIKey = *testConfig.Common.L2RPCApiKey
		}
	} else {
		// set up mocked local feedernet server because starknet-devnet does not provide one
		localDevnetFeederSrv := starknet.NewTestFeederServer()
		chainDetails.FeederURL = localDevnetFeederSrv.URL
	}

	c = &Common{
		TestConfig:   testConfig,
		ChainDetails: chainDetails,
		TestEnvDetails: &TestEnvDetails{
			TestDuration: duration,
		},
		RPCDetails: &RPCDetails{
			P2PPort:             "6690",
			RPCL2Internal:       chainDetails.L2RPCInternal,
			RPCL2InternalAPIKey: chainDetails.L2RPCInternalAPIKey,
		},
	}
	// provide getters for TestConfig (pointers to chain + rpc details)
	c.TestConfig.GetChainID = func() string { return c.ChainDetails.ChainID }
	c.TestConfig.GetFeederURL = func() string { return c.ChainDetails.FeederURL }
	c.TestConfig.GetRPCL2Internal = func() string { return c.RPCDetails.RPCL2Internal }
	c.TestConfig.GetRPCL2InternalAPIKey = func() string { return c.RPCDetails.RPCL2InternalAPIKey }

	return c
}

func (c *Common) Default(t *testing.T, namespacePrefix string) (*Common, error) {
	c.TestEnvDetails.K8Config = &environment.Config{
		NamespacePrefix: fmt.Sprintf("starknet-%s", namespacePrefix),
		TTL:             c.TestEnvDetails.TestDuration,
		Test:            t,
	}

	if *c.TestConfig.Common.InsideK8s {
		tomlString, err := c.TestConfig.GetNodeConfigTOML()
		if err != nil {
			return nil, err
		}
		var overrideFn = func(_ interface{}, target interface{}) {
			ctfconfig.MustConfigOverrideChainlinkVersion(c.TestConfig.ChainlinkImage, target)
		}
		cd := chainlink.NewWithOverride(0, map[string]any{
			"toml":     tomlString,
			"replicas": *c.TestConfig.OCR2.NodeCount,
			"chainlink": map[string]interface{}{
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu":    "2000m",
						"memory": "4Gi",
					},
					"limits": map[string]interface{}{
						"cpu":    "2000m",
						"memory": "4Gi",
					},
				},
			},
			"db": map[string]any{
				"image": map[string]any{
					"version": *c.TestConfig.Common.PostgresVersion,
				},
				"stateful": c.TestConfig.Common.Stateful,
			},
		}, c.TestConfig.ChainlinkImage, overrideFn)
		c.Env = environment.New(c.TestEnvDetails.K8Config).
			AddHelm(devnet.New(nil)).
			AddHelm(mock_adapter.New(nil)).
			AddHelm(cd)
	}

	return c, nil
}

func (c *Common) SetLocalEnvironment(t *testing.T) {
	// Run scripts to set up local test environment
	log.Info().Msg("Starting starknet-devnet container...")
	err := exec.Command("../../scripts/devnet.sh").Run()
	require.NoError(t, err, "Could not start devnet container")
	// TODO: add hardhat too
	log.Info().Msg("Starting postgres container...")
	err = exec.Command("../../scripts/postgres.sh").Run()
	require.NoError(t, err, "Could not start postgres container")
	log.Info().Msg("Starting mock adapter...")
	err = exec.Command("../../scripts/mock-adapter.sh").Run()
	require.NoError(t, err, "Could not start mock adapter")
	log.Info().Msg("Starting core nodes...")
	cmd := exec.Command("../../scripts/core.sh")
	cmd.Env = append(os.Environ(), fmt.Sprintf("CL_CONFIG=%s", c.ChainlinkConfig))
	err = cmd.Run()
	require.NoError(t, err, "Could not start core nodes")
	log.Info().Msg("Set up local stack complete.")

	// Set ChainlinkNodeDetails
	var nodeDetails []*environment.ChainlinkNodeDetail
	var basePort = 50100
	for i := 0; i < *c.TestConfig.OCR2.NodeCount; i++ {
		dbLocalIP := fmt.Sprintf("postgresql://postgres:postgres@chainlink.postgres:5432/starknet_test_%d?sslmode=disable", i+1)
		nodeDetails = append(nodeDetails, &environment.ChainlinkNodeDetail{
			ChartName: "unused",
			PodName:   "unused",
			LocalIP:   "http://127.0.0.1:" + strconv.Itoa(basePort+i),
			// InternalIP: "http://host.container.internal:" + strconv.Itoa(basePort+i), // TODO: chainlink.core.${i}:6688
			InternalIP: fmt.Sprintf("http://chainlink.core.%d:6688", i+1), // TODO: chainlink.core.1:6688
			DBLocalIP:  dbLocalIP,
		})
	}
	c.Env.ChainlinkNodeDetails = nodeDetails
}

func (c *Common) TearDownLocalEnvironment(t *testing.T) {
	log.Info().Msg("Tearing down core nodes...")
	err := exec.Command("../../scripts/core.down.sh").Run()
	require.NoError(t, err, "Could not tear down core nodes")
	log.Info().Msg("Tearing down mock adapter...")
	err = exec.Command("../../scripts/mock-adapter.down.sh").Run()
	require.NoError(t, err, "Could not tear down mock adapter")
	log.Info().Msg("Tearing down postgres container...")
	err = exec.Command("../../scripts/postgres.down.sh").Run()
	require.NoError(t, err, "Could not tear down postgres container")
	log.Info().Msg("Tearing down devnet container...")
	err = exec.Command("../../scripts/devnet.down.sh").Run()
	require.NoError(t, err, "Could not tear down devnet container")
	log.Info().Msg("Tear down local stack complete.")
}

func (c *Common) CreateNodeKeysBundle(nodes []*client.ChainlinkClient) ([]client.NodeKeysBundle, error) {
	nkb := make([]client.NodeKeysBundle, 0)
	for _, n := range nodes {
		p2pkeys, err := n.MustReadP2PKeys()
		if err != nil {
			return nil, err
		}

		peerID := p2pkeys.Data[0].Attributes.PeerID
		txKey, _, err := n.CreateTxKey(c.ChainDetails.ChainName, c.ChainDetails.ChainID)
		if err != nil {
			return nil, err
		}
		ocrKey, _, err := n.CreateOCR2Key(c.ChainDetails.ChainName)
		if err != nil {
			return nil, err
		}

		nkb = append(nkb, client.NodeKeysBundle{
			PeerID:  peerID,
			OCR2Key: *ocrKey,
			TXKey:   *txKey,
		})
	}
	return nkb, nil
}

// CreateJobsForContract Creates and sets up the bootstrap jobs as well as OCR jobs
func (c *Common) CreateJobsForContract(cc *ChainlinkClient, observationSource string, juelsPerFeeCoinSource string, ocrControllerAddress string, accountAddresses []string) error {
	// Define node[0] as bootstrap node
	cc.bootstrapPeers = []client.P2PData{
		{
			InternalIP:   cc.ChainlinkNodes[0].InternalIP(),
			InternalPort: c.RPCDetails.P2PPort,
			PeerID:       cc.NKeys[0].PeerID,
		},
	}

	// Defining relay config
	bootstrapRelayConfig := job.JSONConfig{
		"nodeName":       fmt.Sprintf("starknet-OCRv2-%s-%s", "node", uuid.New().String()),
		"accountAddress": accountAddresses[0],
		"chainID":        c.ChainDetails.ChainID,
	}

	oracleSpec := job.OCR2OracleSpec{
		ContractID:                  ocrControllerAddress,
		Relay:                       c.ChainDetails.ChainName,
		RelayConfig:                 bootstrapRelayConfig,
		ContractConfigConfirmations: 1, // don't wait for confirmation on devnet
	}
	// Setting up bootstrap node
	jobSpec := &client.OCR2TaskJobSpec{
		Name:           fmt.Sprintf("starknet-OCRv2-%s-%s", "bootstrap", uuid.New().String()),
		JobType:        "bootstrap",
		OCR2OracleSpec: oracleSpec,
	}
	_, _, err := cc.ChainlinkNodes[0].CreateJob(jobSpec)
	if err != nil {
		return err
	}

	var p2pBootstrappers []string

	for i := range cc.bootstrapPeers {
		p2pBootstrappers = append(p2pBootstrappers, cc.bootstrapPeers[i].P2PV2Bootstrapper())
	}

	sourceValueBridge := &client.BridgeTypeAttributes{
		Name: "mockserver-bridge",
		URL:  c.RPCDetails.MockServerEndpoint + "/" + strings.TrimPrefix(c.RPCDetails.MockServerURL, "/"),
	}

	// Setting up job specs
	for nIdx, n := range cc.ChainlinkNodes {
		if nIdx == 0 {
			continue
		}
		err := n.MustCreateBridge(sourceValueBridge)
		if err != nil {
			return err
		}
		relayConfig := job.JSONConfig{
			"nodeName":       bootstrapRelayConfig["nodeName"],
			"accountAddress": accountAddresses[nIdx],
			"chainID":        bootstrapRelayConfig["chainID"],
		}

		oracleSpec = job.OCR2OracleSpec{
			ContractID:                  ocrControllerAddress,
			Relay:                       c.ChainDetails.ChainName,
			RelayConfig:                 relayConfig,
			PluginType:                  "median",
			OCRKeyBundleID:              null.StringFrom(cc.NKeys[nIdx].OCR2Key.Data.ID),
			TransmitterID:               null.StringFrom(cc.NKeys[nIdx].TXKey.Data.ID),
			P2PV2Bootstrappers:          pq.StringArray{strings.Join(p2pBootstrappers, ",")},
			ContractConfigConfirmations: 1, // don't wait for confirmation on devnet
			PluginConfig: job.JSONConfig{
				"juelsPerFeeCoinSource": juelsPerFeeCoinSource,
			},
		}

		jobSpec = &client.OCR2TaskJobSpec{
			Name:              fmt.Sprintf("starknet-OCRv2-%d-%s", nIdx, uuid.New().String()),
			JobType:           "offchainreporting2",
			OCR2OracleSpec:    oracleSpec,
			ObservationSource: observationSource,
		}
		_, err = n.MustCreateJob(jobSpec)
		if err != nil {
			return err
		}
	}
	return nil
}
