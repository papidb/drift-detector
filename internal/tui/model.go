package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/papidb/drift-detector/internal/cloud/aws"
	"github.com/papidb/drift-detector/internal/drift"
	"github.com/papidb/drift-detector/internal/parser"
)

type model struct {
	drifts []drift.Drift
	cursor int
	// ready  bool
	err error
}

func loadConfigs(ctx context.Context, instanceID, awsJSONPath, tfPath string) (parser.ParsedEC2Config, parser.ParsedEC2Config, error) {
	nilParserConfig := parser.ParsedEC2Config{}
	configs, err := parser.ParseTerraformHCLFile(tfPath)
	config := parser.EC2Config{}
	for _, cfg := range configs {
		if cfg.Name == instanceID {
			config = cfg.Data
			break
		}
	}
	if err != nil {
		return parser.ParsedEC2Config{}, parser.ParsedEC2Config{}, fmt.Errorf("failed to parse tf config: %w", err)
	}

	// Load AWS config from file or fetch
	var awsCfg *parser.EC2Config
	var name string

	if awsJSONPath != "" {
		awsFileReader, fileCloser, err := aws.ReaderFromFilePath(awsJSONPath)
		if err != nil {
			return nilParserConfig, nilParserConfig, fmt.Errorf("failed to parse aws json: %w", err)
		}

		awsCfg, err = aws.FetchEC2Instance(ctx, nil, awsFileReader, instanceID)
		if err != nil {
			return parser.ParsedEC2Config{}, parser.ParsedEC2Config{}, fmt.Errorf("failed to fetch aws instance: %w", err)
		}
		name = instanceID
		defer fileCloser()
	} else {
		awsConfig, err := aws.NewEC2Client(ctx)

		if err != nil {
			return nilParserConfig, nilParserConfig, fmt.Errorf("failed to create aws client: %w", err)
		}
		awsCfg, err = aws.FetchEC2Instance(ctx, awsConfig, nil, instanceID)
		if err != nil {
			return parser.ParsedEC2Config{}, parser.ParsedEC2Config{}, fmt.Errorf("failed to fetch aws instance: %w", err)
		}
		name = instanceID
	}

	return parser.ParsedEC2Config{Name: name, Data: config},
		parser.ParsedEC2Config{Name: name, Data: *awsCfg},
		nil
}

func Run(ctx context.Context, instanceID, awsJSONPath, tfPath string) error {
	oldCfg, newCfg, err := loadConfigs(ctx, instanceID, awsJSONPath, tfPath)
	if err != nil {
		return err
	}

	drifts, err := drift.CompareEC2Configs(oldCfg, newCfg)
	if err != nil {
		return err
	}

	p := tea.NewProgram(model{
		drifts: drifts,
	})

	_, err = p.Run()
	return err
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.drifts)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}
	if len(m.drifts) == 0 {
		return "✅ No drift detected!\n\nPress q to quit."
	}

	s := "🔍 Drift Summary:\n\n"
	for i, d := range m.drifts {
		cursor := "  "
		if i == m.cursor {
			cursor = "👉"
		}
		s += fmt.Sprintf("%s %s: %v → %v\n", cursor, d.Name, d.OldValue, d.NewValue)
	}
	s += "\nPress ↑/↓ to navigate, q to quit."
	return s
}
