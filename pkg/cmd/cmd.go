package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/DevSecOpsDocs/nuclearpond/pkg/core"
	"github.com/DevSecOpsDocs/nuclearpond/pkg/helpers"
	"github.com/DevSecOpsDocs/nuclearpond/pkg/server"

	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
)

var asciiBanner = figure.NewFigure("Nuclear Pond", "", true).String()

var rootCmd = &cobra.Command{
	Use:     "nuclearpond",
	Short:   "A CLI tool for Nuclear Pond to run nuclei in parallel",
	Long:    "Nuclear Pond invokes nuclei in parallel through invoking lambda functions, customizes command line flags, specifies output, and batches requests.",
	Example: `nuclearpond run -t devsecopsdocs.com -a $(echo -ne "-t dns" | base64) -o cmd`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(asciiBanner)
		cmd.Help()
	},
}

var silent bool
var target string
var targets string
var nucleiArgs string
var region string
var functionName string
var batchSize int
var output string
var threads int

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute nuclei tasks",
	Long:  "Executes nuclei tasks in parallel by invoking lambda asynchronously",
	Run: func(cmd *cobra.Command, args []string) {
		if !silent {
			fmt.Println(asciiBanner)
			fmt.Println("                                                                  devsecopsdocs.com")
			fmt.Println()
		}

		if nucleiArgs == "" {
			log.Fatal("Nuclei arguments are required")
			os.Exit(1)
		}

		if targets == "" && target == "" {
			log.Fatal("Either a target or a list of targets is required")
			os.Exit(1)
		}

		if targets != "" {
			urls := helpers.ReadUrlsFromFile(targets)
			urls = helpers.RemoveEmpty(urls)
			log.Println("Running nuclear pond against", len(urls), "targets")
			batches := helpers.SplitSlice(urls, batchSize)
			log.Println("Splitting targets into", len(batches), "individual executions")
			log.Println("Running with " + fmt.Sprint(threads) + " threads")
			core.ExecuteScans(batches, output, functionName, nucleiArgs, threads, silent, region)
		} else {
			log.Println("Running nuclei against the target", target)
			log.Println("Running with " + fmt.Sprint(threads) + " threads")
			batches := [][]string{{target}}
			core.ExecuteScans(batches, output, functionName, nucleiArgs, threads, silent, region)
		}
	},
}

var startServer = &cobra.Command{
	Use:   "service",
	Short: "Launch API to launch run tasks to the nuclei runner.",
	Long:  "Executes nuclei through an API through asynchronous lambda functions",
	Run: func(cmd *cobra.Command, args []string) {
		// Print banner
		fmt.Println(asciiBanner)
		fmt.Println("                                                                  devsecopsdocs.com")
		fmt.Println()
		// Start server
		log.Println("Running nuclear pond http server on port 8080")
		log.Println("http://localhost:8080")
		server.HandleRequests()
	},
}

func init() {
	// run subcommand
	// Mark flags as required
	runCmd.MarkFlagRequired("args")
	runCmd.MarkFlagRequired("output")
	// General flags
	runCmd.Flags().BoolVarP(&silent, "silent", "s", false, "silent command line output")
	runCmd.Flags().StringVarP(&target, "target", "t", "", "individual target to specify")
	runCmd.Flags().StringVarP(&targets, "targets", "l", "", "list of targets in a file")
	runCmd.Flags().StringVarP(&nucleiArgs, "args", "a", "", "nuclei arguments as base64 encoded string")
	runCmd.Flags().IntVarP(&batchSize, "batch-size", "b", 1, "batch size for number of targets per execution")
	runCmd.Flags().StringVarP(&output, "output", "o", "cmd", "output type to save nuclei results(s3, cmd, or json)")
	runCmd.Flags().IntVarP(&threads, "threads", "c", 1, "number of threads to run lambda functions, default is 1 which will be slow")
	// Region flag
	runCmd.Flags().StringVarP(&region, "region", "r", "", "AWS region to run nuclei")
	if region == "" {
		var ok bool // Declare ok here to avoid shadowing
		region, ok = os.LookupEnv("AWS_REGION") // Removed := to modify the existing region variable
		if !ok {
			runCmd.MarkFlagRequired("region")
		} else {
			runCmd.Flags().Set("region", region)
		}
	}

	// Function name flag
	runCmd.Flags().StringVarP(&functionName, "function-name", "f", "", "AWS Lambda function name")
	if functionName == "" {
		functionName, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME")
		if !ok {
			runCmd.MarkFlagRequired("function-name")
		} else {
			runCmd.Flags().Set("function-name", functionName)
		}
	}
}

// Execute executes the root command.
func Execute() error {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	rootCmd.HasHelpSubCommands()
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(startServer)

	return rootCmd.Execute()
}
