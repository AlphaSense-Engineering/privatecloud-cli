// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/util"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// errKubectlNotAvailable is the error that is returned when kubectl is not available in PATH.
var errKubectlNotAvailable = errors.New("kubectl is not available in PATH")

const (
	// minArgsCount is the minimum number of arguments the command expects.
	minArgsCount = 4

	// maxArgsCount is the maximum number of arguments the command expects.
	maxArgsCount = 5
)

// logMsgSleeping is the message that is logged when sleeping for a given amount of time.
const logMsgSleeping = "sleeping for %s"

const (
	// flagForce is the name of the flag for the force flag.
	flagForce = "force"

	// flagForceShort is the short name of the flag for the force flag.
	flagForceShort = "f"
)

// kubectlBin is the binary name for kubectl.
const kubectlBin = "kubectl"

var (
	// constPhasesToWaitForWithCrossplane is the list of phases to wait for to proceed to the second step of the installation.
	//
	// Do not modify this variable, it is supposed to be constant.
	constPhasesToWaitForWithCrossplane = append(constPhasesToWaitFor, "Crossplane")

	// constPhasesToWaitFor is the list of phases to wait for to proceed to the third step of the installation.
	//
	// Do not modify this variable, it is supposed to be constant.
	constPhasesToWaitFor = append([]string{"Deploying", "ConfiguringSolr", "Bootstrap"}, constPhasesToWaitForCompleted...)

	// constPhasesToWaitForCompleted is the list of phases to wait for to consider the installation completed.
	//
	// Do not modify this variable, it is supposed to be constant.
	constPhasesToWaitForCompleted = []string{"Ready"}
)

// installCmd is the command to install Private Cloud Kubernetes resources from the YAML files.
type installCmd struct {
	// logger is the logger.
	logger *log.Logger
	// cobraCmd is the Cobra command.
	cobraCmd *cobra.Command
	// checkCmd is the Check command.
	checkCmd *checkCmd
}

var _ cmd = &installCmd{}

// run is the run function for the Install command.
//
// nolint:funlen
func (c *installCmd) run(cobraCmd *cobra.Command, args []string) {
	const (
		// logMsgInstallationStarted is the message that is logged when the installation is started.
		logMsgInstallationStarted = "installation started"

		// logMsgKubectlChecked is the message that is logged when kubectl is checked.
		logMsgKubectlChecked = "kubectl checked"

		// logMsgInstallationCompleted is the message that is logged when the installation is completed.
		logMsgInstallationCompleted = "installation completed"
	)

	c.logger.Info(logMsgInstallationStarted)

	context := args[0]

	var secretsFile *string

	var firstStepFile, secondStepFile, thirdStepFile string

	firstStepFileIndex := 1

	if len(args) == maxArgsCount {
		secretsFile = &args[1]

		firstStepFileIndex = 2
	}

	firstStepFile = args[firstStepFileIndex]

	secondStepFile = args[firstStepFileIndex+1]

	thirdStepFile = args[firstStepFileIndex+2]

	if !util.FlagBool(cobraCmd, flagForce) {
		c.checkCmd.run(cobraCmd, []string{firstStepFile})
	}

	if _, err := exec.LookPath(kubectlBin); err != nil {
		c.logger.Fatal(errKubectlNotAvailable)
	}

	c.logger.Debug(logMsgKubectlChecked)

	if err := util.Exec(c.logger, nil, kubectlBin, "config", "use-context", context); err != nil {
		c.logger.Fatal(err)
	}

	const (
		// countOnce is a constant that is used to apply a file once.
		countOnce = 1

		// countTwice is a constant that is used to apply a file twice.
		countTwice = 2
	)

	if secretsFile != nil {
		if err := c.applyFile(*secretsFile, countOnce); err != nil {
			c.logger.Fatal(err)
		}
	}

	if err := c.applyFile(firstStepFile, countTwice); err != nil {
		c.logger.Fatal(err)
	}

	c.waitForPhases(constPhasesToWaitForWithCrossplane)

	if err := c.applyFile(secondStepFile, countOnce); err != nil {
		c.logger.Fatal(err)
	}

	c.waitForPhases(constPhasesToWaitFor)

	if err := c.applyFile(thirdStepFile, countOnce); err != nil {
		c.logger.Fatal(err)
	}

	c.waitForPhases(constPhasesToWaitForCompleted)

	c.logger.Info(logMsgInstallationCompleted)
}

// applyFile is the function that applies the file.
func (c *installCmd) applyFile(file string, count int) error {
	const (
		// errExitStatusOne is the error that is returned when the exit status is 1.
		errExitStatusOne = "exit status 1"

		// logMsgApplyingFile is the message that is logged when applying the file.
		logMsgApplyingFile = "applying file %s..."

		// logMsgExpectedErrorOccurred is the message that is logged when an expected error occurs and the file is applied again.
		logMsgExpectedErrorOccurred = "expected error occurred, applying again"

		// logMsgFileApplied is the message that is logged when the file is applied.
		logMsgFileApplied = "file %s applied"

		// sleepInterval is the interval of time to sleep between each apply.
		sleepInterval = 1 * time.Minute
	)

	c.logger.Infof(logMsgApplyingFile, file)

	for i := 0; i < count; i++ {
		if err := util.Exec(c.logger, nil, kubectlBin, "apply", "--server-side", "--force-conflicts", "-f", file); err != nil {
			// If the resource mapping is not found on the first apply and the requested apply count is greater than 1,
			// then we can safely ignore the error and proceed to the next apply.
			if count > 1 && i == 0 && strings.Contains(err.Error(), errExitStatusOne) {
				c.logger.Debug(logMsgExpectedErrorOccurred)
			} else {
				return err
			}
		}

		c.logger.Infof(logMsgSleeping, sleepInterval)

		time.Sleep(sleepInterval)
	}

	c.logger.Infof(logMsgFileApplied, file)

	return nil
}

// waitForPhases is the function that waits for the phase of the EnvConfig to be one of the phases in the list.
func (c *installCmd) waitForPhases(phases []string) {
	const (
		// logMsgWaitingForPhases is the message that is logged when waiting for the EnvConfig to be in any of the following phases.
		logMsgWaitingForPhases = "waiting for EnvConfig to be in any of the following phases: %s (current phase: %s)"

		// logMsgCouldNotFindEnvConfig is the message that is logged when the EnvConfig is not found.
		logMsgCouldNotFindEnvConfig = "could not find EnvConfig"

		// logMsgGotPhase is the message that is logged when the correct phase is obtained.
		logMsgGotPhase = "got phase %s, proceeding"
	)

	// sleepInterval is the interval of time to sleep between each check.
	const sleepInterval = 30 * time.Second

	for {
		var outBuf bytes.Buffer

		if err := util.Exec(c.logger, &outBuf, kubectlBin, "get", "envconfig", "-o", "json"); err != nil {
			c.logger.Fatal(err)
		}

		// outputData is the structure of the output of the `kubectl get envconfig -o json` command.
		type outputData struct {
			// Items is the list of items.
			Items []struct {
				// Status is the status of the item.
				Status struct {
					// Phase is the phase of the item.
					Phase string `json:"phase"`
				} `json:"status"`
			} `json:"items"`
		}

		var data outputData

		if err := json.Unmarshal(outBuf.Bytes(), &data); err != nil {
			c.logger.Fatal(err)
		}

		if len(data.Items) == 0 {
			c.logger.Fatal(logMsgCouldNotFindEnvConfig)
		}

		phase := data.Items[0].Status.Phase

		c.logger.Debugf(logMsgWaitingForPhases, phases, phase)

		if slices.Contains(phases, phase) {
			c.logger.Debugf(logMsgGotPhase, phase)

			break
		}

		c.logger.Debugf(logMsgSleeping, sleepInterval)

		time.Sleep(sleepInterval)
	}
}

// newInstallCmd is the constructor for the installCmd.
func newInstallCmd(logger *log.Logger, cobraCmd *cobra.Command) *installCmd {
	return &installCmd{
		logger:   logger,
		cobraCmd: cobraCmd,
	}
}

// Install returns a Cobra command to install Private Cloud Kubernetes resources from the YAML files.
func Install(logger *log.Logger) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "install <context> [<secrets_file>] <first_step_file> <second_step_file> <third_step_file>",
		Short: "Install Private Cloud",
		Args:  cobra.RangeArgs(minArgsCount, maxArgsCount),
	}

	cmd := newInstallCmd(logger, cobraCmd)

	cmd.checkCmd = newCheckCmd(logger, cobraCmd)

	cobraCmd.Long = cmd.checkCmd.longMsg("Install installs Private Cloud Kubernetes resources from the specified YAML files.")

	cobraCmd.Run = cmd.run

	cobraCmd.Flags().BoolP(flagForce, flagForceShort, false, "force the installation")

	cmd.checkCmd.flags(false)

	return cobraCmd
}
