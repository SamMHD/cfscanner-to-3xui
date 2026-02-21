package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate outbounds from IP scan result and config templates",
	RunE:  runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	ips, err := readIPsFromCSV("ip-scan-result.csv")
	if err != nil {
		return err
	}
	configs, err := readConfigs("configs")
	if err != nil {
		return err
	}
	var outbounds []map[string]interface{}
	for _, ip := range ips {
		for _, cfg := range configs {
			ob, err := cloneAndSetAddress(cfg, ip)
			if err != nil {
				return err
			}
			outbounds = append(outbounds, ob)
		}
	}
	f, err := os.Create("generated-outbounds.json")
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(outbounds); err != nil {
		return err
	}
	fmt.Println("[generate] completed")
	return nil
}

func readIPsFromCSV(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	var ips []string
	for i, row := range rows {
		if i == 0 && len(row) > 0 && row[0] == "IP Address" {
			continue
		}
		if len(row) > 0 && row[0] != "" {
			ips = append(ips, row[0])
		}
	}
	return ips, nil
}

func readConfigs(dir string) ([]map[string]interface{}, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var configs []map[string]interface{}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var cfg map[string]interface{}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func outboundPrefix() string {
	p := strings.TrimSpace(os.Getenv("OUTBOUND_PREFIX"))
	if p == "" {
		return "cf-clean-"
	}
	return p
}

func cloneAndSetAddress(cfg map[string]interface{}, ip string) (map[string]interface{}, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	prefix := outboundPrefix()
	protocol, _ := out["protocol"].(string)
	switch protocol {
	case "trojan":
		if settings, ok := out["settings"].(map[string]interface{}); ok {
			if servers, ok := settings["servers"].([]interface{}); ok && len(servers) > 0 {
				if s, ok := servers[0].(map[string]interface{}); ok {
					s["address"] = ip
				}
			}
		}
		out["tag"] = prefix + "trojan-" + ip
	case "vless":
		if settings, ok := out["settings"].(map[string]interface{}); ok {
			if vnext, ok := settings["vnext"].([]interface{}); ok {
				for _, v := range vnext {
					if m, ok := v.(map[string]interface{}); ok {
						m["address"] = ip
					}
				}
			}
		}
		out["tag"] = prefix + "vless-" + ip
	}
	return out, nil
}
