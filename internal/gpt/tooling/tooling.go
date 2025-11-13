package tooling

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/packages/param"
	"github.com/openai/openai-go/v2/responses"
)

type ToolKind string

const (
	ToolKindMCP ToolKind = "mcp"
)

type Spec struct {
	Kind ToolKind
	MCP  *MCPOptions
}

type MCPOptions struct {
	ServerLabel    string
	ServerURL      param.Opt[string]
	ApprovalPolicy string
}

func (o MCPOptions) validate() error {
	if strings.TrimSpace(o.ServerLabel) == "" {
		return errors.New("mcp server label is required")
	}
	if !o.ServerURL.Valid() {
		return errors.New("mcp server url is required")
	}
	if strings.TrimSpace(o.ApprovalPolicy) == "" {
		return errors.New("mcp approval policy is required")
	}

	return nil
}

func (s Spec) Validate() error {
	switch s.Kind {
	case ToolKindMCP:
		if s.MCP == nil {
			return errors.New("mcp options must be provided")
		}

		return s.MCP.validate()
	default:
		return fmt.Errorf("unsupported tool kind: %s", s.Kind)
	}
}

func (s Spec) ToResponseTool() (responses.ToolUnionParam, error) {
	if err := s.Validate(); err != nil {
		return responses.ToolUnionParam{}, err
	}

	switch s.Kind {
	case ToolKindMCP:
		return responses.ToolUnionParam{
			OfMcp: &responses.ToolMcpParam{
				ServerLabel: s.MCP.ServerLabel,
				ServerURL:   s.MCP.ServerURL,
				RequireApproval: responses.ToolMcpRequireApprovalUnionParam{
					OfMcpToolApprovalSetting: openai.Opt(s.MCP.ApprovalPolicy),
				},
			},
		}, nil
	default:
		return responses.ToolUnionParam{}, fmt.Errorf("unsupported tool kind: %s", s.Kind)
	}
}

type Provider interface {
	Tools(context.Context) ([]responses.ToolUnionParam, error)
}

type StaticProvider struct {
	specs []Spec
}

func NewStaticProvider(specs []Spec) *StaticProvider {
	copied := make([]Spec, len(specs))
	for i, spec := range specs {
		copied[i] = spec
		if spec.MCP != nil {
			opts := *spec.MCP
			copied[i].MCP = &opts
		}
	}

	return &StaticProvider{specs: copied}
}

func (p *StaticProvider) Tools(ctx context.Context) ([]responses.ToolUnionParam, error) {
	if p == nil {
		return nil, nil
	}

	_ = ctx

	tools := make([]responses.ToolUnionParam, 0, len(p.specs))
	for _, spec := range p.specs {
		tool, err := spec.ToResponseTool()
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}

	return tools, nil
}

func DefaultSpecs() []Spec {
	return []Spec{
		{
			Kind: ToolKindMCP,
			MCP: &MCPOptions{
				ServerLabel:    "deepwiki",
				ServerURL:      param.NewOpt("https://mcp.deepwiki.com/mcp"),
				ApprovalPolicy: "never",
			},
		},
	}
}
