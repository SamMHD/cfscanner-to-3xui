/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
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

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func envBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func init() {
	var minDelay, maxDelay, downloadTime int
	var maxLossRate float64

	// Latency test threads (default 200, max 1000)
	task.Routines = envInt("SCAN_N", 200)
	// Latency test times per IP
	task.PingTimes = envInt("SCAN_T", 4)
	// Number of IPs to run download test on (lowest latency first)
	task.TestCount = envInt("SCAN_DN", 10)
	// Max seconds per IP download test
	downloadTime = envInt("SCAN_DT", 10)
	// Port for latency/download test
	task.TCPPort = envInt("SCAN_TP", 443)
	// URL for HTTPing/download test
	task.URL = envStr("SCAN_URL", "https://speed.cloudflare.com/__down?bytes=52428800")

	// Use HTTP for latency test instead of TCP
	task.Httping = envBool("SCAN_HTTPING", false)
	// Valid HTTP status code for HTTPing (e.g. 200)
	task.HttpingStatusCode = envInt("SCAN_HTTPING_CODE", 0)
	// Comma-separated airport codes to match region (HTTPing only)
	task.HttpingCFColo = envStr("SCAN_CFCOLO", "")

	// Max average latency (ms); filter out higher
	maxDelay = envInt("SCAN_TL", 9999)
	// Min average latency (ms); filter out lower
	minDelay = envInt("SCAN_TLL", 0)
	// Max loss rate 0–1; filter out higher
	maxLossRate = envFloat("SCAN_TLR", 1)
	// Min download speed (MB/s); filter out lower
	task.MinSpeed = envFloat("SCAN_SL", 0)

	// How many results to print (0 = no print, exit after test)
	utils.PrintNum = envInt("SCAN_P", 10)
	// IP range file path
	task.IPFile = envStr("SCAN_F", "ip.txt")
	// Inline IP ranges (comma-separated)
	task.IPText = envStr("SCAN_IP", "")
	// Output CSV path
	utils.Output = envStr("SCAN_O", "ip-scan-result.csv")

	// Disable download test; sort by latency only
	task.Disable = envBool("SCAN_DD", false)
	// Test every IP in range (IPv4); default one per /24
	task.TestAll = envBool("SCAN_ALLIP", false)

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
