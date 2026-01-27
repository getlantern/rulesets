package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagernet/sing-box/common/srs"
	"github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
)

func main() {
	inputDir := flag.String("input_dir", "./csv", "Directory containing input CSV files")
	outputDir := flag.String("output_dir", "./srs", "Directory to save output SRS files")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, `
Convert CSV rule files to SRS format.
Each CSV file should have at least two columns and a mandatory header(which is skipped), where the first column determines the kind of rule (domain, domain_suffix, package_name, process_name, etc.) and the second column contains the corresponding value. e.g.:

rule_type,value
domain,example.com
domain_suffix,example.org

`)
		flag.PrintDefaults()
	}

	flag.Parse()

	err := filepath.Walk(*inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".csv") {
			plainRuleSet, err := convertCSVToPlainRuleSet(path)
			if err != nil {
				return err
			}

			outputPath := filepath.Join(*outputDir, strings.TrimSuffix(info.Name(), ".csv")+".srs")
			srsFile, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create SRS file: %w", err)
			}
			defer srsFile.Close()

			if err := srs.Write(srsFile, plainRuleSet, constant.RuleSetVersionCurrent); err != nil {
				return fmt.Errorf("failed to write SRS file %q: %w", outputPath, err)
			}

		}
		return nil
	})
	if err != nil {
		slog.Error("Error processing files", "error", err)
	}
}

func convertCSVToPlainRuleSet(inputPath string) (option.PlainRuleSet, error) {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return option.PlainRuleSet{}, err
	}
	jsonContent, err := csvToJson(content)
	if err != nil {
		return option.PlainRuleSet{}, err
	}
	ruleset, err := json.UnmarshalExtended[option.PlainRuleSet](jsonContent)
	if err != nil {
		return option.PlainRuleSet{}, err
	}
	return ruleset, nil
}

func csvToJson(csv []byte) ([]byte, error) {
	tmpMap := make(map[string][]string)
	lines := strings.Split(string(csv), "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}
		kv := strings.Split(line, ",")
		if len(kv) != 2 {
			slog.Warn("unexpected row", slog.String("line", line))
			continue
		}
		if _, exist := tmpMap[kv[0]]; !exist {
			tmpMap[kv[0]] = make([]string, 0)
		}
		tmpMap[kv[0]] = append(tmpMap[kv[0]], kv[1])
	}
	jsonMap, err := json.Marshal(tmpMap)
	if err != nil {
		return nil, err
	}
	return jsonMap, nil
}
