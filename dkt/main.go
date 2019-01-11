// Command dkt runs docker-compose with a set of docket files.
//
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"

	"github.com/bloomberg/docket/internal/compose"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func executeSubcommand(cobraCmd *cobra.Command, cobraCmdArgs []string) {
	ctx := context.Background()

	prefix := viper.GetString("prefix")
	if prefix == "" {
		prefix = "docket"
	}

	mode := viper.GetString("mode")
	if mode == "" {
		panic(fmt.Errorf("error: use -m|--mode or set DOCKET_MODE"))
	}

	cmp, cleanup, err := compose.NewCompose(ctx, prefix, mode)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			panic(err)
		}
	}()

	args := []string{cobraCmd.CalledAs()}
	args = append(args, cobraCmdArgs...)

	cmd := cmp.Command(ctx, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cobraCmd.Printf("Running: %v\n", cmd.Args)

	signal.Ignore(os.Interrupt)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

type subcmd struct {
	cmd  string
	help string
}

// Run docker-compose --help to find its subcommands so we can use them as our own.
// Using subcommands lets us avoid requiring '--' to separate dkt flags from docker-compose flags.
func getDockerComposeCommands() ([]subcmd, error) {
	cmd := exec.CommandContext(context.Background(), "docker-compose", "--help") // #nosec
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	subs := []subcmd{}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	seenHeader := false
	pattern := regexp.MustCompile(`^  ([^ ]+) +(.+)$`)
	for scanner.Scan() {
		if scanner.Text() == "Commands:" {
			seenHeader = true
		} else if seenHeader {
			mm := pattern.FindStringSubmatch(scanner.Text())
			if len(mm) != 3 {
				return nil, fmt.Errorf("couldn't match line %q", scanner.Text())
			}
			if mm[1] == "help" {
				continue
			}
			subs = append(subs, subcmd{cmd: mm[1], help: mm[2]})
		}
	}

	return subs, nil
}

func main() {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		TraverseChildren: true,
		Use:              "dkt",
		Short:            "root command",
		Long: `
dkt runs docker-compose with a set of docket files matching the mode and
optional prefix.

Any arguments that aren't dkt-specific will be passed through to docker-compose.`[1:],
	}

	subs, err := getDockerComposeCommands()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, s := range subs {
		cmd := &cobra.Command{
			DisableFlagParsing: true,
			Use:                s.cmd,
			Short:              s.help,
			Run:                executeSubcommand,
		}
		rootCmd.AddCommand(cmd)
	}

	cobra.OnInitialize(func() {
		viper.SetEnvPrefix("DOCKET")
		if err := viper.BindEnv("mode"); err != nil {
			panic(err)
		}
		if err := viper.BindEnv("prefix"); err != nil {
			panic(err)
		}
	})

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringP("prefix", "p", "", "prefix (or set DOCKET_PREFIX)")
	if err := viper.BindPFlag("prefix", rootCmd.PersistentFlags().Lookup("prefix")); err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringP("mode", "m", "", "mode (or set DOCKET_MODE)")
	if err := viper.BindPFlag("mode", rootCmd.PersistentFlags().Lookup("mode")); err != nil {
		panic(err)
	}

	// The subcommands signal failure by panic'ing.
	// If there's a better way to do this, please let me know.
	// I don't want to call os.Exit() in the subcommand because it bypasses defers.
	defer func() {
		if x := recover(); x != nil {
			fmt.Println(x)
			os.Exit(1)
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
