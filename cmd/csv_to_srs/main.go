package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagernet/sing-box/common/srs"
	"github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
)

func main() {
	inputDir := flag.String("input_dir", "./csv", "Directory containing input CSV files")
	outputDir := flag.String("output_dir", "./srs", "Directory to save output SRS files")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, `
Convert CSV rule files to SRS format.
Each CSV file should have at least two columns and a mandatory header(which is skipped), where the first column determine the kind of rule (domain, domain_suffix, package_name, process_name, etc.) and the second column contains the corresponding value. e.g.:

rule_type,value
domain,example.com
domain-suffix,example.org

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
	csvFile, err := os.Open(inputPath)
	if err != nil {
		return option.PlainRuleSet{}, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer csvFile.Close()

	ruleset := option.PlainRuleSet{
		Rules: []option.HeadlessRule{
			{
				Type: constant.RuleTypeDefault,
				DefaultOptions: option.DefaultHeadlessRule{
					Domain:       badoption.Listable[string]{},
					DomainSuffix: badoption.Listable[string]{},
					PackageName:  badoption.Listable[string]{},
					ProcessName:  badoption.Listable[string]{},
					IPCIDR:       badoption.Listable[string]{},
				},
			},
		},
	}
	r := csv.NewReader(csvFile)
	i := 0
	for {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return option.PlainRuleSet{}, fmt.Errorf("failed to read record: %w", err)
		}

		if i == 0 {
			i++
			continue
		}

		ruleType := record[0]
		switch ruleType {
		case "domain":
			ruleset.Rules[0].DefaultOptions.Domain = append(ruleset.Rules[0].DefaultOptions.Domain, record[1])
		case "domain_suffix":
			ruleset.Rules[0].DefaultOptions.DomainSuffix = append(ruleset.Rules[0].DefaultOptions.DomainSuffix, record[1])
		case "package_name":
			ruleset.Rules[0].DefaultOptions.PackageName = append(ruleset.Rules[0].DefaultOptions.PackageName, record[1])
		case "process_name":
			ruleset.Rules[0].DefaultOptions.ProcessName = append(ruleset.Rules[0].DefaultOptions.ProcessName, record[1])
		case "ip_cidr":
			ruleset.Rules[0].DefaultOptions.IPCIDR = append(ruleset.Rules[0].DefaultOptions.IPCIDR, record[1])
		default:
			return option.PlainRuleSet{}, fmt.Errorf("unknown rule type: %s", ruleType)
		}
	}
	return ruleset, nil
}
