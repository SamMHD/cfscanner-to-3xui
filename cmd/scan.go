/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/Ptechgithub/CloudflareScanner/task"
	"github.com/Ptechgithub/CloudflareScanner/utils"
	"github.com/spf13/cobra"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Test the latency and speed of all IP addresses of Cloudflare CDN, and get the fastest IP",
	Long: `CloudflareScanner
Test the latency and speed of all IP addresses of Cloudflare CDN, and get the fastest IP (IPv4+IPv6)!
https://github.com/Ptechgithub/CloudflareScanner

Options:
    -n 200
        Latency test threads; more threads lead to faster latency testing, do not set too high for low-performance devices (e.g., routers); (default 200, maximum 1000)
    -t 4
        Latency test times; number of times to test latency for a single IP; (default 4 times)
    -dn 10
        Download test count; after latency testing and sorting, number of IPs to test download speed from lowest latency; (default 10)
    -dt 10
        Download test time; maximum time for download speed test of a single IP, should not be too short; (default 10 seconds)
    -tp 443
        Specify test port; port used for latency test/download test; (default port 443)
    -url https://speed.cloudflare.com/__down?bytes=52428800
        Specify test address; address used for latency test (HTTPing)/download test, default address is not guaranteed to be available, it is recommended to self-host;

    -httping
        Switch test mode; switch latency test mode to HTTP protocol, test address used is from [-url] parameter; (default TCPing)
    -httping-code 200
        Valid status code; valid HTTP status code returned during HTTPing latency test, only one is allowed; (default 200 301 302)
    -cfcolo HKG,KHH,NRT,LAX,SEA,SJC,FRA,MAD
        Match specified region; region name is local airport code, separated by English comma, only available in HTTPing mode; (default all regions)

    -tl 200
        Maximum average latency; only output IPs with latency lower than specified maximum average latency, various upper and lower limit conditions can be combined; (default 9999 ms)
    -tll 40
        Minimum average latency; only output IPs with latency higher than specified minimum average latency; (default 0 ms)
    -tlr 0.2
        Maximum loss rate; only output IPs with loss rate lower than/equal to specified loss rate, range 0.00~1.00, 0 filters out any loss IPs; (default 1.00)
    -sl 5
        Minimum download speed; only output IPs with download speed higher than specified download speed, stop testing when enough IPs are gathered [-dn]; (default 0.00 MB/s)

    -p 10
        Display result count; directly display specified number of results after testing, when 0, results are not displayed and program exits; (default 10)
    -f ip.txt
        IP range data file; if path contains spaces, please enclose in quotes; supports other CDN IP ranges; (default ip.txt)
    -ip 1.1.1.1,2.2.2.2/24,2606:4700::/32
        Specify IP range data; specify IP range data to be tested directly through parameters, separated by English comma; (default none)
    -o result.csv
        Write result file; if path contains spaces, please enclose in quotes; leave empty to not write to file [-o ""]; (default result.csv)

    -dd
        Disable download test; after disabling, test results are sorted by latency (default sorted by download speed); (default enabled)
    -allip
        Test all IPs; test each IP in IP range (IPv4 only) (default randomly test one IP in each /24 range)

    -v
        Print program version + check for updates
    -h
        Print help instructions`,
	Run: func(cmd *cobra.Command, args []string) {
		task.InitRandSeed() // Set random seed

		// Start latency testing + filter delay/loss
		pingData := task.NewPing().Run().FilterDelay().FilterLossRate()
		// Start download speed testing
		speedData := task.TestDownloadSpeed(pingData)
		utils.ExportCsv(speedData) // Export to file
		speedData.Print()          // Print results
		endPrint()
	},
}

func init() {

	var minDelay, maxDelay, downloadTime int
	var maxLossRate float64
	flag.IntVar(&task.Routines, "n", 200, "Latency test threads")
	flag.IntVar(&task.PingTimes, "t", 4, "Latency test times")
	flag.IntVar(&task.TestCount, "dn", 10, "Download test count")
	flag.IntVar(&downloadTime, "dt", 10, "Download test time")
	flag.IntVar(&task.TCPPort, "tp", 443, "Specify test port")
	flag.StringVar(&task.URL, "url", "https://speed.cloudflare.com/__down?bytes=52428800", "Specify test address")

	flag.BoolVar(&task.Httping, "httping", false, "Switch test mode")
	flag.IntVar(&task.HttpingStatusCode, "httping-code", 0, "Valid status code")
	flag.StringVar(&task.HttpingCFColo, "cfcolo", "", "Match specified region")

	flag.IntVar(&maxDelay, "tl", 9999, "Maximum average latency")
	flag.IntVar(&minDelay, "tll", 0, "Minimum average latency")
	flag.Float64Var(&maxLossRate, "tlr", 1, "Maximum loss rate")
	flag.Float64Var(&task.MinSpeed, "sl", 0, "Minimum download speed")

	flag.IntVar(&utils.PrintNum, "p", 10, "Display result count")
	flag.StringVar(&task.IPFile, "f", "ip.txt", "IP range data file")
	flag.StringVar(&task.IPText, "ip", "", "Specify IP range data")
	flag.StringVar(&utils.Output, "o", "ip-scan-result.csv", "Output result file")

	flag.BoolVar(&task.Disable, "dd", false, "Disable download test")
	flag.BoolVar(&task.TestAll, "allip", false, "Test all IPs")

	flag.Parse()

	if task.MinSpeed > 0 && time.Duration(maxDelay)*time.Millisecond == utils.InputMaxDelay {
		fmt.Println("[Tip] When using [-sl] parameter, it is recommended to use [-tl] parameter to avoid continuous testing due to insufficient number of [-dn]...")
	}
	utils.InputMaxDelay = time.Duration(maxDelay) * time.Millisecond
	utils.InputMinDelay = time.Duration(minDelay) * time.Millisecond
	utils.InputMaxLossRate = float32(maxLossRate)
	task.Timeout = time.Duration(downloadTime) * time.Second
	task.HttpingCFColomap = task.MapColoMap()

	rootCmd.AddCommand(scanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func endPrint() {
	if utils.NoPrintResult() {
		return
	}
	if runtime.GOOS == "windows" { // If Windows, need to press Enter or Ctrl+C to exit (avoids closing after completion when run by double-clicking)
		fmt.Printf("Press Enter or Ctrl+C to exit.")
		fmt.Scanln()
	}
}
