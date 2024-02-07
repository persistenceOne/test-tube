package testenv

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	// cometbft

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmlog "github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"

	// cosmos-sdk
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// wasmd
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	// persistence
	"github.com/persistenceOne/persistenceCore/v9/app"
)

type TestEnv struct {
	App                *app.Application
	Ctx                sdk.Context
	ValPrivs           []*secp256k1.PrivKey
	NodeHome           string
	ParamTypesRegistry *ParamTypeRegistry
}

func NewTestEnv(chainID, nodeHome string) *TestEnv {
	// Allow testing unoptimized contract
	wasmtypes.MaxWasmSize = 1024 * 1024 * 1024 * 1024 * 1024

	valPriv := secp256k1.GenPrivKey()
	env := new(TestEnv)
	env.ValPrivs = []*secp256k1.PrivKey{valPriv}
	env.NodeHome = nodeHome
	env.App = SetupPersistenceApp(chainID, nodeHome, valPriv)
	env.Ctx = env.App.BaseApp.NewContext(false, tmproto.Header{Height: 0, ChainID: chainID, Time: time.Now().UTC()})
	env.ParamTypesRegistry = NewParamTypeRegistry()

	// Setup validator signing info
	valAddr := sdk.ValAddress(valPriv.PubKey().Address())
	val, found := env.App.StakingKeeper.GetValidator(env.Ctx, valAddr)
	requireTrue("validator found", found)

	consAddr, err := val.GetConsAddr()
	requireNoErr(err)

	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		env.Ctx.BlockHeight(),
		0,
		time.Unix(0, 0),
		false,
		0,
	)
	env.App.SlashingKeeper.SetValidatorSigningInfo(env.Ctx, consAddr, signingInfo)

	return env
}

func SetupPersistenceApp(chainID string, nodeHome string, valPriv *secp256k1.PrivKey) *app.Application {
	setConfig()
	db := dbm.NewMemDB()
	appOpts := simtestutil.AppOptionsMap{
		flags.FlagHome:            nodeHome,
		server.FlagTrace:          true,
		server.FlagInvCheckPeriod: 5,
	}

	logger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout))
	appInstance := app.NewApplication(
		logger,
		db,
		nil,
		true,
		app.GetEnabledProposals(),
		appOpts,
		[]wasm.Option{},
		baseapp.SetChainID(chainID),
	)

	genState := getGenesisState(valPriv)
	stateBytes, err := json.MarshalIndent(genState, "", " ")
	requireNoErr(err)

	concensusParams := simtestutil.DefaultConsensusParams
	concensusParams.Block = &tmproto.BlockParams{
		MaxBytes: 22020096,
		MaxGas:   -1,
	}

	appInstance.InitChain(abci.RequestInitChain{
		ChainId:         chainID,
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: concensusParams,
		AppStateBytes:   stateBytes,
	})

	return appInstance
}

// setCfg params at the package state
func setConfig() {
	sdk.DefaultBondDenom = app.BondDenom
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	cfg.SetCoinType(app.CoinType)
	cfg.SetPurpose(app.Purpose)
	cfg.Seal()
}

func getGenesisState(valPriv *secp256k1.PrivKey) app.GenesisState {
	ba := authtypes.NewBaseAccount(valPriv.PubKey().Address().Bytes(), valPriv.PubKey(), 0, 0)
	genAccounts := []authtypes.GenesisAccount{ba}
	balance := banktypes.Balance{
		Address: ba.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
	}

	tmPub, err := cryptocodec.ToTmPubKeyInterface(valPriv.PubKey())
	requireNoErr(err)

	validator := tmtypes.NewValidator(tmPub, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	encCfg := app.MakeEncodingConfig()
	genState, err := simtestutil.GenesisStateWithValSet(encCfg.Codec, app.NewDefaultGenesisState(), valSet, genAccounts, balance)
	requireNoErr(err)

	return overrideGenesis(encCfg.Codec, genState)
}

func overrideGenesis(cdc codec.Codec, genesisState app.GenesisState) app.GenesisState {
	var wasmGen wasm.GenesisState
	cdc.MustUnmarshalJSON(genesisState[wasm.ModuleName], &wasmGen)
	wasmGen.Params = wasmtypes.Params{
		// Allow store code without gov
		CodeUploadAccess:             wasmtypes.AllowEverybody,
		InstantiateDefaultPermission: wasmtypes.AccessTypeEverybody,
	}
	genesisState[wasm.ModuleName] = cdc.MustMarshalJSON(&wasmGen)

	stakingGen := stakingtypes.GetGenesisStateFromAppState(cdc, genesisState)
	stakingGen.Params.UnbondingTime = time.Hour * 24 * 7 * 2 // 2 weeks
	genesisState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingGen)
	return genesisState
}

func (env *TestEnv) GetValidatorAddresses() []string {
	validators := env.App.StakingKeeper.GetAllValidators(env.Ctx)
	var addresses []string
	for _, validator := range validators {
		addresses = append(addresses, validator.OperatorAddress)
	}

	return addresses
}

// BeginNewBlock begins a new block with a proposer.
func (env *TestEnv) BeginNewBlock(timeIncreaseSeconds uint64) {
	validators := env.App.StakingKeeper.GetAllValidators(env.Ctx)
	if len(validators) == 0 {
		panic("expected at least 1 validator")
	}
	validator := validators[0]
	valConsAddr, err := validator.GetConsAddr()
	requireNoErr(err)

	valAddr := valConsAddr.Bytes()
	newBlockTime := env.Ctx.BlockTime().Add(time.Duration(timeIncreaseSeconds) * time.Second)

	header := tmproto.Header{ChainID: env.Ctx.ChainID(), Height: env.Ctx.BlockHeight() + 1, Time: newBlockTime}
	newCtx := env.Ctx.WithBlockTime(newBlockTime).WithBlockHeight(env.Ctx.BlockHeight() + 1)
	env.Ctx = newCtx
	lastCommitInfo := abci.CommitInfo{
		Votes: []abci.VoteInfo{{
			Validator:       abci.Validator{Address: valAddr, Power: 1000},
			SignedLastBlock: true,
		}},
	}
	reqBeginBlock := abci.RequestBeginBlock{Header: header, LastCommitInfo: lastCommitInfo}

	env.App.BeginBlock(reqBeginBlock)
	env.Ctx = env.App.NewContext(false, reqBeginBlock.Header)
}

func requireNoErr(err error) {
	if err != nil {
		panic(err)
	}
}

// func requireNoNil(name string, nilable any) {
// 	if nilable == nil {
// 		panic(fmt.Sprintf("%s must not be nil", name))
// 	}
// }

func requireTrue(name string, b bool) {
	if !b {
		panic(fmt.Sprintf("%s must be true", name))
	}
}
