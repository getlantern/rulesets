# rulesets
This repository contains sing-box rule sets for geo-unblocking, traffic route and ad blocking.

## Creating a new rule set file

Create a pull request that adds a CSV file to the `csv` directory (or edit one of the CSV files). The file should be named `{country}-{type}.csv`, where:
- `{country}` is the [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2) country code (e.g., `us` for the United States, `jp` for Japan). You can also set as `global` if this rule should be applied to all countries.
- `{type}` could be anything that categorizes the rule set, such as `geo`, `adblock`, `ai`, `crypto`, etc.

Each CSV file should contain a list of rules (for example, domains, domain suffixes, domain keywords, packages, process names, or process paths), one per line, with a single header row and no additional columns. The header is currently ignored by the tooling, but for consistency all new CSV files should use `rule_type,value` as the header. For example:
```csv
rule_type,value
domain,lantern.io
domain_suffix,example.org
domain_keyword,google
package_name,com.example.app
process_name,example.exe
process_path,C:\Program Files\Example\example.exe
process_path_regex,^C:\\\\Program Files\\\\Example\\\\.*
ip_cidr,192.0.2.0/24
```

## Generating sing-box rule set files

After adding changes, committing and creating a pull request, GitHub Actions will automatically generate the corresponding sing-box rule set files and place them in the `srs` directory. But you can also generate them locally by following these steps:

1. Make sure you have [Go](https://golang.org/dl/) installed on your machine.
2. Clone this repository to your local machine.
3. Navigate to the repository directory in your terminal.
4. Run the following command to generate the rule set files:

```bash
go run ./cmd/csv_to_srs/main.go -input_dir ./csv -output_dir ./srs
```

This command will read all CSV files from the `csv` directory and generate the corresponding sing-box rule set files in the `srs` directory.
