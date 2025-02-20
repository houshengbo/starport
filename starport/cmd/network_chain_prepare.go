package starportcmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tendermint/starport/starport/pkg/clispinner"
	"github.com/tendermint/starport/starport/pkg/goenv"
	"github.com/tendermint/starport/starport/services/network"
	"github.com/tendermint/starport/starport/services/network/networkchain"
)

const (
	flagForce = "force"
)

// NewNetworkChainPrepare returns a new command to prepare the chain for launch
func NewNetworkChainPrepare() *cobra.Command {
	c := &cobra.Command{
		Use:   "prepare [launch-id]",
		Short: "Prepare the chain for launch",
		Args:  cobra.ExactArgs(1),
		RunE:  networkChainPrepareHandler,
	}

	c.Flags().BoolP(flagForce, "f", false, "Force the prepare command to run even if the chain is not launched")
	c.Flags().AddFlagSet(flagNetworkFrom())
	c.Flags().AddFlagSet(flagSetKeyringBackend())
	c.Flags().AddFlagSet(flagSetHome())

	return c
}

func networkChainPrepareHandler(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool(flagForce)

	nb, err := newNetworkBuilder(cmd)
	if err != nil {
		return err
	}
	defer nb.Cleanup()

	// parse launch ID
	launchID, err := network.ParseID(args[0])
	if err != nil {
		return err
	}

	n, err := nb.Network()
	if err != nil {
		return err
	}

	// fetch chain information
	chainLaunch, err := n.ChainLaunch(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	if !force && !chainLaunch.LaunchTriggered {
		return fmt.Errorf("chain %d has not launched yet. use --force to prepare anyway", launchID)
	}

	c, err := nb.Chain(networkchain.SourceLaunch(chainLaunch))
	if err != nil {
		return err
	}

	// fetch the information to construct genesis
	genesisInformation, err := n.GenesisInformation(cmd.Context(), launchID)
	if err != nil {
		return err
	}

	if err := c.Prepare(cmd.Context(), genesisInformation); err != nil {
		return err
	}

	chainHome, err := c.Home()
	if err != nil {
		return err
	}
	binaryName, err := c.BinaryName()
	if err != nil {
		return err
	}
	binaryDir := filepath.Dir(filepath.Join(goenv.Bin(), binaryName))

	fmt.Printf("%s Chain is prepared for launch\n", clispinner.OK)
	fmt.Println("\nYou can start your node by running the following command:")
	commandStr := fmt.Sprintf("%s start --home %s", binaryName, chainHome)
	fmt.Printf("\t%s/%s\n", binaryDir, infoColor(commandStr))

	return nil
}
